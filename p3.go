package main

import (
	"github.com/achadha235/p3/rpc/storagerpc"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"runtime"
)

type masterStorageServer struct {
	client            chan int
	self              storagerpc.Node
	servers           map[int]*storagerpc.Node
	nextServerId      int
	nextTransactionId int
	requestNodeId     chan chan int
	proposeChannel    chan proposeRequest
}

type cohortStorageServer struct {
	rpc            *rpc.Client
	prepareChannel chan prepareRequest
	commitChannel  chan commitRequest
	getChannel     chan getRequest
	putChannel     chan putRequest

	storage map[string]string // Key value storage
	undoLog map[int]storagerpc.LogEntry  // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	redoLog map[int]storagerpc.LogEntry  // TransactionId to Store 1. Key 2. TransactionID. (New)Value
}



type proposeRequest struct {
	args  *storagerpc.ProposeArgs
	reply chan error
}

type prepareRequest struct {
	args  *storagerpc.PrepareArgs
	reply chan error
}

type commitRequest struct {
	args  *storagerpc.CommitArgs
	reply chan error
}

type getRequest struct {
	args  *storagerpc.GetArgs
	reply chan storagerpc.GetReply
}

type putRequest struct {
	args  *storagerpc.PutArgs
	reply chan error
}



func main() {
	master := StartMasterServer()
	cohort := StartCohortServer()
	fmt.Println(master)
	fmt.Println(cohort)

	client, err := rpc.DialHTTP("tcp", ":3000")
	if err != nil {
		log.Fatalln("dialing:", err)
	}

	args := &storagerpc.PutArgs{"hello", "world"}
	var reply *storagerpc.PutReply
	err = client.Call("CohortStorageServer.Put", args, &reply)
	if err != nil {
		log.Fatalln("Put error:", err)
	}
	fmt.Println("Reply returned")
	fmt.Println(reply)

	args2 := &storagerpc.GetArgs{"hello"}
	var reply2 *storagerpc.GetReply
	err = client.Call("CohortStorageServer.Get", args2, &reply2)
	if err != nil {
		log.Fatalln("Get error: ", err)
	}

	fmt.Println("Reply val: ", reply2.Value)

}

// 2PC leader handler
func StartMasterServer() storagerpc.MasterStorageServer {
	fmt.Println("Entering StartMasterServer")
	defer fmt.Println("Exit StartMasterServer")

	server := new(masterStorageServer)
	server.nextServerId = 1
	server.nextTransactionId = 1
	server.requestNodeId = make(chan chan int)
	server.proposeChannel = make(chan proposeRequest)
	server.servers = make(map[int]*storagerpc.Node)

	rpc.RegisterName("MasterStorageServer", server)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":3000")
	if e != nil {
		log.Fatalln("listen error:", e)
	}
	go http.Serve(l, nil)
	go server.masterServerHandler()

	return server
}

// 2PC cohort handler
func StartCohortServer() storagerpc.CohortStorageServer {
	fmt.Println("Entering StartCohortServer")
	defer fmt.Println("Exit StartCohortServer")

	client, err := rpc.DialHTTP("tcp", ":3000")
	if err != nil {
		log.Fatalln("dialing:", err)
	}
	args := &storagerpc.RegisterServerArgs{}
	var reply *storagerpc.RegisterServerReply
	err = client.Call("MasterStorageServer.RegisterServer", args, &reply)
	if err != nil {
		log.Fatalln("Register error:", err)
	}
	fmt.Println("RegisterServer returned")
	fmt.Println(reply)

	// Set up cohort rpc
	server := new(cohortStorageServer)

	server.rpc = client
	server.prepareChannel = make(chan prepareRequest)
	server.commitChannel = make(chan commitRequest)
	server.getChannel = make(chan getRequest)
	server.putChannel = make(chan putRequest)

	server.storage = make(map[string]string)
	server.redoLog = make(map[int]storagerpc.LogEntry)
	server.undoLog = make(map[int]storagerpc.LogEntry)

	rpc.RegisterName("CohortStorageServer", server)

	fmt.Println("Cohort server handler started...")
	go server.cohortServerHandler()

	return server
}



func (ss *masterStorageServer) RegisterServer(args *storagerpc.RegisterServerArgs, reply *storagerpc.RegisterServerReply) error {
	fmt.Println("Enter RegisterServer")
	defer fmt.Println("Exit RegisterServer")

	replyChan := make(chan int)
	ss.requestNodeId <- replyChan
	newNodeId := <-replyChan

	reply.Status = storagerpc.OK
	reply.NodeId = newNodeId
	return nil
}
func (ss *masterStorageServer) Propose(args *storagerpc.ProposeArgs, reply *storagerpc.ProposeReply) error {
	fmt.Println("Enter Propose")
	defer fmt.Println("Exit Propose")

	r := make(chan error)
	ss.proposeChannel <- proposeRequest{
		args,
		r,
	}
	error := <-r
	if error != nil {
		reply.Status = storagerpc.OK
	} else {
		reply.Status = storagerpc.OK

	}
	return error
}

func (ss *cohortStorageServer) Prepare(args *storagerpc.PrepareArgs, reply *storagerpc.PrepareReply) error {
	fmt.Println("Enter Prepare")
	defer fmt.Println("Exit Prepare")

	fmt.Println("Number of goroutines")
	fmt.Println(runtime.NumGoroutine())
	fmt.Println("Making channel")
	r := make(chan error)
	ss.prepareChannel <- prepareRequest{
		args,
		r,
	}
	fmt.Println("sent request")

	error := <-r
	fmt.Println("got error")

	if error != nil {
		reply.Status = storagerpc.OK
	} else {
		reply.Status = storagerpc.OK

	}
	return error
}

func (ss *cohortStorageServer) Commit(args *storagerpc.CommitArgs, reply *storagerpc.CommitReply) error {
	fmt.Println("Enter Commit")
	defer fmt.Println("Exit Commit")

	r := make(chan error)
	ss.commitChannel <- commitRequest{
		args,
		r,
	}
	error := <-r
	if error != nil {
	}
	return nil
}

func (ss *cohortStorageServer) Get(args *storagerpc.GetArgs, reply *storagerpc.GetReply) error {
	fmt.Println("Enter Get")
	defer fmt.Println("Exit Get")

	r := make(chan storagerpc.GetReply)
	ss.getChannel <- getRequest{
		args,
		r,
	}
	getReply := <- r
	reply.Key = getReply.Key
	reply.Value = getReply.Value
	return nil
}

func (ss *cohortStorageServer) Put(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	fmt.Println("Enter Put")
	defer fmt.Println("Exit Put")

	r := make(chan error)
	ss.putChannel <- putRequest{
		args,
		r,
	}
	putReply := <-r
	fmt.Println(putReply)

	return nil
}

func (ss *masterStorageServer) masterServerHandler() {
	fmt.Println("Master server handler starting")
	defer fmt.Println("Master server handler quitting")

	for {
		select {
		case replyChan := <-ss.requestNodeId:

			client, err := rpc.DialHTTP("tcp", ":3000")
			if err != nil {
				log.Fatalln("dialing:", err)
			}
			fmt.Println("Connection established")
			newNode := &storagerpc.Node{
				ss.nextServerId,
				"localhost:3000",
				false,
				client,
			}
			ss.servers[newNode.NodeId] = newNode
			ss.nextServerId++
			replyChan <- ss.nextServerId

		case proposeReq := <-ss.proposeChannel:

			// Prepare for something
			args := &storagerpc.PrepareArgs{
				ss.nextTransactionId,
				proposeReq.args.Key,
				proposeReq.args.Value,
			}
			ss.nextTransactionId++
			go ss.prepareCohort(proposeReq, args)
		}
	}
}

func (ss *masterStorageServer) prepareCohort(proposeReq proposeRequest, args *storagerpc.PrepareArgs) {
	fmt.Println("Preparing cohort")
	stat := storagerpc.CommitStatus(storagerpc.Commit)
	var err error

	fmt.Println("Preparing cohort...")
	for nodeId, node := range ss.servers {
		fmt.Println(nodeId)
		var reply *storagerpc.PrepareReply
		err := node.Client.Call("CohortStorageServer.Prepare", args, &reply)
		if err != nil {
			stat = storagerpc.Rollback
		}
	}
	fmt.Println("Commiting cohort")
	for _, node := range ss.servers {
		args2 := &storagerpc.CommitArgs{
			args.TransactionId,
			stat,
		}
		var reply2 *storagerpc.CommitReply
		err := node.Client.Call("CohortStorageServer.Commit", args2, &reply2)
		if err != nil {
			proposeReq.reply <- err
		}
	}
	fmt.Println(err)
	proposeReq.reply <- nil
}

func (ss *cohortStorageServer) cohortServerHandler() {
	fmt.Println("Starting cohort server handler")

	for {
		select {
		case prepareReq := <-ss.prepareChannel:

			fmt.Println("Got a prepare request in cohortServerHandler")

			oldValue := ss.storage[prepareReq.args.Key]
			newValue := prepareReq.args.Value
			transactionId := prepareReq.args.TransactionId
			key := prepareReq.args.Key

			undoLogEntry := storagerpc.LogEntry{transactionId, key, oldValue}
			redoLogEntry := storagerpc.LogEntry{transactionId, key, newValue}
			ss.undoLog[transactionId] = undoLogEntry
			ss.redoLog[transactionId] = redoLogEntry

			prepareReq.reply <- nil

			// Make an undo log

		case commitReq := <-ss.commitChannel:

			if commitReq.args.Status == storagerpc.Commit {
				log, exists := ss.redoLog[commitReq.args.TransactionId]
				if !exists {
				}
				ss.storage[log.Key] = log.Value

				commitReq.reply <- nil // Handle this error if not found
			} else { // Abort
				log, exists := ss.undoLog[commitReq.args.TransactionId]
				if !exists {
				}
				ss.storage[log.Key] = log.Value

				commitReq.reply <- nil // Handle this error if not found
			}

		case getReq := <-ss.getChannel:
			fmt.Println("entering getReq case")
			val := ss.storage[getReq.args.Key]
			
			rsp := storagerpc.GetReply{
				getReq.args.Key,
				val,
			}
			getReq.reply <- rsp

		case putReq := <-ss.putChannel:
			fmt.Println("entering putReq case")
			args2 := &storagerpc.ProposeArgs{
				putReq.args.Key,
				putReq.args.Value,
			}
			var reply2 *storagerpc.ProposeReply
			doneCh := make(chan *rpc.Call, 1)
			ss.rpc.Go("MasterStorageServer.Propose", args2, &reply2, doneCh)
			go ss.handlePrepareRPC(putReq, doneCh)
		}
	}
}

func (ss *cohortStorageServer) handlePrepareRPC(putReq putRequest, doneCh chan *rpc.Call) {
	callObj := <-doneCh
	fmt.Println("Put err: ", callObj.Error)
	putReq.reply <- callObj.Error
}
