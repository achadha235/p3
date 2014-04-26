package libstore

import (
	"achadha235/p3/rpc/storagerpc"
	"achadha235/p3/util"
	"net/rpc"
	"time"
)

type libstore struct {
	client         *rpc.Client // client used for storage communication
	connections    map[string]*rpc.Client
	port           string // port of storageServer assoc with RPC client
	storageServers []storagerpc.Node
	//  port of the StorageServer being used by libstore
	serverHostPort string
}

func NewLibstore(masterServerHostPort, myHostPort string) (Libstore, error) {
	/* Upon creation, an instance of the Libstore will first contact the coordinator node
	using the GetServers RPC, which will retrieve a list of available storage servers for the
	session use */

	connectTries := 0
	maxRetry := 5
	client, err := rpc.DialHTTP("tcp", masterServerHostPort)
	if err != nil {
		return nil, err
	}

	connectionsMap := make(map[string]*rpc.Client)

	ls := &libstore{
		client:         client,
		connections:    connectionsMap,
		storageServers: nil,
		serverHostPort: "", // placeholder
	}

	args := &storagerpc.GetServersArgs{}
	var reply storagerpc.GetServersReply

	ls.connections[masterServerHostPort] = client

	for i := 0; i < maxRetry; i++ {
		ls.connections[masterServerHostPort].Call("StorageServer.GetServers", args, reply)
		if reply.Status != storagerpc.OK {
			connectTries++
			time.Sleep(time.Second)
			continue
		} else {
			ls.storageServers = reply.Servers
			// pick random server to connect to

			err := ls.connectToRandom()
			if err != nil {
				log.Println("Failed to connect to random storageServer")
				return nil, err
			}

			return ls, nil
		}
	}

	return nil, errors.New("Max retry reached without OK status")

}

// selects a random server and connects to it
// updates libstore client and cache connection
func (ls *libstore) connectToRandom() error {
	randInd := util.random(0, len(ls.storageServers)-1)
	hostport := ls.StorageServers[randInd]
	// if connection exists, reuse it
	if _, ok := ls.connections[hostport]; ok {
		ls.client = ls.connections[hostport]
		return nil
	}

	for numTries := 3; numTries >= 0; numTries-- {
		cli, err := rpc.DialHTTP("tcp", hostport)
		if err != nil {
			continue
		}

		ls.port = hostport
		ls.connections[hostport] = cli
		ls.client = cli
		return nil
	}

	return err
}

func (ls *libstore) Get(key string) (string, error) {
	args := &storagerpc.GetArgs{Key: key}
	var reply storagerpc.GetReply

	// might need to retry this call a few times
	// if it fails every time, server is dead
	err := ls.client.Call("StorageServer.Get", args, &reply)
	if err != nil {
		// remove connection and reconnect to a diff server
		ls.connections[ls.port] = nil
		ls.connectToRandom()
		return err
	}

	if reply.Status == storagerpc.OK {
		return reply.Value, nil
	}

	return "", errors.New("RPC Get returned status: " + reply.Status)
}

func (ls *libstore) Put(key, value string) error {
	args := &storagerpc.PutArgs{Key: key, Value: value}
	var reply storagerpc.PutReply

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

func (ls *libstore) GetList(key string) ([]string, error) {
	args := &storagerpc.GetArgs{Key: key}
	var result storagerpc.GetListReply

	err := ls.client.Call("StorageServer.GetList", args, &reply)
	if err != nil {
		ls.connections[ls.port] = nil
		ls.connectToRandom()
		return err
	}

	if reply.Status == storagerpc.OK {
		return reply.Value, nil
	}

	return "", errors.New("RPC GetList returned status: " + reply.Status)
}

func (ls *libstore) AppendToList(key, newItem string) error {
	list, err := ls.GetList(key)
	if err != nil {
		return err
	} else if existsInList(newItem, list) {
		return errors.New("Item exists in list")
	}

	args := &storagerpc.PutArgs{Key: key, Value: newItem}
	var reply storagerpc.PutReply

	err = ls.client.Call("StorageServer.AppendToList", args, &reply)
	if err != nil {
		ls.connections[ls.port] = nil
		ls.connectToRandom()
		return err
	}

	if reply.Status == storagerpc.OK {
		return nil
	}

	return errors.New("RPC AppendToList returned status: " + reply.Status)
}

func (ls *libstore) RemoveFromList(key, removeItem string) error {
	args := &storagerpc.PutArgs{Key: key, Value: removeItem}
	var reply storagerpc.PutReply

	list, err := ls.GetList(key)
	if err != nil {
		return err
	}

	if !existsInList(removeItem, list) {
		return errors.New("Item to remove does not exists in list")
	}

	err = ls.client.Call("StorageServer.RemoveFromList", args, &reply)
	if err != nil {
		ls.connections[ls.port] = nil
		ls.connectToRandom()
		return err
	}

	return nil
}

func existsInList(item string, list []string) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == item {
			return true
		}
	}
	return false
}
