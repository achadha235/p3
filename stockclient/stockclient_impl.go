// This file contains a StockClient implementation

package stockclient

import (
	"achadha235/p3/rpc/stockrpc"
	"net"
	"net/rpc"
	"strconv"
)

type stockClient struct {
	client *rpc.Client
}

func NewStockClient(host string, port int) (StockClient, error) {
	cli, err := rpc.DialHTTP("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	return &stockClient{client: cli}, nil
}

func (sc *stockClient) LoginUser(userID, password string) (stockrpc.Status, []byte, error) {
	args := &stockrpc.LoginUserArgs{UserID: userID, Password: password}
	var reply stockrpc.LoginUserReply
	if err := sc.client.Call("StockServer.LoginUser", args, &reply); err != nil {
		return 0, make([]byte, 0), err
	}

	return reply.Status, reply.SessionKey, nil
}

func (sc *stockClient) CreateUser(userID, password string) (stockrpc.Status, error) {
	args := &stockrpc.CreateUserArgs{UserID: userID, Password: password}
	var reply stockrpc.CreateUserReply
	if err := sc.client.Call("StockServer.CreateUser", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) CreateTeam(teamID, password string, sessionKey []byte) (stockrpc.Status, error) {
	args := &stockrpc.CreateTeamArgs{UserID: userID, Password: password, SessionKey: sessionKey}
	var reply stockrpc.CreateTeamReply
	if err := sc.client.Call("StockServer.CreateTeam", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) JoinTeam(teamID, password string, sessionKey []byte) (stockrpc.Status, error) {
	args := &stockrpc.JoinTeamArgs{TeamID: teamID, Password: password, SessionKey: sessionKey}
	var reply stockrpc.JoinTeamReply
	if err := sc.client.Call("StockServer.JoinTeam", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) LeaveTeam(teamID string, sessionKey []byte) (stockrpc.Status, error) {
	args := &stockrpc.LeaveTeamArgs{TeamID: teamID, SessionKey: sessionKey}
	var reply stockrpc.LeaveTeamReply
	if err := sc.client.Call("StockServer.LeaveTeam", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) MakeTransaction(action, teamID, ticker string, quantity int, sessionKey []byte) (stockrpc.Status, error) {
	args := &stockrpc.MakeTransactionArgs{Action: action, TeamID: teamID, Ticker: ticker, Quantity: quantity, SessionKey: sessionKey}
	var reply stockrpc.MakeTransactionReply
	if err := sc.client.Call("StockServer.MakeTransaction", args, &reply); err != nil {
		return 0, err
	}
	return reply.Status, nil
}

func (sc *stockClient) GetPortfolio(teamID string) ([]stockrpc.Holding, stockrpc.Status, error) {
	args := &stockrpc.GetPortfolioArgs{TeamID: teamID}
	var reply stockrpc.GetPortfolioReply
	if err := sc.client.Call("StockServer.GetPortfolio", args, &reply); err != nil {
		mt := make([]stockrpc.Holding, 0)
		return mt, 0, err
	}
	return reply.Stocks, reply.Status, nil
}

func (sc *stockClient) GetPrice(ticker string) (int64, stockrpc.Status, error) {
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
