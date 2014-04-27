package storagerpc
import (
	"net/rpc"
)

type Status int

const (
	NotReady = iota
	OK
)

type CommitStatus int

const (
	Commit = iota
	Rollback
)

type Node struct {
	NodeId   int
	HostPort string
	Master   bool
	Client   *rpc.Client
}


type LogEntry struct {
	TransactionId int
	Key           string
	Value         string
}

type RegisterServerArgs struct {
}

type RegisterServerReply struct {
	Status int
	NodeId int
}

type ProposeArgs struct {
	Key   string
	Value string
}

type ProposeReply struct {
	Status Status
}

type PrepareArgs struct {
	TransactionId int
	Key           string
	Value         string
}

type PrepareReply struct {
	Status Status
}

type CommitArgs struct {
	TransactionId int
	Status        CommitStatus
}

type CommitReply struct {
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	Key   string
	Value string
}

type PutArgs struct {
	Key   string
	Value string
}

type PutReply struct {
}