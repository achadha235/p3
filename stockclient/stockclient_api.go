// This file contains the API for the StockClient

package stockclient

import "achadha235/p3/rpc/stockrpc"

// StockClient defines the interface for the implementation
type StockClient struct {
  CreateUser(userID, password string) (stockrpc.Status, error)
  CreateTeam(teamID, password string) (stockrpc.Status, error)
  JoinTeam(teamID, password string) (stockrpc.Status, error)
  LeaveTeam(teamID string) (stockrpc.Status, error)
  MakeTransaction(action, teamID, ticker) (stockrpc.Status, error)
  GetPortfolio(teamID string) ([]stockrpc.Holding, stockrpc.Status, error)
  GetPrice(ticker string) (int64, stockrpc.Status, error)
  Close() error
}