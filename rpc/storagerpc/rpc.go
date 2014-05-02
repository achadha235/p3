package storagerpc

type RemoteCohortStorageServer interface { // This should be RemoteCohortServer
	Prepare(*PrepareArgs, *PrepareReply) error
	Commit(*CommitArgs, *CommitReply) error
	Get(*GetArgs, *GetReply) error
	RegisterServer(*RegisterArgs, *RegisterReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
}

type CohortStorageServer struct {
	RemoteCohortStorageServer
}

func Wrap(s RemoteCohortStorageServer) RemoteCohortStorageServer {
	return &CohortStorageServer{s}
}
