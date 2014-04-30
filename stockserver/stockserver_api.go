package stockserver

import "achadha235/p3/rpc/stockrpc"

type StockServer interface {

	// Login with the specified userID and password
	// Replies with status 'NoSuchUser' if user does not exist
	// Reply with status 'PermissionDenied' if password is incorrect
	// Replies with sessionKey for the user to use to make subsequent requests that session
	LoginUser(args *stockrpc.LoginUserArgs, reply *stockrpc.LoginUserReply) error

	// Create a user with the specified userID and password.
	// Reply with status 'Exists' if userID already exists.
	CreateUser(args *stockrpc.CreateUserArgs, reply *stockrpc.CreateUserReply) error

	// Create a team with the specified teamID and password
	// Reply with status 'Exists' if teamID already exists.
	CreateTeam(args *stockrpc.CreateTeamArgs, reply *stockrpc.CreateTeamReply) error

	// Join a team with the specified teamID and password
	// Reply with status 'PermissionDenied' if password is incorrect
	JoinTeam(args *stockrpc.JoinTeamArgs, reply *stockrpc.JoinTeamReply) error

	// Remove specified userID from specified teamID
	// Reply with status 'NoSuchUser' or 'NoSuchTeam' if either invalid
	// Reply with status 'PermissionDenied' if user does not have valid permission
	LeaveTeam(args *stockrpc.LeaveTeamArgs, reply *stockrpc.LeaveTeamReply) error

	// Make transaction of type action for specified teamID, ticker, and quantity
	// Reply with status 'NoSuchUser', 'NoSuchTeam', or 'NoSuchTicker' if invalid
	// Reply with status 'PermissionDenied' if user is not on specified teamID
	// Reply with status 'InsufficientQuantity' if system cannot complete transaction
	MakeTransaction(args *stockrpc.MakeTransactionArgs, reply *stockrpc.MakeTransactionReply) error

	// Replies with an array of holdings describing the portfolio of the specified teamID
	// Reply with status 'NoSuchTeam' if specified teamID invalid
	// Reply with status 'PermissionDenied' if user is not on specified teamID
	GetPortfolio(args *stockrpc.GetPortfolioArgs, reply *stockrpc.GetPortfolioReply) error

	// Replies with an int64 representing, in cents, the current price of the specified ticker
	// Reply with status 'NoSuchTicker' if specified ticker does not exist.
	GetPrice(args *stockrpc.GetPriceArgs, reply *stockrpc.GetPriceReply) error
}
