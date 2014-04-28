package storageserver

type StorageServer interface {

	// RegisterServer adds a storage server to the ring. It replies with
	// status NotReady if not all nodes in the ring have joined. Once
	// all nodes have joined, it should reply with status OK and a list
	// of all connected nodes in the ring.
	RegisterServer(*storagerpc.RegisterArgs, *storagerpc.RegisterReply) error

	// GetServers retrieves a list of all connected nodes in the ring. It
	// replies with status NotReady if not all nodes in the ring have joined.
	GetServers(*storagerpc.GetServersArgs, *storagerpc.GetServersReply) error

	// Get retrieves the specified key from the data store and replies with
	// the key's value and a lease if one was requested. If the key does not
	// fall within the storage server's range, it should reply with status
	// WrongServer. If the key is not found, it should reply with status
	// KeyNotFound.
	Get(*storagerpc.GetArgs, *storagerpc.GetReply) error

	// Put inserts the specified key/value pair into the data store. If
	// the key does not fall within the storage server's range, it should
	// reply with status WrongServer.
	Put(*storagerpc.PutArgs, *storagerpc.PutReply) error
}
