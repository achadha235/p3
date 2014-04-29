package libstore

import (
	"achadha235/p3/coordinator"
	"achadha235/p3/datatypes"
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"net/rpc"
	"time"
)

type libstore struct {
	client               *rpc.Client            // MasterStorageServer connection
	connections          map[string]*rpc.Client // hostport --> RPC connection to host
	coord                coordinator.Coordinator
	storageServers       []storagerpc.Node
	masterServerHostPort string // host+port of MasterStorageServer
}

func NewLibstore(masterServerHostPort, myHostPort string) (Libstore, error) {
	/* Upon creation, an instance of the Libstore will first contact the coordinator node
	using the GetServers RPC, which will retrieve a list of available storage servers for the
	session use */

	client, err := util.TryDial(masterServerHostPort)
	if err != nil {
		return nil, err
	}

	// launch the coordinator
	coord, err := coordinator.StartCoordinator(masterServerHostPort)
	if err != nil {
		return nil, err
	}

	connectionsMap := make(map[string]*rpc.Client)
	connectionsMap[masterServerHostPort] = client

	ls := &libstore{
		client:               client,
		connections:          connectionsMap,
		coord:                coord,
		storageServers:       nil,
		masterServerHostPort: masterServerHostPort,
	}

	args := &storagerpc.GetServersArgs{}
	var reply storagerpc.GetServersReply

	// attempt to get the list of servers in the ring from the MasterStorageServers
	for i := 0; i < util.MaxConnectAttempts; i++ {
		ls.connections[masterServerHostPort].Call("StorageServer.GetServers", args, reply)
		if reply.Status != datatypes.OK {
			time.Sleep(time.Second)
			continue
		} else {
			ls.storageServers = reply.Servers
			break
		}
	}

	// connect to each of the storageServers when we acquire the list
	for i := 0; i < len(ls.storageServers); i++ {
		hostport := ls.storageServers[i].HostPort
		if hostport != masterServerHostPort { // Dont dial the master twice
			cli, err := util.TryDial(hostport)
			if err != nil {
				return nil, err
			}

			ls.connections[hostport] = cli
		}
	}

	return ls, nil
}

/*
Here we want to perform some basic checks outside of 2PC such as
	whether or not the user/team exists, a ticker is valid, etc.

	Other permissions, such as a user being on a team or other
	cases like a balance going invalid should be handled in the 2PC
*/

func (ls *libstore) Get(key string) (string, datatypes.Status, error) {
	args := &storagerpc.GetArgs{Key: key}
	var reply storagerpc.GetReply

	ss := util.FindServerFromKey(key, ls.storageServers)

	err := ls.connections[ss.HostPort].Call("StorageServer.Get", args, &reply)
	if err != nil {
		return "", datatypes.BadData, err
	}

	if reply.Status == datatypes.OK {
		return reply.Value, datatypes.OK, nil
	}

	return "", reply.Status, nil
}

/* Transact performs basic checks such as whether a user/team
   exists, followed by a call to coord.PerformTransaction,
	 which will use RPC to attempt the transaction */
func (ls *libstore) Transact(name datatypes.TransactionType, data *datatypes.DataArgs) (datatypes.Status, error) {
	switch name {
	case datatypes.CreateUser:
		// Check if user exists
		if ls.checkExists(data.User.UserID) {
			return datatypes.Exists, nil
		}

		// Coordinator does rest
		status, err := ls.coord.PerformTransaction(name, *data)
		return status, err

	case datatypes.CreateTeam:
		// Check if team exists
		if ls.checkExists(data.Team.TeamID) {
			return datatypes.Exists, nil
		}

		// Coordinator does rest
		status, err := ls.coord.PerformTransaction(name, *data)
		return status, err

	// Check if user and team exists for both JoinTeam/LeaveTeam
	case datatypes.JoinTeam:
	case datatypes.LeaveTeam:
		if !ls.checkExists(data.User.UserID) {
			return datatypes.NoSuchUser, nil
		} else if !ls.checkExists(data.Team.TeamID) {
			return datatypes.NoSuchTeam, nil
		}

		status, err := ls.coord.PerformTransaction(name, *data)
		return status, err

	case datatypes.MakeTransaction:
		for i := 0; i < len(data.Requests); i++ {
			// non-OK status then leg of transaction is invalid so cancel all
			if stat := ls.checkRequest(data.Requests[i]); stat != datatypes.OK {
				return stat, nil
			}
		}

		status, err := ls.coord.PerformTransaction(name, *data)
		return status, err
	}

	return datatypes.NoSuchAction, nil
}

// return non-OK status if the team or ticker in the request are invalid
func (ls *libstore) checkRequest(req datatypes.Request) datatypes.Status {
	if !ls.checkExists(req.TeamID) {
		return datatypes.NoSuchTeam
	} else if !ls.checkExists(req.Ticker) {
		return datatypes.NoSuchTicker
	}

	return datatypes.OK
}

// return true if the key (id) exists
func (ls *libstore) checkExists(id string) bool {
	_, status, _ := ls.Get(id)
	// if status is OK then the id was found on the node
	return status == datatypes.OK
}
