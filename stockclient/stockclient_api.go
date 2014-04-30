// This file contains the API for the StockClient

package stockclient

import "achadha235/p3/datatypes"

// StockClient defines the interface for the implementation
type StockClient interface {
	LoginUser(userID, password string) (datatypes.Status, []byte, error)
	CreateUser(userID, password string) (datatypes.Status, error)
	CreateTeam(sessionKey []byte, teamID, password string) (datatypes.Status, error)
	JoinTeam(sessionKey []byte, teamID, password string) (datatypes.Status, error)
	LeaveTeam(sessionKey []byte, teamID string) (datatypes.Status, error)
	MakeTransaction(sessionKey []byte, requests []datatypes.Request) (datatypes.Status, error)
	GetPortfolio(teamID string) ([]datatypes.Holding, datatypes.Status, error)
	GetPrice(ticker string) (uint64, datatypes.Status, error)
	Close() error
}
