// This file provides a type-safe wrapper that should be used to register
// the StockServer to receive RPCs from the StockClient.

package stockrpc

type RemoteStockServer interface {
	CreateUser(args *CreateUserArgs, reply *CreateUserReply) error
	CreateTeam(args *CreateTeamArgs, reply *CreateTeamReply) error
	JoinTeam(args *JoinTeamArgs, reply *JoinTeamReply) error
	LeaveTeam(args *LeaveTeamArgs, reply *LeaveTeamReply) error
	MakeTransaction(args *MakeTransactionArgs, reply *MakeTransactionReply) error
	GetPortfolio(args *GetPortfolioArgs, reply *GetPortfolioReply) error
	GetPrice(args *GetPriceArgs, reply *GetPriceReply) error
}

type StockServer struct {
	// Embed RemoteStockServer interface into the StockServer struct for type-safe use
	RemoteStockServer
}

func Wrap(s RemoteStockServer) RemoteStockServer {
	return &StockServer{s}
}
