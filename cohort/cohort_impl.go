package main

import (
	"log"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/achadha235/p3/rpc/storagerpc"
	"github.com/achadha235/p3/util"
	"github.com/achadha235/p3/datatypes"
	"errors"
	"sync"
	"encoding/json"
)

type cohortStorageServer struct {
	rpc              *rpc.Client  
	nodeId        	 int
	master 			 bool
	masterHostPort   string
	selfHostPort	 string 
	servers map[int] *storagerpc.Node    // Consistent hashing ring. Empty if not instance is not master.
	
	storage map[string]string           // Key value storage
	locks 	map[string]*sync.RWMutex			// Locks for acessing storage
	undoLog map[int]storagerpc.LogEntry // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	redoLog map[int]storagerpc.LogEntry // TransactionId to Store 1. Key 2. TransactionID. (New)Value
}

func main() {
	fmt.Println("Hello world ");
	fmt.Println(storagerpc.Commit)
	ss, err := NewCohortServer("localhost:3000", "localhost:4000", 15, 1)

	fmt.Println(ss, err)
}

func NewCohortServer (masterHostPort string, selfHostPort string, nodeId int, numNodes int) (storagerpc.RemoteCohortServer, error) {
	ss := new(cohortStorageServer)

	ss.SetTickers()
	ss.nodeId = nodeId
	ss.masterHostPort = masterHostPort 
	ss.selfHostPort = selfHostPort 

	ss.servers = make(map[int]*storagerpc.Node)    // Consistent hashing ring. Empty if not instance is not master.
	ss.storage = make(map[string]string)
	ss.locks = make(map[string]*sync.RWMutex)
	ss.undoLog = make(map[int]storagerpc.LogEntry) // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	ss.redoLog = make(map[int]storagerpc.LogEntry) // TransactionId to Store 1. Key 2. TransactionID. (New)Value

	rpc.RegisterName("RemoteCohortServer", storagerpc.Wrap(ss))
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", selfHostPort)
	if e != nil {
		log.Fatalln("listen error:", e)
		return nil, e
	}
	go http.Serve(l, nil)

	if masterHostPort == selfHostPort {
		ss.master = true
	} else {
		ss.master = false
		fmt.Println("Cohort starting...")
		cli, err := util.TryDial(masterHostPort)
		if err != nil {
			return nil, err
		}
		ss.rpc = cli
		fmt.Println(cli)
	}
	return ss, nil
}

func (ss *cohortStorageServer) SetTickers(){
	ss.tickers := make(map[string]uint64)
	ss.tickers["APPL"] = 500
}
	
func (ss *cohortStorageServer) Commit(args *storagerpc.CommitArgs, reply *storagerpc.CommitReply) error {

//				if commitReq.args.Status == storagerpc.Commit {
// 				log, exists := ss.redoLog[commitReq.args.TransactionId]
// 				if !exists {
// 				}
// 				ss.storage[log.Key] = log.Value

// 				commitReq.reply <- nil // Handle this error if not found
// 			} else { // Abort
// 				log, exists := ss.undoLog[commitReq.args.TransactionId]
// 				if !exists {
// 				}
// 				ss.storage[log.Key] = log.Value

// 				commitReq.reply <- nil // Handle this error if not found
// 			}
	return errors.New("Not implemented")
}
func (ss *cohortStorageServer) ExecuteTransaction(args *storagerpc.TransactionArgs, reply *storagerpc.TransactionReply) error {

// 			args2 := &storagerpc.ProposeArgs{
// 				putReq.args.Key,
// 				putReq.args.Value,
// 			}
// 			var reply2 *storagerpc.ProposeReply
// 			doneCh := make(chan *rpc.Call, 1)
// 			ss.rpc.Go("MasterStorageServer.Propose", args2, &reply2, doneCh)
// 			go ss.handlePrepareRPC(putReq, doneCh)
	return errors.New("Not implemented")
}
func (ss *cohortStorageServer) Prepare(args *storagerpc.PrepareArgs, reply *storagerpc.PrepareReply) error {
	op := args.Name
	switch op {
	case op == datatypes.AddUser:
		key := "user-" + args.Data.User.UserID
		userString, exists := ss.storage[key]
		if exists {
			reply.Status = Exists
		} else {
			newB, err := json.Marshal(args.Data.User)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			} 
			newValue := string(newB)

			undoLogEntry := storagerpc.LogEntry{transactionId, key, userString}
			redoLogEntry := storagerpc.LogEntry{transactionId, key, newValue}
			ss.undoLog[args.TransactionId] = undoLogEntry
			ss.redoLog[args.TransactionId] = redoLogEntry	
			reply.Staus = datatypes.OK
			return nil
		}
	case op == datatypes.AddTeam:
		key := "team-" + args.Data.Team.TeamID
		teamString, exists := ss.storage[key]
		if exists {
			reply.Status = Exists
		} else {
			newB, err := json.Marshal(args.Data.Team)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			} 
			newValue := string(newB)

			undoKvp := []KeyValuePair{KeyValuePair{key, teamString}}
			redoKvp := []KeyValuePair{KeyValuePair{key, newValue}}

			undoLogEntry := storagerpc.LogEntry{args.TransactionId, undoKvp}
			redoLogEntry := storagerpc.LogEntry{args.TransactionId, redoKvp}
			ss.undoLog[args.TransactionId] = undoLogEntry
			ss.redoLog[args.TransactionId] = redoLogEntry	
			reply.Staus = datatypes.OK

			return nil

	case op == datatypes.AddUserToTeamList
		undoKvp := make([]KeyValuePair, 1)
		redoKvp := make([]KeyValuePair, 1)

		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Team.Team.TeamID
		isCorrectServer = ss.isCorrectServer(teamKey);

		if isCorrectServer {
			teamString, teamExists := ss.storage[teamKey]
			if !teamExists { 
				reply.Status = datatypes.NoSuchTeam
				return nil
			}
			var team datatypes.Team
			err := json.Unmarshal([]byte(teamString), &team)
			if err {
				reply.Status = datatypes.BadData
				return nil
			}

			if args.Data.Pw == team.HashPW {
				reply.Status = datatypes.PermissionDenied
				return nil
			}

			undoKvp[0] = KeyValuePair{teamKey, teamString}
			team.Teams = Append(team.Teams, teamKey)
			newTeamBytes, err := json.Marshal(team)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			redoKvp[0] = KeyValuePair{teamKey, string(newTeamBytes)}
		}

		undoLogEntry := storagerpc.LogEntry{args.TransactionId, undoKvp}
		redoLogEntry := storagerpc.LogEntry{args.TransactionId, redoKvp}
		ss.undoLog[args.TransactionId] = undoLogEntry
		ss.redoLog[args.TransactionId] = redoLogEntry	

		if isCorrectServer {
			reply.Status = datatypes.OK
		} else {
			reply.Status = datatypes.WrongServer
		}
		return nil

	case op == datatypes.AddTeamToUserList:

		undoKvp := make([]KeyValuePair, 1)
		redoKvp := make([]KeyValuePair, 1)

		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Team.Team.TeamID

		if ss.isCorrectServer(userKey){
			userString, userExists := ss.storage[userKey]
			if !userExists { 
				reply.Status = datatypes.NoSuchUser
				return nil
			}
			var user datatypes.User
			err := json.Unmarshal([]byte(userString), &user)
			if err {
				reply.Status = BadData
				return nil
			}
			undoKvp[0] = KeyValuePair{userKey, userString}
			user.Teams = Append(user.Teams, teamKey)
			newUserBytes, err := json.Marshal(user)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			redoKvp[0] = KeyValuePair{userKey, string(newUserBytes)}
		} else {
			reply.Status = datatypes.WrongServer
			return nil
		}
		
		undoLogEntry := storagerpc.LogEntry{args.TransactionId, undoKvp}
		redoLogEntry := storagerpc.LogEntry{args.TransactionId, redoKvp}
		ss.undoLog[args.TransactionId] = undoLogEntry
		ss.redoLog[args.TransactionId] = redoLogEntry	

		reply.Status = datatypes.OK
		return nil
						
	case op == RemoveUserFromTeamList:
		undoKvp := make([]KeyValuePair, 2)
		redoKvp := make([]KeyValuePair, 2)

		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Team.Team.TeamID
		isCorrectServer = ss.isCorrectServer(teamKey);

		if isCorrectServer {
			teamString, teamExists := ss.storage[teamKey]
			if !teamExists { 
				reply.Status = datatypes.NoSuchTeam
				return nil
			}
			var team datatypes.Team
			err := json.Unmarshal([]byte(teamString), &team)
			if err {
				reply.Status = BadData
				return nil
			}
			undoKvp[0] = KeyValuePair{teamKey, teamString}

			team.Teams = Remove(team.Teams, teamKey)
			newTeamBytes, err := json.Marshal(team)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			redoKvp[0] = KeyValuePair{teamKey, string(newTeamBytes)}
		}

		undoLogEntry := storagerpc.LogEntry{args.TransactionId, undoKvp}
		redoLogEntry := storagerpc.LogEntry{args.TransactionId, redoKvp}
		ss.undoLog[args.TransactionId] = undoLogEntry
		ss.redoLog[args.TransactionId] = redoLogEntry	

		if isCorrectServer {
			reply.Status = datatypes.OK
		} else {
			reply.Status = datatypes.WrongServer
		}
		return nil


	case op == RemoveTeamFromUserList:
		undoKvp := make([]KeyValuePair, 1)
		redoKvp := make([]KeyValuePair, 1)

		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Team.Team.TeamID

		if ss.isCorrectServer(userKey){
			userString, userExists := ss.storage[userKey]
			if !userExists { 
				reply.Status = datatypes.NoSuchUser
				return nil
			}
			var user datatypes.User
			err := json.Unmarshal([]byte(userString), &user)
			if err {
				reply.Status = BadData
				return nil
			}
			undoKvp[0] = KeyValuePair{userKey, userString}
			user.Teams = Remove(user.Teams, teamKey)
			newUserBytes, err := json.Marshal(user)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			redoKvp[0] = KeyValuePair{userKey, string(newUserBytes)}
		} else {
			reply.Status = datatypes.WrongServer
			return nil
		}
		
		undoLogEntry := storagerpc.LogEntry{args.TransactionId, undoKvp}
		redoLogEntry := storagerpc.LogEntry{args.TransactionId, redoKvp}
		ss.undoLog[args.TransactionId] = undoLogEntry
		ss.redoLog[args.TransactionId] = redoLogEntry	

		reply.Status = datatypes.OK
		return nil		


	case op == datatypes.Buy:
		undoKvp := make([]KeyValuePair, 1)
		redoKvp := make([]KeyValuePair, 1)

		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Team.Team.TeamID

		if ss.isCorrectServer(userKey){
			userString, userExists := ss.storage[userKey]
			if !userExists { 
				reply.Status = datatypes.NoSuchUser
				return nil
			}
			var user datatypes.User
			err := json.Unmarshal([]byte(userString), &user)
			if err {
				reply.Status = BadData
				return nil
			}
			undoKvp[0] = KeyValuePair{userKey, userString}
			user.Teams = Remove(user.Teams, teamKey)
			newUserBytes, err := json.Marshal(user)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			redoKvp[0] = KeyValuePair{userKey, string(newUserBytes)}
		} else {
			reply.Status = datatypes.WrongServer
			return nil
		}
		
		undoLogEntry := storagerpc.LogEntry{args.TransactionId, undoKvp}
		redoLogEntry := storagerpc.LogEntry{args.TransactionId, redoKvp}
		ss.undoLog[args.TransactionId] = undoLogEntry
		ss.redoLog[args.TransactionId] = redoLogEntry	

		reply.Status = datatypes.OK
		return nil		


	case op == datatypes.Sell:
		




	}
	return errors.New("Not implemented")
}

func (ss *cohortStorageServer) Get(args *storagerpc.GetArgs, reply *storagerpc.GetReply) error {
	lock := ss.getOrCreateRWMutex(args.Key)
	reply.Key = args.Key

	lock.RLock()
	value, exists := ss.storage[args.Key]
	reply.Value = value
	if !exists {
		reply.StorageStatus = storagerpc.KeyNotFound
	} else {
		reply.StorageStatus = storagerpc.OK
	}
	lock.RUnlock()
	reply.Status = datatypes.OK 

	return nil
}

func (ss *cohortStorageServer) getOrCreateRWMutex(key string) (*sync.RWMutex) {
	mtx, exists := ss.locks[key]
	if !exists {
		newLock := new(sync.RWMutex)
		ss.locks[key] = newLock
		return newLock
	} else {
		return mtx
	}
}

func (ss *cohortStorageServer) isCorrectServer(key string) bool {
	l := len(ss.servers)
	if l == 1 {
		return true
	}
	hashed := util.StoreHash(key)
	last := ss.servers[l - 1]
	if hash > last.NodeId && ss.nodes[0].NodeId == ss.nodeId {
		return true
	}
	for i := 0; i < l; i++ {
		if hash <= ss.servers[i].NodeID {
			if ss.nodeId == ss.nodes[i].NodeID {
				return true
			}
		}
	}
}

// Does not preserve order
func Remove(str string, s []string) ([]string){
	found := false
	at := 0
	for i := 0; !found && i < len() &&; i++ {
		if list[i] == id {
			found = true
			at = i
		}
	}
	if !found {
		return list
	} else {
		p := len(list) - 1
		list[at] = list[p]
		return list[:p]
	}
}

// func (ss *cohortStorageServer) createLogs(args *storagerpc.PrepareArgs){
// 	switch args.
// }

	// ss.rpc = client
	// ss.storage = make(map[string]string)
	// ss.redoLog = make(map[int]storagerpc.LogEntry)
	// ss.undoLog = make(map[int]storagerpc.LogEntry)
	
	// rpc.RegisterName("RemoteCohortServer", RemoteCohortServer(ss))
	// rpc.HandleHTTP()
	// l, e := net.Listen("tcp", ":3000")
	// if e != nil {
	// 	log.Fatalln("listen error:", e)
	// }
	// go http.Serve(l, nil)

	// if masterHostPort == selfHostPort {
	// 	joined := 0
	// 	for joined < numNodes {
	// 		// Wait for servers to join.
	// 		fmt.Println("Waiting for servers to join")
	// 	}
	// 	// Needs to wait for expected ring to join
	// } else {
	// // Dial the master. 
		
	// 		
	// 		
	// 		
	// 		

	// }


// // 2PC cohort handler
// func StartCohortServer() storagerpc.CohortStorageServer {

// 	client, err := rpc.DialHTTP("tcp", ":3000")
// 	if err != nil {
// 		log.Fatalln("dialing:", err)
// 	}
// 	args := &storagerpc.RegisterServerArgs{}
// 	var reply *storagerpc.RegisterServerReply
// 	err = client.Call("MasterStorageServer.RegisterServer", args, &reply)
// 	if err != nil {
// 		log.Fatalln("Register error:", err)
// 	}

// 	// Set up cohort rpc
// 	

// 	
// 	
// 	
// 	
// 	

// 	
// 	
// 	

// 	rpc.RegisterName("CohortStorageServer", server)

// 	go server.cohortServerHandler()

// 	return server
// }






// func (ss *cohortStorageServer) Get(args *storagerpc.GetArgs, reply *storagerpc.GetReply) error {
// 	r := make(chan storagerpc.GetReply)
// 	ss.getChannel <- getRequest{
// 		args,
// 		r,
// 	}
// 	getReply := <-r
// 	reply.Key = getReply.Key
// 	reply.Value = getReply.Value
// 	return nil
// }



// func (ss *cohortStorageServer) cohortServerHandler() {
// 	for {
// 		select {
// 		case prepareReq := <-ss.prepareChannel:

// 			

// 		case commitReq := <-ss.commitChannel:
// 			

// 		case getReq := <-ss.getChannel:


// 		case putReq := <-ss.putChannel:

// 		}
// 	}
// }

// func (ss *cohortStorageServer) handlePrepareRPC(putReq putRequest, doneCh chan *rpc.Call) {
// 	callObj := <-doneCh
// 	putReq.reply <- callObj.Error
// }
