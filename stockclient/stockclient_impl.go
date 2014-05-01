// This file contains a StockClient implementation

package stockclient

import (
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/rpc/stockrpc"
	"net/rpc"
	"log"
)

type stockClient struct {
	client *rpc.Client
}

func NewStockClient(hostport string) (StockClient, error) {
	cli, err := rpc.DialHTTP("tcp", hostport)
	if err != nil {
		return nil, err
	}
	return &stockClient{client: cli}, nil
}

func (sc *stockClient) LoginUser(userID, password string) (datatypes.Status, []byte, error) {
	args := &stockrpc.LoginUserArgs{UserID: userID, Password: password}
	var reply stockrpc.LoginUserReply
	if err := sc.client.Call("StockServer.LoginUser", args, &reply); err != nil {
		return 0, make([]byte, 0), err
	}

	return reply.Status, reply.SessionKey, nil
}

func (sc *stockClient) CreateUser(userID, password string) (datatypes.Status, error) {
	args := &stockrpc.CreateUserArgs{UserID: userID, Password: password}
	var reply stockrpc.CreateUserReply
	if err := sc.client.Call("StockServer.CreateUser", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) CreateTeam(sessionKey []byte, teamID, password string) (datatypes.Status, error) {
	args := &stockrpc.CreateTeamArgs{TeamID: teamID, Password: password, SessionKey: sessionKey}
	var reply stockrpc.CreateTeamReply
	if err := sc.client.Call("StockServer.CreateTeam", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) JoinTeam(sessionKey []byte, teamID, password string) (datatypes.Status, error) {
	log.Println("Joining team")
	defer log.Println("Joining team returning")

	args := &stockrpc.JoinTeamArgs{TeamID: teamID, Password: password, SessionKey: sessionKey}
	var reply stockrpc.JoinTeamReply
	if err := sc.client.Call("StockServer.JoinTeam", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) LeaveTeam(sessionKey []byte, teamID string) (datatypes.Status, error) {

	args := &stockrpc.LeaveTeamArgs{TeamID: teamID, SessionKey: sessionKey}
	var reply stockrpc.LeaveTeamReply
	if err := sc.client.Call("StockServer.LeaveTeam", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) MakeTransaction(sessionKey []byte, requests []datatypes.Request) (datatypes.Status, error) {
	args := &stockrpc.MakeTransactionArgs{Requests: requests, SessionKey: sessionKey}
	var reply stockrpc.MakeTransactionReply
	if err := sc.client.Call("StockServer.MakeTransaction", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) GetPortfolio(teamID string) ([]datatypes.Holding, datatypes.Status, error) {
	args := &stockrpc.GetPortfolioArgs{TeamID: teamID}
	var reply stockrpc.GetPortfolioReply
	if err := sc.client.Call("StockServer.GetPortfolio", args, &reply); err != nil {
		mt := make([]datatypes.Holding, 0)
		return mt, 0, err
	}
	return reply.Stocks, reply.Status, nil
}

func (sc *stockClient) GetPrice(ticker string) (uint64, datatypes.Status, error) {
	args := &stockrpc.GetPriceArgs{Ticker: ticker}
	var reply stockrpc.GetPriceReply
	if err := sc.client.Call("StockServer.GetPrice", args, &reply); err != nil {
		return 0, 0, err
	}
	return reply.Price, reply.Status, nil
}

func (sc *stockClient) Close() error {
	return sc.client.Close()
}
