package coordinator

import (
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"errors"
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
	masterStorageServer *rpc.Client            // RPC connection to masterStorage
	servers             []storagerpc.Node      // slice of storage servers
	nextTransactionIds  map[string]int         // [hostport] --> next TransactionID for that StorageServer
	connections         map[string]*rpc.Client // [hostport] --> client conn
	nextCommitId        int
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

	// create conns to be cached and add masterServer to map
	conns := make(map[string]*rpc.Client)
	conns[masterServerHostPort] = cli

	// create and init unique Ids for every node in the ring
	nextIds := make(map[string]int)
	for i := 0; i < len(servers); i++ {
		nextIds[servers[i].HostPort] = 1
	}

	// create the coordinator
	coord := &coordinator{
		masterStorageServer: cli,
		servers:             servers,
		nextTransactionIds:  nextIds,
		connections:         conns,
		nextCommitId:        1,
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

/* PerformTransaction will break apart a transaction and
   send the appropriate updates to each corresponding
   StorageServer using a 2PC propose call */
func (coord *coordinator) PerformTransaction(name storagerpc.TransactionType, data string) (storagerpc.TransactionStatus, error) {
	switch name {
	// TODO: use data to break up route appropriately
	case storagerpc.CreateUser:
	case storagerpc.CreateTeam:
	case storagerpc.JoinTeam:
	case storagerpc.LeaveTeam:
	case storagerpc.MakeTransaction:
	}

	return 0, errors.New("No such method name")
}

/*func (coord *coordinator) Propose(args *storagerpc.ProposeArgs, reply *storagerpc.ProposeReply) error {*/

/* callName: Get/Put/Execute
 * data: JSON-marshaled obj for use with callName */
func (coord *coordinator) Propose(transactionId int, callName, data string) (storagerpc.Status, error) {
	// Prepare for transaction
	hostport := ""
	tArgs := &storagerpc.PrepareArgs{
		// Not correct right now placeholder
		TransactionId: coord.nextTransactionIds[hostport],
		Key:           callName,
		Value:         data,
	}

	coord.nextTransactionIds[hostport]++

	stat := storagerpc.CommitStatus(storagerpc.Commit)
	var err error
	resultStatus := storagerpc.OK

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
		replyStatus := rpcReply.Reply.(*storagerpc.ProposeReply).Status
		if rpcReply.Error != nil || replyStatus != storagerpc.OK {
			resultStatus := replyStatus
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
			return 0, rpcReply.Error
		}
	}
	return storagerpc.Status(resultStatus), nil
}
