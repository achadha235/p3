package coordinator

import (
/*    "achadha/p3/rpc/storage"*/
)

type Coordinator interface {
	RegisterServer(*RegisterServerArgs, *RegisterServerReply) error
	Propose(*ProposeArgs, *ProposeReply) error
}
