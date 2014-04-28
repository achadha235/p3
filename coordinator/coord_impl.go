package coordinator

import (
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	MaxConnectAttempts = 5
)

/*
Implementation Plan:
1. Move everything to single thread (done)
2. Hook in Transact to break up transactions to appropriate servers

*/

type coordinator struct {
	masterStorageServer *rpc.Client       // RPC connection to masterStorage
	servers             []storagerpc.Node // slice of storage servers
	nextServerId        int
	nextTransactionId   int
	/*	client            chan int*/
	/*	self              storagerpc.Node*/
	/*	requestNodeId     chan chan int*/
	/*	proposeChannel    chan proposeRequest*/
}

type proposeRequest struct {
	args  *storagerpc.ProposeArgs
	reply chan error
}

func StartCoordinator(masterServerHostPort string) (Coordinator, error) {
	cli, err := util.TryDial(masterServerHostPort)
	if err != nil {
		return nil, err
	}

	args := &storagerpc.GetServersArgs{}
	var reply storagerpc.GetServersReply

	// attempt to get the list of servers in the ring from the MasterStorageServer
	var servers []storagerpc.Node
	for t := MaxConnectAttempts; ; t-- {
		err := cli.Call("StorageServer.GetServers", args, reply)
		if reply.Status == storagerpc.OK {
			servers := reply.Servers
			break
		} else if t <= 0 {
			// StorageServers not ready
			return nil, err
		}
		// Wait a second before retrying
		time.Sleep(time.Second)
	}

	// create the coordinator
	coord := &coordinator{
		masterStorageServer: cli,
		servers:             servers,
		nextServerId:        1,
		nextTransactionId:   1,
	}

	rpc.RegisterName("Coordinator", coord)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":3000")
	if e != nil {
		log.Fatalln("listen error:", e)
	}
	go http.Serve(l, nil)

	return coord, nil
}

func (coord *coordinator) Propose(args *storagerpc.ProposeArgs, reply *storagerpc.ProposeReply) error {

	// Prepare for transaction
	tArgs := &storagerpc.PrepareArgs{
		TransactionId: coord.nextTransactionId,
		Key:           args.Key,
		Value:         args.Value,
	}

	coord.nextTransactionId++
	stat := storagerpc.CommitStatus(storagerpc.Commit)
	var err error

	// channel to receive async replies from CohortServers
	doneCh := make(chan *rpc.Call, len(coord.servers))

	// send out Prepare call to all nodes
	for i := 0; i < len(coord.servers); i++ {
		var reply *storagerpc.PrepareReply
		coord.servers[i].Client.Go("CohortStorageServer.Prepare", tArgs, &reply, doneCh)
	}

	// receive replies from prepare
	for i := 0; i < len(coord.servers); i++ {
		rpcReply := <-doneCh

		// if RPC fails or non-OK status then Rollback
		if rpcReply.Error != nil || rpcReply.Reply.(*storagerpc.ProposeReply).Status != storagerpc.OK {
			stat = storagerpc.Rollback
		}
	}

	// send the Commit call to all nodes with the updated status
	for _, node := range coord.servers {
		commitArgs := &storagerpc.CommitArgs{
			TransactionId: tArgs.TransactionId,
			Status:        stat,
		}

		var commitReply *storagerpc.CommitReply
		node.Client.Go("CohortStorageServer.Commit", commitArgs, &commitReply, doneCh)
	}

	// receive Ack from all nodes
	for i := 0; i < len(coord.servers); i++ {
		rpcReply := <-doneCh

		// TODO: if RPC fails then retry sending message until all received (?)
		if rpcReply.Error != nil {
			return rpcReply.Error
		}
	}

	reply.Status = storagerpc.OK
	return nil
}

/*
func (coord *coordinator) coordinatorHandler() {
	for {
		select {
		case replyChan := <-coord.requestNodeId:
			client, err := rpc.DialHTTP("tcp", ":3000")
			if err != nil {
				log.Fatalln("dialing:", err)
			}
			newNode := &storagerpc.Node{
				NodeId:   coord.nextServerId,
				HostPort: "localhost:3000",
				Master:   false,
				Client:   client,
			}
			coord.servers[newNode.NodeId] = newNode
			coord.nextServerId++
			replyChan <- coord.nextServerId
		}
	}
}*/
