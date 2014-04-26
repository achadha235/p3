// This file contains the API for the StockClient

package stockclient

import "achadha235/p3/rpc/stockrpc"

// StockClient defines the interface for the implementation
type StockClient interface {
	LoginUser(userID, password string) (stockrpc.Status, []byte, error)
	CreateUser(userID, password string) (stockrpc.Status, error)
	CreateTeam(teamID, password string, sessionKey []byte) (stockrpc.Status, error)
	JoinTeam(teamID, password string, sessionKey []byte) (stockrpc.Status, error)
	LeaveTeam(teamID string, sessionKey []byte) (stockrpc.Status, error)
	MakeTransaction(action, teamID, ticker string, quantity int, sessionKey []byte) (stockrpc.Status, error)
	GetPortfolio(teamID string) ([]stockrpc.Holding, stockrpc.Status, error)
	GetPrice(ticker string) (int64, stockrpc.Status, error)
	Close() error
}
