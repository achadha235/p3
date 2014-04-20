package storagerpc

type Status int
const (
	OK			Status = iota // The RPC was successful
	KeyNotFound 			  // Requested key was not found
	NotAuthorized		      // Not allowed to access the following key
	KeyExits				  // Key already exists
	NotReady				  // Server not ready
)

type Node struct {
	HostPort string
	NodeID	 uint32
	Coordinator uint32
}

type RegisterArgs struct {
	ServerInfo Node
}

type RegisterReply struct {
	Status Status
	Servers []Node
	Coordinator uint32			
}

type GetServersArgs struct {
}

type GetServersReply  struct {
	Status Status
	Servers []Node
	Coordinator uint32
}

type GetArgs struct {
	Key       string
	SessionSecrect string
}

type GetReply struct {
	Status Status
	Value  string
}

type PrepareArgs {
	Status Status
	Value string
}

type PutArgs struct {
	Key   string
	Value string
	SessionSecrect string
}

type PutReply struct {
	Status Status
}
