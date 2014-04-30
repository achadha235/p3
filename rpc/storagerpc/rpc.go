package storagerpc

type RemoteCohortServer interface { // This should be RemoteCohortServer
	Prepare(*PrepareArgs, *PrepareReply) error
	Commit(*CommitArgs, *CommitReply) error
	Get(*GetArgs, *GetReply) error
	GetServers(*GetServersArgs, *GetServersReply) error
}

type CohortServer struct {
	RemoteCohortServer
}

func Wrap(s RemoteCohortServer) RemoteCohortServer {
	return &CohortServer{s}
}
