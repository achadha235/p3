package main

import (
	"flag"
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/stockclient"
	"log"
	"os"
	"time"
)

type testFunc struct {
	name string
	f    func() bool
}

var (
	hostport   = flag.String("server", "localhost", "StockServer hostport")
	client     stockclient.StockClient
	sessionKey []byte
)

var statusMap = map[datatypes.Status]string{
	datatypes.OK:                   "OK",
	datatypes.NoSuchUser:           "NoSuchUser",
	datatypes.NoSuchTeam:           "NoSuchTeam",
	datatypes.NoSuchTicker:         "NoSuchTicker",
	datatypes.NoSuchAction:         "NoSuchAction",
	datatypes.NoSuchHolding:        "NoSuchHolding",
	datatypes.InsufficientQuantity: "InsufficientQuantity",
	datatypes.Exists:               "Exists",
	datatypes.PermissionDenied:     "PermissionDenied",
	datatypes.BadData:              "BadData",
	datatypes.NoSuchSession:		"NoSuchSession",
	0:                              "Unknown",
}

var LOGE = log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds)

func initStockClient(stockServerHostPort string) error {
	cli, err := stockclient.NewStockClient(stockServerHostPort)
	if err != nil {
		return err
	}

	client = cli
	/*	status, err := client.CreateUser("user", "pass")
		if err != nil {
			LOGE.Println("FAIL: Could not create user. Received status: ", statusMap[status])
			return err
		}

		status, key, err := client.LoginUser("user", "pass")
		if err != nil {
			LOGE.Println("FAIL: Could not login. Received status: ", statusMap[status])
			return err
		}

		sessionKey = key*/

	return nil
}

// Check error and status
func checkErrorStatus(err error, status, expectedStatus datatypes.Status) bool {
	if err != nil {
		LOGE.Println("FAIL: unexpected error returned:", err)
		return true
	}
	if status != expectedStatus {
		LOGE.Printf("FAIL: incorrect status %s, expected status %s\n", statusMap[status], statusMap[expectedStatus])
		return true
	}
	return false
}

func checkPortfolio(holdings, expectedHoldings []datatypes.Holding) bool {
	if len(holdings) != len(expectedHoldings) {
		LOGE.Printf("FAIL: incorrect holdings %v, expected holdings %v\n", holdings, expectedHoldings)
		return true
	}

	for i := 0; i < len(holdings); i++ {
		h := holdings[i]
		eh := expectedHoldings[i]
		if h.Ticker != eh.Ticker {
			LOGE.Printf("FAIL: incorrect holding ticker %v, expected holdings %v\n", holdings, expectedHoldings)
			return true
		}
		if h.Quantity != eh.Quantity {
			LOGE.Printf("FAIL: incorrect holding quantity %v, expected holdings %v\n", holdings, expectedHoldings)
			return true
		}
	}

	return false
}

func testCreateUserValid() bool {
	status, err := client.CreateUser("user1", "pass")
	return checkErrorStatus(err, status, datatypes.OK)
}

func testCreateUserDuplicate() bool {
	client.CreateUser("dupuser", "pw")
	status, err := client.CreateUser("dupuser", "alreadythere")
	return checkErrorStatus(err, status, datatypes.Exists)
}

func testCreateTeamValid() bool {
	client.CreateUser("testcreateuser", "pass")
	_, sessionKey, _ := client.LoginUser("testcreateuser", "pass")
	status, err := client.CreateTeam(sessionKey, "team1", "teampass")
	return checkErrorStatus(err, status, datatypes.OK)
}

func testCreateTeamDuplicate() bool {
	_, sessionKey, _ := client.LoginUser("testcreateuser", "pass")
	client.CreateTeam(sessionKey, "dupteam", "teampass")
	status, err := client.CreateTeam(sessionKey, "dupteam", "alreadythere")
	return checkErrorStatus(err, status, datatypes.Exists)
}

func testLoginUserValid() bool {
	client.CreateUser("loginuser", "pass")
	status, _, err := client.LoginUser("loginuser", "pass")
	return checkErrorStatus(err, status, datatypes.OK)
}

func testLoginUserInvalidUser() bool {
	status, _, err := client.LoginUser("nouser", "doesntmatter")
	return checkErrorStatus(err, status, datatypes.NoSuchUser)
}

func testLoginUserInvalidPassword() bool {
	client.CreateUser("loginuser2", "pass")
	status, _, err := client.LoginUser("loginuser2", "incorrect")
	return checkErrorStatus(err, status, datatypes.PermissionDenied)
}

func testJoinTeamValid() bool {
	client.CreateUser("testJoinTeamValidUser", "pass")
	_, sessionKey, _ := client.LoginUser("testJoinTeamValidUser", "pass")
	client.CreateTeam(sessionKey, "testValidTeamJoin", "pw")
	status, err := client.JoinTeam(sessionKey, "testValidTeamJoin", "pw")
	return checkErrorStatus(err, status, datatypes.OK)
}
func testJoinTeamInvalidTeam() bool {
	client.CreateUser("testJoinTeamInvalidTeamUser", "pass")
	_, sessionKey, _ := client.LoginUser("testJoinTeamInvalidTeamUser", "pass")
	status, err := client.JoinTeam(sessionKey, "aninvalidteam", "doesntmatter")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)




}

func testJoinTeamInvalidPassword() bool {
	client.CreateUser("testJoinTeamInvalidPassword", "pass")
	_, sessionKey, _ := client.LoginUser("testJoinTeamInvalidPassword", "pass")
	client.CreateTeam(sessionKey, "testJoinTeamInvalidPassword", "pw")
	status, err := client.JoinTeam(sessionKey, "testJoinTeamInvalidPassword", "incorrectpw")
	return checkErrorStatus(err, status, datatypes.PermissionDenied)
}


func testLeaveTeamInvalidTeam() bool {
	client.CreateUser("testLeaveTeamInvalidTeam", "pass")
	_, sessionKey, _ := client.LoginUser("testLeaveTeamInvalidTeam", "pass")

	status, err := client.LeaveTeam(sessionKey, "aninvalidteam")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)
}

func testLeaveTeamNotOnTeam() bool {
	client.CreateUser("testLeaveTeamNotOnTeamOwner", "hey")
	_, sessionKey, _ := client.LoginUser("testLeaveTeamNotOnTeamOwner", "hey")
	client.CreateTeam(sessionKey, "leaveteam", "bye")
	client.CreateUser("newUser", "hey")


	client.CreateUser("testLeaveTeamNotOnTeamRando", "randomguypass")
	_, randomSesh, _ := client.LoginUser("testLeaveTeamNotOnTeamRando", "randomguypass")
	status, err := client.LeaveTeam(randomSesh, "leaveteam")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)
}

func testLeaveTeamValid() bool {
	client.CreateUser("testLeaveTeamValid", "pass")
	_, sessionKey, _ := client.LoginUser("testLeaveTeamValid", "pass")
	client.CreateTeam(sessionKey, "testLeaveTeamValid", "bye")
	status, err := client.JoinTeam(sessionKey, "testLeaveTeamValid", "bye")
	time.Sleep(time.Second)

	status, err = client.LeaveTeam(sessionKey, "testLeaveTeamValid")

	return checkErrorStatus(err, status, datatypes.OK)
}

func testBuyOneValid() bool {
	client.CreateUser("testBuyOneValid", "pass")
	_, sessionKey, _ := client.LoginUser("testBuyOneValid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testBuyOneValidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testBuyOneValidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyOneValidTeam",
			Ticker: "APPL",
			Quantity:10,
		},
	}
	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.OK)
}

func testBuyOneInvalid() bool {
	client.CreateUser("testBuyOneInvalid", "pass")
	_, sessionKey, _ := client.LoginUser("testBuyOneInvalid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testBuyOneInvalidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testBuyOneInvalidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyOneInvalidTeam",
			Ticker: "APPL",
			Quantity:100000000000,
		},
	}
	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.InsufficientQuantity)
}

func testSellOneValid() bool {
	client.CreateUser("testSellOneValid", "pass")
	_, sessionKey, _ := client.LoginUser("testSellOneValid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testSellOneValidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testSellOneValidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testSellOneValidTeam",
			Ticker: "APPL",
			Quantity:10,
		},
	}
	client.MakeTransaction(sessionKey, reqs)
	
	reqs = []datatypes.Request{ 
		datatypes.Request{
			Action:"sell",
			TeamID:"testSellOneValidTeam",
			Ticker: "APPL",
			Quantity:4,
		},
	}	
	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.OK)
}

func testSellOneInvalid() bool {
	client.CreateUser("testSellOneInvalid", "pass")
	_, sessionKey, _ := client.LoginUser("testSellOneInvalid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testSellOneInvalidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testSellOneInvalidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testSellOneInvalidTeam",
			Ticker: "APPL",
			Quantity:10,
		},
	}
	client.MakeTransaction(sessionKey, reqs)
	reqs = []datatypes.Request{ 
		datatypes.Request{
			Action:"sell",
			TeamID:"testSellOneInvalidTeam",
			Ticker: "APPL",
			Quantity:100,
		},
	}	
	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.InsufficientQuantity)
}


func testBuyMultiValid() bool {
	client.CreateUser("testBuyMultiValid", "pass")
	_, sessionKey, _ := client.LoginUser("testBuyMultiValid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testBuyMultiValidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testBuyMultiValidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyMultiValidTeam",
			Ticker: "APPL",
			Quantity:10,
		},
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyMultiValidTeam",
			Ticker: "POM",
			Quantity:10,
		},

	}
	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.OK)
}

func testSellMultiValid() bool {
	client.CreateUser("testSellMultiValid", "pass")
	_, sessionKey, _ := client.LoginUser("testSellMultiValid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testSellMultiValidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testSellMultiValidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testSellMultiValidTeam",
			Ticker: "APPL",
			Quantity:10,
		},
		datatypes.Request{
			Action:"buy",
			TeamID:"testSellMultiValidTeam",
			Ticker: "POM",
			Quantity:50,
		},
	}

	client.MakeTransaction(sessionKey, reqs)
	reqs = []datatypes.Request{ 
		datatypes.Request{
			Action:"sell",
			TeamID:"testSellMultiValidTeam",
			Ticker: "APPL",
			Quantity:3,
		},
		datatypes.Request{
			Action:"sell",
			TeamID:"testSellMultiValidTeam",
			Ticker: "POM",
			Quantity:3,
		},

	}

	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.OK)
}


func testBuyMultiInvalid() bool {
	client.CreateUser("testBuyMultiInvalid", "pass")
	_, sessionKey, _ := client.LoginUser("testBuyMultiInvalid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testBuyMultiInvalidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testBuyMultiInvalidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyMultiInvalidTeam",
			Ticker: "POM",
			Quantity:10,
		},
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyMultiInvalidTeam",
			Ticker: "DNB",
			Quantity:1,
		},
		datatypes.Request{
			Action:"buy",
			TeamID:"testBuyMultiInvalidTeam",
			Ticker: "APPL",
			Quantity:100000000000,
		},
	}
	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.InsufficientQuantity)
}

func testSellMultiInvalid() bool {
	client.CreateUser("testSellMultiInvalid", "pass")
	_, sessionKey, _ := client.LoginUser("testSellMultiInvalid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testSellMultiInvalidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testSellMultiInvalidTeam", "pw")

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testSellMultiInvalidTeam",
			Ticker: "APPL",
			Quantity:10,
		},
		datatypes.Request{
			Action:"buy",
			TeamID:"testSellMultiInvalidTeam",
			Ticker: "POM",
			Quantity:50,
		},
	}

	client.MakeTransaction(sessionKey, reqs)
	reqs = []datatypes.Request{ 
		datatypes.Request{
			Action:"sell",
			TeamID:"testSellMultiInvalidTeam",
			Ticker: "APPL",
			Quantity:3,
		},
		datatypes.Request{
			Action:"sell",
			TeamID:"testSellMultiInvalidTeam",
			Ticker: "POM",
			Quantity:400,
		},
	}

	status, err := client.MakeTransaction(sessionKey, reqs)
	return checkErrorStatus(err, status, datatypes.InsufficientQuantity)
}


func testGetPortfolioValid() bool {
	client.CreateUser("testGetPortfolioValid", "pass")
	_, sessionKey, _ := client.LoginUser("testGetPortfolioValid", "pass")
	status, _ := client.CreateTeam(sessionKey, "testGetPortfolioValidTeam", "pw")
	status, _ = client.JoinTeam(sessionKey, "testGetPortfolioValidTeam", "pw")

	startAmount := uint64(50)
	sellAmount := uint64(10)
	remaining := startAmount - sellAmount

	reqs := []datatypes.Request{ 
		datatypes.Request{
			Action:"buy",
			TeamID:"testGetPortfolioValidTeam",
			Ticker: "APPL",
			Quantity:startAmount,
		},
		datatypes.Request{
			Action:"buy",
			TeamID:"testGetPortfolioValidTeam",
			Ticker: "POM",
			Quantity:startAmount,
		},

	}
	client.MakeTransaction(sessionKey, reqs)

	reqs = []datatypes.Request{ 
		datatypes.Request{
			Action:"sell",
			TeamID:"testGetPortfolioValidTeam",
			Ticker: "APPL",
			Quantity:sellAmount,
		},
		datatypes.Request{
			Action:"sell",
			TeamID:"testGetPortfolioValidTeam",
			Ticker: "POM",
			Quantity:sellAmount,
		},
	}
	// Check if amount is right
	holdings, status, err := client.GetPortfolio("testGetPortfolioValidTeam")
	for i:= 0; i<len(holdings); i++ {
		if holdings[i].Quantity != remaining {
			return false
		}
	}
	return checkErrorStatus(err, status, datatypes.OK)
}

func testGetPortfolioInvalid() bool {
	_, status, err := client.GetPortfolio("testGetPortfolioInvalid")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)
}


func main() {
	tests := []testFunc{
		{"testCreateUserValid", testCreateUserValid},
		{"testCreateUserDuplicate", testCreateUserDuplicate},

		{"testCreateTeamValid", testCreateTeamValid},
		{"testCreateTeamDuplicate", testCreateTeamDuplicate},

		{"testLoginUserValid", testLoginUserValid},
		{"testLoginUserInvalidUser", testLoginUserInvalidUser},
		{"testLoginUserInvalidPassword", testLoginUserInvalidPassword},

		{"testJoinTeamValid", testJoinTeamValid},
		{"testJoinTeamInvalidTeam", testJoinTeamInvalidTeam},
		{"testJoinTeamInvalidPassword", testJoinTeamInvalidPassword},

		{"testLeaveTeamInvalidTeam", testLeaveTeamInvalidTeam},
		{"testLeaveTeamNotOnTeam", testLeaveTeamNotOnTeam},
		{"testLeaveTeamValid", testLeaveTeamValid},
		{"testBuyOneValid", testBuyOneValid},
		{"testBuyOneInvalid", testBuyOneInvalid},
		{"testSellOneValid", testSellOneValid},
		{"testSellOneInvalid", testSellOneInvalid},
		{"testBuyMultiValid", testBuyMultiValid},
		{"testSellMultValid", testSellMultiValid},
		{"testBuyMultiInvalid", testBuyMultiInvalid},
		{"testSellMultiInValid", testSellMultiInvalid},

		{"testGetPortfolioValid", testGetPortfolioValid},
		{"testGetPortfolioInvalid", testGetPortfolioInvalid},
	}

	_ = tests

	flag.Parse()
	if flag.NFlag() < 1 {
		LOGE.Fatalln("Usage: -server=<hostport>")
	}

	if err := initStockClient(*hostport); err != nil {
		LOGE.Fatalln("Failed to connect to StockServer: ", err)
	}

	// Run tests.
	for _, t := range tests {
		log.Println("Starting ", t.name, "...")
		fail := t.f()
		if fail {
			log.Println("Failed ", t.name, ".")
		} else {
			log.Println("Passed ", t.name, "!")

		}


	}

	LOGE.Println("Finished tests")
}
