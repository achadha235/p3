package cohort

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/rpc/storagerpc"
	"github.com/achadha235/p3/util"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type cohortStorageServer struct {
	rpc            *rpc.Client
	nodeId         int
	master         bool
	masterHostPort string
	selfHostPort   string
	servers        map[int]*storagerpc.Node // Consistent hashing ring. Empty if not instance is not master.
	tickers        map[string]uint64
	numNodes       int

	storage map[string]string        // Key value storage
	locks   map[string]*sync.RWMutex // Locks for acessing storage
	undoLog map[int]LogEntry         // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	redoLog map[int]LogEntry         // TransactionId to Store 1. Key 2. TransactionID. (New)Value
}

type KeyValuePair struct {
	Key   string
	Value string
}

type LogEntry struct {
	TransactionId int
	Logs          []KeyValuePair
}

func NewCohortServer(masterHostPort string, selfHostPort string, nodeId int, numNodes int) (storagerpc.RemoteCohortServer, error) {
	ss := new(cohortStorageServer)

	ss.nodeId = nodeId
	ss.masterHostPort = masterHostPort
	ss.selfHostPort = selfHostPort

	ss.servers = make(map[int]*storagerpc.Node) // Consistent hashing ring. Empty if not instance is not master.
	ss.storage = make(map[string]string)
	ss.locks = make(map[string]*sync.RWMutex)
	ss.undoLog = make(map[int]LogEntry) // TransactionId to Store 1. Key 2. TransactionId. (Old)Value
	ss.redoLog = make(map[int]LogEntry) // TransactionId to Store 1. Key 2. TransactionID. (New)Value

	ss.numNodes = numNodes
	ss.tickers = make(map[string]uint64)
	ss.setTickers()

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
		//TODO:  if master we should be listening for other clients to connect!!!!!
		// Add ss.nodes []storagerpc.Node
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

func (ss *cohortStorageServer) setTickers() {
	ss.tickers["APPL"] = 500
}

func (ss *cohortStorageServer) Commit(args *storagerpc.CommitArgs, reply *storagerpc.CommitReply) error {
	var exists bool
	var commitLog LogEntry
	if args.Status == storagerpc.Commit {
		commitLog, exists := ss.redoLog[args.TransactionId]
	} else {
		commitLog, exists := ss.undoLog[args.TransactionId]
	}
	if !exists {
		return errors.New("Commit without prepare not possible")
	}
	for i := 0; i < len(commitLog.Logs); i++ {
		mtx := ss.getOrCreateRWMutex(commitLog.Logs[i].Key)
		mtx.Lock()
		ss.storage[commitLog.Logs[i].Key] = commitLog.Logs[i].Value

		mtx.Unlock()
	}
	return nil
}

func (ss *cohortStorageServer) UpdateLogs(transactionId int, undoKVP, redoKVP []KeyValuePair) {
	undoLogEntry := LogEntry{
		TransactionId: transactionId,
		Logs:          undoKVP,
	}
	redoLogEntry := LogEntry{
		TransactionId: transactionId,
		Logs:          redoKVP,
	}

	ss.undoLog[transactionId] = undoLogEntry
	ss.redoLog[transactionId] = redoLogEntry
}

func (ss *cohortStorageServer) Prepare(args *storagerpc.PrepareArgs, reply *storagerpc.PrepareReply) error {
	op := args.Name

	switch {
	case op == datatypes.AddUser:
		key := "user-" + args.Data.User.UserID
		userString, exists := ss.storage[key]
		if exists {
			reply.Status = datatypes.Exists
			return nil
		} else {
			newB, err := json.Marshal(args.Data.User)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			newValue := string(newB)

			undoKvp := []KeyValuePair{KeyValuePair{Key: key, Value: userString}}
			redoKvp := []KeyValuePair{KeyValuePair{Key: key, Value: newValue}}
			ss.UpdateLogs(args.TransactionId, undoKvp, redoKvp)
			reply.Status = datatypes.OK
			return nil
		}
	case op == datatypes.AddTeam:
		key := "team-" + args.Data.Team.TeamID
		teamString, exists := ss.storage[key]
		if exists {
			reply.Status = datatypes.Exists
			return nil
		}

		newB, err := json.Marshal(args.Data.Team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}
		newValue := string(newB)

		undoKvp := []KeyValuePair{KeyValuePair{Key: key, Value: teamString}}
		redoKvp := []KeyValuePair{KeyValuePair{Key: key, Value: newValue}}

		ss.UpdateLogs(args.TransactionId, undoKvp, redoKvp)
		reply.Status = datatypes.OK

		return nil

	case op == datatypes.AddUserToTeamList:
		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Data.Team.TeamID

		if ss.isCorrectServer(args.Data.Team.TeamID) {
			reply.Status = datatypes.BadData
			return errors.New("Wrong Server")
		}

		teamString, teamExists := ss.storage[teamKey]
		if !teamExists {
			reply.Status = datatypes.NoSuchTeam
			return nil
		}
		var team datatypes.Team
		err := json.Unmarshal([]byte(teamString), &team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		err = bcrypt.CompareHashAndPassword([]byte(team.HashPW), []byte(args.Data.Pw))
		if err != nil {
			reply.Status = datatypes.PermissionDenied
			return nil
		}

		team.Users = append(team.Users, args.Data.User.UserID)
		newTeamBytes, err := json.Marshal(team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		undoKvp := []KeyValuePair{KeyValuePair{Key: teamKey, Value: teamString}}
		redoKvp := []KeyValuePair{KeyValuePair{Key: teamKey, Value: string(newTeamBytes)}}

		ss.UpdateLogs(args.TransactionId, undoKvp, redoKvp)
		reply.Status = datatypes.OK
		return nil

	case op == datatypes.AddTeamToUserList:
		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Data.Team.TeamID

		if !ss.isCorrectServer(args.Data.User.UserID) {
			reply.Status = datatypes.BadData
			return errors.New("Wrong Server")
		}

		userString, userExists := ss.storage[userKey]
		if !userExists {
			reply.Status = datatypes.NoSuchUser
			return nil
		}
		var user datatypes.User
		err := json.Unmarshal([]byte(userString), &user)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}
		user.Teams = append(user.Teams, args.Data.Team.TeamID)
		newUserBytes, err := json.Marshal(user)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		undoKvp := []KeyValuePair{KeyValuePair{Key: userKey, Value: userString}}
		redoKvp := []KeyValuePair{KeyValuePair{Key: userKey, Value: string(newUserBytes)}}

		ss.UpdateLogs(args.TransactionId, undoKvp, redoKvp)
		reply.Status = datatypes.OK
		return nil

	case op == datatypes.RemoveUserFromTeamList:
		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Data.Team.TeamID

		if !ss.isCorrectServer(args.Data.Team.TeamID) {
			reply.Status = datatypes.BadData
			return errors.New("Wrong Server")
		}

		teamString, teamExists := ss.storage[teamKey]
		if !teamExists {
			reply.Status = datatypes.NoSuchTeam
			return nil
		}
		var team datatypes.Team
		err := json.Unmarshal([]byte(teamString), &team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		team.Users = remove(team.Users, args.Data.User.UserID)
		newTeamBytes, err := json.Marshal(team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		undoKvp := []KeyValuePair{KeyValuePair{Key: teamKey, Value: teamString}}
		redoKvp := []KeyValuePair{KeyValuePair{Key: teamKey, Value: string(newTeamBytes)}}

		ss.UpdateLogs(args.TransactionId, undoKvp, redoKvp)
		reply.Status = datatypes.OK
		return nil

	case op == datatypes.RemoveTeamFromUserList:
		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Data.Team.TeamID

		if !ss.isCorrectServer(args.Data.User.UserID) {
			reply.Status = datatypes.BadData
			return errors.New("Wrong Server")
		}

		userString, userExists := ss.storage[userKey]
		if !userExists {
			reply.Status = datatypes.NoSuchUser
			return nil
		}
		var user datatypes.User
		err := json.Unmarshal([]byte(userString), &user)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}
		user.Teams = remove(user.Teams, args.Data.Team.TeamID)
		newUserBytes, err := json.Marshal(user)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		undoKvp := []KeyValuePair{KeyValuePair{Key: userKey, Value: userString}}
		redoKvp := []KeyValuePair{KeyValuePair{Key: userKey, Value: string(newUserBytes)}}

		ss.UpdateLogs(args.TransactionId, undoKvp, redoKvp)
		reply.Status = datatypes.OK
		return nil

	case op == datatypes.Buy:
		// In an operation there is only one request
		req := args.Data.Requests[0]
		tickerName := req.Ticker

		// keys for lookup in storage map
		tickerKey := "ticker-" + tickerName
		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Data.Team.TeamID
		holdingKey := "holding-" + args.Data.Team.TeamID + "-" + tickerName

		if !ss.isCorrectServer(args.Data.Team.TeamID) {
			reply.Status = datatypes.BadData
			return errors.New("Wrong Server")
		}

		teamStr, ok := ss.storage[teamKey]
		if !ok {
			reply.Status = datatypes.NoSuchTeam
			return nil
		}

		var team datatypes.Team
		err := json.Unmarshal([]byte(teamStr), &team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		tickerStr, ok := ss.storage[tickerKey]
		if !ok {
			reply.Status = datatypes.NoSuchTicker
			return nil
		}

		var ticker datatypes.Ticker
		err = json.Unmarshal([]byte(tickerStr), &ticker)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		cost := req.Quantity * ticker.Price
		var newBalance uint64
		if newBalance = team.Balance - cost; newBalance < 0 {
			reply.Status = datatypes.InsufficientQuantity
			return nil
		}

		undoKVP := make([]KeyValuePair, 0)
		redoKVP := make([]KeyValuePair, 0)

		var holding datatypes.Holding
		var oldValue string
		if holdingStr, ok := team.Holdings[tickerName]; !ok {
			// no holding exists for requested ticker
			oldValue = ""
			holding = datatypes.Holding{
				Ticker:   tickerName,
				Quantity: req.Quantity,
				Acquired: time.Now(),
			}
		} else {
			oldValue = holdingStr
			err = json.Unmarshal([]byte(holdingStr), &holding)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}

			holding.Quantity += req.Quantity
			holding.Acquired = time.Now()
		}

		newHoldingBytes, err := json.Marshal(holding)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		undoKVP = append(undoKVP, KeyValuePair{Key: holdingKey, Value: oldValue})
		redoKVP = append(redoKVP, KeyValuePair{Key: holdingKey, Value: string(newHoldingBytes)})

		// save the old team to Undo log before updating
		undoKVP = append(undoKVP, KeyValuePair{Key: teamKey, Value: teamStr})

		// save the updated holdingID in the team's holding list
		team.Holdings[tickerName] = holdingKey
		// save the new balance
		team.Balance = newBalance

		newTeamBytes, err := json.Marshal(team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		redoKVP = append(redoKVP, KeyValuePair{Key: teamKey, Value: string(newTeamBytes)})

		ss.UpdateLogs(args.TransactionId, undoKVP, redoKVP)
		reply.Status = datatypes.OK
		return nil

	case op == datatypes.Sell:
		// In an operation there is only one request
		req := args.Data.Requests[0]
		tickerName := req.Ticker

		// keys for lookup in storage map
		tickerKey := "ticker-" + tickerName
		userKey := "user-" + args.Data.User.UserID
		teamKey := "team-" + args.Data.Team.TeamID
		/*		holdingKey := "holding-" + args.Data.TeamID + "-" + tickerName*/

		if !ss.isCorrectServer(args.Data.Team.TeamID) {
			reply.Status = datatypes.BadData
			return errors.New("Wrong Server")
		}

		teamStr, ok := ss.storage[teamKey]
		if !ok {
			reply.Status = datatypes.NoSuchTeam
			return nil
		}
		var team datatypes.Team
		err := json.Unmarshal([]byte(teamStr), &team)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		tickerStr, ok := ss.storage[tickerKey]
		if !ok {
			reply.Status = datatypes.NoSuchTicker
			return nil
		}
		var ticker datatypes.Ticker
		err = json.Unmarshal([]byte(tickerStr), &ticker)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		// if no holding found, then you can't sell anything
		holdingKey, ok := team.Holdings[tickerName]
		if !ok {
			reply.Status = datatypes.InsufficientQuantity
			return nil
		}
		var holding datatypes.Holding
		holdingStr, ok := ss.storage[holdingKey]
		if !ok {
			reply.Status = datatypes.NoSuchHolding
			return nil
		}
		err = json.Unmarshal([]byte(holdingStr), &holding)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}

		// Check if transaction is possible

		// not enough shares on team
		if holding.Quantity < req.Quantity {
			reply.Status = datatypes.InsufficientQuantity
			return nil
		}

		undoKVP := make([]KeyValuePair, 0)
		redoKVP := make([]KeyValuePair, 0)

		undoKVP = append(undoKVP, KeyValuePair{Key: teamKey, Value: teamStr}, KeyValuePair{Key: holdingKey, Value: holdingStr})

		profit := ticker.Price * req.Quantity

		// update the new balance
		team.Balance += profit

		// update the holding information
		holding.Quantity -= req.Quantity
		holding.Acquired = time.Now()

		// if they sold all the stock, remove the holding completely
		if holding.Quantity == 0 {
			// do the actual deletes in Commit???
			/*			delete(ss.storage, holdingKey)*/
			delete(team.Holdings, tickerName)
			redoKVP = append(redoKVP, KeyValuePair{Key: holdingKey, Value: ""})
		} else {
			newHoldingBytes, err := json.Marshal(holding)
			if err != nil {
				reply.Status = datatypes.BadData
				return nil
			}
			redoKVP = append(redoKVP, KeyValuePair{Key: holdingKey, Value: string(newHoldingBytes)})
		}

		newTeamBytes, err := json.Marshal(holding)
		if err != nil {
			reply.Status = datatypes.BadData
			return nil
		}
		redoKVP = append(redoKVP, KeyValuePair{Key: teamKey, Value: string(newTeamBytes)})

		ss.UpdateLogs(args.TransactionId, undoKVP, redoKVP)
		reply.Status = datatypes.OK
		return nil
	}

	reply.Status = datatypes.NoSuchAction
	return errors.New("Operation not defined on cohort")
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

func (ss *cohortStorageServer) getOrCreateRWMutex(key string) *sync.RWMutex {
	mtx, exists := ss.locks[key]
	if !exists {
		newLock := new(sync.RWMutex)
		ss.locks[key] = newLock
		return newLock
	} else {
		return mtx
	}
}

func (ss *cohortStorageServer) GetServers(args *storagerpc.GetServersArgs, reply *storagerpc.GetServersReply) error {
	if len(ss.servers) < ss.numNodes {
		reply.Status = storagerpc.NotReady
		return nil
	}

	reply.Status = storagerpc.OK
	reply.Servers = ss.servers
	return nil
}

func (ss *cohortStorageServer) isCorrectServer(key string) bool {
	l := len(ss.servers)
	if l == 1 {
		return true
	}
	hashed := util.StoreHash(key)
	last := ss.servers[l-1]
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

// Remove an string from a slice of strings
func remove(list []string, id string) []string {
	for i := 0; i < len(list); i++ {
		if list[i] == id {
			return append(list[0:i], list[i+1]...)
		}
	}

	return list
}
