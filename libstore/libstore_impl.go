package libstore

import (
	"achadha235/p3/coordinator"
	"achadha235/p3/rpc/stockrpc"
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"net/rpc"
	"time"
)

type libstore struct {
	client               *rpc.Client            // MasterStorageServer connection
	connections          map[string]*rpc.Client // hostport --> RPC connection to host
	coord                coordinator
	storageServers       []storagerpc.Node
	masterServerHostPort string // host+port of MasterStorageServer
}

func (ls *libstore) findServerFromKey(key string) *storagerpc.Node {
	if ls.storageServers == nil || len(ls.storageServers) == 0 {
		return nil
	}

	hashed := StoreHash(strings.Split(key, ":")[0])

	current := ls.storageServers[0]
	for i := 0; i < len(ls.storageServers); i++ {
		current := ls.storageServers[i]
		if hashed <= current.NodeID {
			return &current
		}
	}
	// current should be last
	if hashed > current.NodeID {
		return &ls.storageServers[0]
	}

	// not found
	return nil
}

func NewLibstore(masterServerHostPort, myHostPort string) (Libstore, error) {
	/* Upon creation, an instance of the Libstore will first contact the coordinator node
	using the GetServers RPC, which will retrieve a list of available storage servers for the
	session use */

	client, err := util.TryDial(masterServerHostPort)
	if err != nil {
		return nil, err
	}

	connectionsMap := make(map[string]*rpc.Client)
	connectionsMap[masterServerHostPort] = client

	ls := &libstore{
		client:               client,
		connections:          connectionsMap,
		storageServers:       nil,
		masterServerHostPort: masterServerHostPort,
	}

	args := &storagerpc.GetServersArgs{}
	var reply storagerpc.GetServersReply

	// attempt to get the list of servers in the ring from the MasterStorageServers
	for i := 0; i < MaxConnectAttempts; i++ {
		ls.connections[masterServerHostPort].Call("StorageServer.GetServers", args, reply)
		if reply.Status != storagerpc.OK {
			time.Sleep(time.Second)
			continue
		} else {
			ls.storageServers = reply.Servers
			break
		}
	}

	// connect to each of the storageServers when we acquire the list
	for i := 0; i < len(ls.storageServers); i++ {
		hostport := newLibstore.storageServers[i].HostPort
		if hostport != masterServerHostPort { // Dont dial the master twice
			cli, err := tryDial(hostport)
			if err != nil {
				return nil, err
			}

			newLibstore.connections[hostport] = cli
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

func (ls *libstore) Get(key string) (string, error) {
	args := &storagerpc.GetArgs{Key: key}
	var reply storagerpc.GetReply

	ss := ls.findServerFromKey(key)

	err := ls.connections[ss.HostPort].Call("StorageServer.Get", args, &reply)
	if err != nil {
		return "", err
	}

	if reply.Status == storagerpc.OK {
		return reply.Value, nil
	}

	return "", errors.New("RPC Get returned status: " + reply.Status)
}

func (ls *libstore) Put(key, value string) error {
	args := &storagerpc.PutArgs{Key: key, Value: value}
	var reply storagerpc.PutReply

	ss := ls.findServerFromKey(key)

	err := ls.client.Call("StorageServer.Put", args, &reply)
	if err != nil {
		ls.connections[ls.port] = nil
		ls.connectToRandom()
		return err
	}

	if reply.Status == storagerpc.OK {
		return nil
	}

	return errors.New("RPC Put returned status: " + reply.Status)
}

/* Transact performs basic checks such as whether a user/team
   exists, followed by a call to coord.PerformTransaction,
	 which will use RPC to attempt the transaction */
func (ls *libstore) Transact(name storagerpc.TransactionType, data string) (storagerpc.TransactionStatus, error) {
	switch name {
	case storagerpc.CreateUser:
		var user stockrpc.User
		err := json.Unmarshal(data, &user)
		if err != nil {
			return 0, err
		}

		// check if user exists
		_, err = ls.Get(user.userID)
		if err != nil {
			return stockrpc.Exists, nil
		}

		return stockrpc.OK, nil

	case storagerpc.CreateTeam:
		var team stockrpc.Team
		err := json.Unmarshal(data, &team)
		if err != nil {
			return 0, err
		}

		// check if team exists
		_, err = ls.Get(team.teamID)
		if err != nil {
			return stockrpc.Exists, nil
		}

	case storagerpc.JoinTeam:
	case storagerpc.LeaveTeam:
	case storagerpc.MakeTransaction:
	}
}
