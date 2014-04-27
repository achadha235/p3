package storageserver

import (
	"github.com/achadha235/p3/rpc/storagerpc"
	"log"
	"net"
	"net/http"
	"net/rpc"
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

	//isRingMaster bool
	//ring map[int]

	storage map[string]string // Key value storage
	undoLog map[int]storagerpc.LogEntry  // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	redoLog map[int]storagerpc.LogEntry  // TransactionId to Store 1. Key 2. TransactionID. (New)Value 
										 // Add a commit 
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

	args2 := &storagerpc.GetArgs{"hello"}
	var reply2 *storagerpc.GetReply
	err = client.Call("CohortStorageServer.Get", args2, &reply2)
	if err != nil {
		log.Fatalln("Get error: ", err)
	}


}

// 2PC leader handler
func StartMasterServer() storagerpc.MasterStorageServer {

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

	go server.cohortServerHandler()

	return server
}



func (ss *masterStorageServer) RegisterServer(args *storagerpc.RegisterServerArgs, reply *storagerpc.RegisterServerReply) error {

	replyChan := make(chan int)
	ss.requestNodeId <- replyChan
	newNodeId := <-replyChan

	reply.Status = storagerpc.OK
	reply.NodeId = newNodeId
	return nil
}
func (ss *masterStorageServer) Propose(args *storagerpc.ProposeArgs, reply *storagerpc.ProposeReply) error {

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

	r := make(chan error)
	ss.prepareChannel <- prepareRequest{
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

func (ss *cohortStorageServer) Commit(args *storagerpc.CommitArgs, reply *storagerpc.CommitReply) error {

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


// CHANGE
func (ss *cohortStorageServer) Put(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	r := make(chan error)
	ss.putChannel <- putRequest{
		args,
		r,
	}
	putReply := <-r
	return nil
}



func (ss *cohortStorageServer) TransactionIsLegal(tx *storagerpc.TransactionArgs) bool {
	switch  {
	case tx.Method == storagerpc.CreateUser:
		return true
	case tx.Method == storagerpc.CreateTeam:
		return true
	case tx.Method == storagerpc.JoinTeam:
		return true
	case tx.Method == storagerpc.LeaveTeam:
		return true
	case tx.Method == storagerpc.PlaceOrder:
		return true
	default : 
		return true
	}

}
func (ss *cohortStorageServer) ExecuteTransaction(tx *storagerpc.TransactionArgs, reply *storagerpc.TransactionArgs) error {
	switch { // Fill in logic to atomically execute a particular type of transaction 
	case tx.Method == storagerpc.CreateUser:
		return nil
	case tx.Method == storagerpc.CreateTeam:
		return nil
	case tx.Method == storagerpc.JoinTeam:
		return nil
	case tx.Method == storagerpc.LeaveTeam:
		return nil
	case tx.Method == storagerpc.PlaceOrder:
		return nil
	default: 
		return nil
	}
}

func (ss *masterStorageServer) masterServerHandler() {
	for {
		select {
		case replyChan := <-ss.requestNodeId:
			client, err := rpc.DialHTTP("tcp", ":3000")
			if err != nil {
				log.Fatalln("dialing:", err)
			}
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
	stat := storagerpc.CommitStatus(storagerpc.Commit)
	var err error
	for nodeId, node := range ss.servers {
		var reply *storagerpc.PrepareReply
		err := node.Client.Call("CohortStorageServer.Prepare", args, &reply)
		if err != nil {
			stat = storagerpc.Rollback
		}
	}
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
	proposeReq.reply <- nil
}

func (ss *cohortStorageServer) cohortServerHandler() {

	for {
		select {
		case prepareReq := <-ss.prepareChannel:


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
			val := ss.storage[getReq.args.Key]
			
			rsp := storagerpc.GetReply{
				getReq.args.Key,
				val,
			}
			getReq.reply <- rsp

		case putReq := <-ss.putChannel:
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
	putReq.reply <- callObj.Error
}
