package main

import (
	//"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"runtime"
)

type Status int

const (
	NotReady = iota
	OK
)

type CommitStatus int

const (
	Commit = iota
	Rollback
)

type Node struct {
	NodeId   int
	HostPort string
	Master   bool
	client   *rpc.Client
}

type masterStorageServer struct {
	client            chan int
	self              Node
	servers           map[int]*Node
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
	undoLog map[int]LogEntry  // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	redoLog map[int]LogEntry  // TransactionId to Store 1. Key 2. TransactionID. (New)Value
}

type LogEntry struct {
	TransactionId int
	Key           string
	Value         string
}

type proposeRequest struct {
	args  *ProposeArgs
	reply chan error
}

type prepareRequest struct {
	args  *PrepareArgs
	reply chan error
}

type commitRequest struct {
	args  *CommitArgs
	reply chan error
}

type getRequest struct {
	args  *GetArgs
	reply chan GetReply
}

type putRequest struct {
	args  *PutArgs
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

	args := &PutArgs{"hello", "world"}
	var reply *PutReply
	err = client.Call("CohortStorageServer.Put", args, &reply)
	if err != nil {
		log.Fatalln("Put error:", err)
	}
	fmt.Println("Reply returned")
	fmt.Println(reply)

	args2 := &GetArgs{"hello"}
	var reply2 *GetReply
	err = client.Call("CohortStorageServer.Get", args2, &reply2)
	if err != nil {
		log.Fatalln("Get error: ", err)
	}

	fmt.Println("Reply val: ", reply2.Value)

}

// 2PC leader handler
func StartMasterServer() *masterStorageServer {
	fmt.Println("Entering StartMasterServer")
	defer fmt.Println("Exit StartMasterServer")

	server := new(masterStorageServer)
	server.nextServerId = 1
	server.nextTransactionId = 1
	server.requestNodeId = make(chan chan int)
	server.proposeChannel = make(chan proposeRequest)
	server.servers = make(map[int]*Node)

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
func StartCohortServer() *cohortStorageServer {
	fmt.Println("Entering StartCohortServer")
	defer fmt.Println("Exit StartCohortServer")

	client, err := rpc.DialHTTP("tcp", ":3000")
	if err != nil {
		log.Fatalln("dialing:", err)
	}
	args := &RegisterServerArgs{}
	var reply *RegisterServerReply
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
	server.redoLog = make(map[int]LogEntry)
	server.undoLog = make(map[int]LogEntry)

	rpc.RegisterName("CohortStorageServer", server)

	fmt.Println("Cohort server handler started...")
	go server.cohortServerHandler()

	return server
}

// RPC code

type RegisterServerArgs struct {
}

type RegisterServerReply struct {
	Status int
	NodeId int
}

type ProposeArgs struct {
	Key   string
	Value string
}

type ProposeReply struct {
	Status Status
}

type PrepareArgs struct {
	TransactionId int
	Key           string
	Value         string
}

type PrepareReply struct {
	Status Status
}

type CommitArgs struct {
	TransactionId int
	Status        CommitStatus
}

type CommitReply struct {
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	Key   string
	Value string
}

type PutArgs struct {
	Key   string
	Value string
}

type PutReply struct {
}

func (ss *masterStorageServer) RegisterServer(args *RegisterServerArgs, reply *RegisterServerReply) error {
	fmt.Println("Enter RegisterServer")
	defer fmt.Println("Exit RegisterServer")

	replyChan := make(chan int)
	ss.requestNodeId <- replyChan
	newNodeId := <-replyChan

	reply.Status = OK
	reply.NodeId = newNodeId
	return nil
}
func (ss *masterStorageServer) Propose(args *ProposeArgs, reply *ProposeReply) error {
	fmt.Println("Enter Propose")
	defer fmt.Println("Exit Propose")

	r := make(chan error)
	ss.proposeChannel <- proposeRequest{
		args,
		r,
	}
	error := <-r
	if error != nil {
		reply.Status = OK
	} else {
		reply.Status = OK

	}
	return error
}

func (ss *cohortStorageServer) Prepare(args *PrepareArgs, reply *PrepareReply) error {
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
		reply.Status = OK
	} else {
		reply.Status = OK

	}
	return error
}

func (ss *cohortStorageServer) Commit(args *CommitArgs, reply *CommitReply) error {
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

func (ss *cohortStorageServer) Get(args *GetArgs, reply *GetReply) error {
	fmt.Println("Enter Get")
	defer fmt.Println("Exit Get")

	r := make(chan GetReply)
	ss.getChannel <- getRequest{
		args,
		r,
	}
	getReply := <-r
	reply.Key = getReply.Key
	reply.Value = getReply.Value
	return nil
}

func (ss *cohortStorageServer) Put(args *PutArgs, reply *PutReply) error {
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
			newNode := &Node{
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
			args := &PrepareArgs{
				ss.nextTransactionId,
				proposeReq.args.Key,
				proposeReq.args.Value,
			}
			ss.nextTransactionId++
			go ss.prepareCohort(proposeReq, args)
		}
	}
}

func (ss *masterStorageServer) prepareCohort(proposeReq proposeRequest, args *PrepareArgs) {
	fmt.Println("Preparing cohort")
	stat := CommitStatus(Commit)
	var err error

	fmt.Println("Preparing cohort...")
	for nodeId, node := range ss.servers {
		fmt.Println(nodeId)
		var reply *PrepareReply
		err := node.client.Call("CohortStorageServer.Prepare", args, &reply)
		if err != nil {
			stat = Rollback
		}
	}
	fmt.Println("Commiting cohort")
	for _, node := range ss.servers {
		args2 := &CommitArgs{
			args.TransactionId,
			stat,
		}
		var reply2 *CommitReply
		err := node.client.Call("CohortStorageServer.Commit", args2, &reply2)
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

			undoLogEntry := LogEntry{transactionId, key, oldValue}
			redoLogEntry := LogEntry{transactionId, key, newValue}
			ss.undoLog[transactionId] = undoLogEntry
			ss.redoLog[transactionId] = redoLogEntry

			prepareReq.reply <- nil

			// Make an undo log

		case commitReq := <-ss.commitChannel:

			if commitReq.args.Status == Commit {
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
			rsp := GetReply{
				getReq.args.Key,
				val,
			}
			getReq.reply <- rsp

		case putReq := <-ss.putChannel:
			fmt.Println("entering putReq case")
			args2 := &ProposeArgs{
				putReq.args.Key,
				putReq.args.Value,
			}
			var reply2 *ProposeReply
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
