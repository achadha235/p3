package coordinator

import (
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/rpc/storagerpc"
	"github.com/achadha235/p3/util"
	"log"
	"net/rpc"
	"time"
)

type coordinator struct {
	masterStorageServer *rpc.Client            // RPC connection to masterStorage
	servers             []storagerpc.Node      // slice of storage servers
	connections         map[string]*rpc.Client // [hostport] --> client conn
	nextOperationId     int
}

// [hostport] --> slice of prepareArgs for operations that the node must execute
type PrepareMap map[string][]*storagerpc.PrepareArgs

func StartCoordinator(masterServerHostPort string) (Coordinator, error) {
	cli, err := util.TryDial(masterServerHostPort)
	if err != nil {
		return nil, err
	}

	args := &storagerpc.GetServersArgs{}
	var reply storagerpc.GetServersReply

	// attempt to get the list of servers in the ring from the MasterStorageServer
	var servers []storagerpc.Node
	for t := util.MaxConnectAttempts; ; t-- {
		err := cli.Call("CohortStorageServer.GetServers", args, &reply)
		log.Println("status: ", reply.Status)
		log.Println("Ok status: ", storagerpc.OK)
		if reply.Status == storagerpc.OK {
			servers = reply.Servers
			break
		} else if t <= 0 {
			log.Println("SS Not Ready")
			// StorageServers not ready
			return nil, err
		}
		// Wait a second before retrying
		time.Sleep(time.Second)
	}

	// create conns to be cached and add masterServer to map
	conns := make(map[string]*rpc.Client)
	conns[masterServerHostPort] = cli

	// create the coordinator
	coord := &coordinator{
		masterStorageServer: cli,
		servers:             servers,
		connections:         conns,
		nextOperationId:     1,
	}

	return coord, nil
}

/* PerformTransaction will break apart a transaction and
   send the appropriate updates to each corresponding
   StorageServer using a 2PC propose call */
func (coord *coordinator) PerformTransaction(name datatypes.TransactionType, data datatypes.DataArgs) (datatypes.Status, error) {
	prepareMap := make(PrepareMap)

	switch name {
	// Add user data on node
	case datatypes.CreateUser:
		args := &storagerpc.PrepareArgs{
			TransactionId: coord.nextOperationId,
			Name:          datatypes.AddUser,
			Data:          data,
		}

		ss := util.FindServerFromKey(data.User.UserID, coord.servers)
		log.Println("ss: ", ss)
		prepareMap[ss.HostPort] = append(prepareMap[ss.HostPort], args)

		coord.nextOperationId++

	// Add team data on node
	case datatypes.CreateTeam:
		args := &storagerpc.PrepareArgs{
			TransactionId: coord.nextOperationId,
			Name:          datatypes.AddTeam,
			Data:          data,
		}

		ss := util.FindServerFromKey("team-"+data.Team.TeamID, coord.servers)
		prepareMap[ss.HostPort] = append(prepareMap[ss.HostPort], args)

		coord.nextOperationId++

	// Add user data to team list and vice-versa for the respective nodes
	case datatypes.JoinTeam:
		// create args for call to node with team info and update prepareMap
		teamArgs := &storagerpc.PrepareArgs{
			TransactionId: coord.nextOperationId,
			Name:          datatypes.AddUserToTeamList,
			Data:          data,
		}

		ssTeam := util.FindServerFromKey("team-"+data.Team.TeamID, coord.servers)
		prepareMap[ssTeam.HostPort] = append(prepareMap[ssTeam.HostPort], teamArgs)

		coord.nextOperationId++

		// create args for call to node with user info and update prepareMap
		userArgs := &storagerpc.PrepareArgs{
			TransactionId: coord.nextOperationId,
			Name:          datatypes.AddTeamToUserList,
			Data:          data,
		}

		ssUser := util.FindServerFromKey(data.User.UserID, coord.servers)
		prepareMap[ssUser.HostPort] = append(prepareMap[ssUser.HostPort], userArgs)

		coord.nextOperationId++

	// Remove user data from team list and vice-versa for the respective nodes
	case datatypes.LeaveTeam:
		teamArgs := &storagerpc.PrepareArgs{
			TransactionId: coord.nextOperationId,
			Name:          datatypes.RemoveUserFromTeamList,
			Data:          data,
		}

		ssTeam := util.FindServerFromKey(data.Team.TeamID, coord.servers)
		prepareMap[ssTeam.HostPort] = append(prepareMap[ssTeam.HostPort], teamArgs)

		coord.nextOperationId++

		userArgs := &storagerpc.PrepareArgs{
			TransactionId: coord.nextOperationId,
			Name:          datatypes.RemoveTeamFromUserList,
			Data:          data,
		}

		ssUser := util.FindServerFromKey(data.User.UserID, coord.servers)
		prepareMap[ssUser.HostPort] = append(prepareMap[ssUser.HostPort], userArgs)

		coord.nextOperationId++

	case datatypes.MakeTransaction:
		var op datatypes.OperationType
		for i := 0; i < len(data.Requests); i++ {
			// Pull and create args for one request at a time
			req := data.Requests[i]
			if req.Action == "buy" {
				op = datatypes.Buy
			} else if req.Action == "sell" {
				op = datatypes.Sell
			} else {
				return datatypes.NoSuchAction, nil
			}

			// prepare args for individual buy/sell operation
			reqSlice := make([]datatypes.Request, 0)
			reqSlice = append(reqSlice, req)
			reqData := datatypes.DataArgs{
				User:     data.User,
				Requests: reqSlice,
			}

			teamArgs := &storagerpc.PrepareArgs{
				TransactionId: coord.nextOperationId,
				Name:          op,
				Data:          reqData,
			}

			ss := util.FindServerFromKey(data.Team.TeamID, coord.servers)
			prepareMap[ss.HostPort] = append(prepareMap[ss.HostPort], teamArgs)

			coord.nextOperationId++
		}
	default:
		return datatypes.NoSuchAction, nil
	}

	stat, err := coord.Propose(prepareMap)
	return stat, err
}

// Propose receives a map of [hostport] --> [prepareArgs] for that node to execute
// and makes async RPC calls to involved nodes to Prepare for 2PC
func (coord *coordinator) Propose(prepareMap PrepareMap) (datatypes.Status, error) {
	// Prepare for transaction

	stat := storagerpc.CommitStatus(storagerpc.Commit)
	resultStatus := datatypes.OK

	// channel to receive async replies from CohortServers
	doneCh := make(chan *rpc.Call, len(coord.servers))

	// send out Prepare call to all nodes
	responsesToExpect := 0
	for hostport, argsList := range prepareMap {
		if _, ok := coord.connections[hostport]; !ok {
			cli, err := util.TryDial(hostport)
			if err != nil {
				return datatypes.BadData, err
			}

			coord.connections[hostport] = cli
		}

		for i := 0; i < len(argsList); i++ {
			prepareArgs := argsList[i]
			var prepareReply storagerpc.PrepareReply
			coord.connections[hostport].Go("CohortStorageServer.Prepare", prepareArgs, &prepareReply, doneCh)
			responsesToExpect++
		}
	}

	// receive replies from prepare
	for i := 0; i < responsesToExpect; i++ {
		rpcReply := <-doneCh

		// if RPC fails or non-OK status then Rollback
		replyStatus := rpcReply.Reply.(*storagerpc.PrepareReply).Status
		if rpcReply.Error != nil || replyStatus != datatypes.OK {
			resultStatus = replyStatus
			stat = storagerpc.Rollback
		}
	}

	// send the Commit call to all nodes with the updated status
	for hostport, args := range prepareMap {
		for i := 0; i < len(prepareMap[hostport]); i++ {
			commitArgs := &storagerpc.CommitArgs{
				TransactionId: args[i].TransactionId,
				Status:        stat,
			}

			var commitReply *storagerpc.CommitReply
			coord.connections[hostport].Go("CohortStorageServer.Commit", commitArgs, &commitReply, doneCh)
		}
	}

	// receive Ack from all nodes
	for i := 0; i < responsesToExpect; i++ {
		rpcReply := <-doneCh

		// TODO: if RPC fails then retry sending message until all received (?)
		if rpcReply.Error != nil {
			return 0, rpcReply.Error
		}
	}
	return datatypes.Status(resultStatus), nil
}
