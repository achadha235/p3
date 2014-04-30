package main

import (
	"flag"
	"github.com/achadha235/p3/datatypes"
	"github.com/achadha235/p3/stockclient"
	"log"
	"os"
)

type testFunc struct {
	name string
	f    func() bool
}

var (
	hostport   = flag.String("hostport", "localhost:9000", "StockServer hostport")
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
	0:                              "Unknown",
}

var LOGE = log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds)

func initStockClient(stockServerHostPort string) error {
	cli, err := stockclient.NewStockClient(stockServerHostPort)
	if err != nil {
		return err
	}

	client = cli
	status, err := client.CreateUser("user", "pass")
	if err != nil {
		LOGE.Println("FAIL: Could not create user. Received status: ", statusMap[status])
		return err
	}

	status, key, err := client.LoginUser("user", "pass")
	if err != nil {
		LOGE.Println("FAIL: Could not login. Received status: ", statusMap[status])
		return err
	}

	sessionKey = key

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
	status, err := client.CreateTeam(sessionKey, "team1", "teampass")
	return checkErrorStatus(err, status, datatypes.OK)
}

func testCreateTeamDuplicate() bool {
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

func testJoinTeamInvalidTeam() bool {
	status, err := client.JoinTeam(sessionKey, "invalidteam", "doesntmatter")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)
}

func testJoinTeamInvalidPassword() bool {
	client.CreateTeam(sessionKey, "jointeam", "pw")
	status, err := client.JoinTeam(sessionKey, "jointeam", "incorrect")
	return checkErrorStatus(err, status, datatypes.PermissionDenied)
}

func testJoinTeamValid() bool {
	client.CreateTeam(sessionKey, "jointeam2", "pw")
	status, err := client.JoinTeam(sessionKey, "jointeam2", "pw")
	return checkErrorStatus(err, status, datatypes.OK)
}

func testLeaveTeamInvalidTeam() bool {
	status, err := client.LeaveTeam(sessionKey, "invalidteam")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)
}

func testLeaveTeamNotOnTeam() bool {
	client.CreateTeam(sessionKey, "leaveteam", "bye")
	client.CreateUser("newUser", "hey")
	_, key, _ := client.LoginUser("newUser", "hey")
	status, err := client.LeaveTeam(key, "leaveteam")
	return checkErrorStatus(err, status, datatypes.NoSuchUser)
}

func testLeaveTeamValid() bool {
	client.CreateTeam(sessionKey, "myteam", "secret")
	status, err := client.LeaveTeam(sessionKey, "myteam")
	return checkErrorStatus(err, status, datatypes.OK)
}

/*func testBuyOneValid() bool {

}

func testBuyMultValid() bool {

}

func testSellOneValid() bool {

}

func testSellMultValid() bool {

}

func testBuySellMultValid() bool {

}

func testBuyMultWithInvalid() bool {

}

func testSellMultWithInvalid() bool {

}

func testBuySellMultWithInvalidBuy() bool {

}

func testBuySellMultWithInvalidSell() bool {

}

func testDoActionsNotLoggedIn() bool {

}

func testDoActionsInvalidSession() bool {

}

func testGetPortfolioInvalidTeam() bool {
	_, status, err := client.GetPortfolio("nada")
	return checkErrorStatus(err, status, datatypes.NoSuchTeam)
}

func testGetPortfolioValid() bool {
}*/

func main() {
	tests := []testFunc{
		{"testCreateUserValid", testCreateUserValid},
		{"testCreateUserDuplicate", testCreateUserDuplicate},

		{"testCreateTeamValid", testCreateTeamValid},
		{"testCreateTeamDuplicate", testCreateTeamDuplicate},

		{"testLoginUserValid", testLoginUserValid},
		{"testLoginUserInvalidUser", testLoginUserInvalidUser},
		{"testLoginUserInvalidPassword", testLoginUserInvalidPassword},

		{"testJoinTeamInvalidTeam", testJoinTeamInvalidTeam},
		{"testJoinTeamInvalidPassword", testJoinTeamInvalidPassword},
		{"testJoinTeamValid", testJoinTeamValid},

		{"testLeaveTeamInvalidTeam", testLeaveTeamInvalidTeam},
		{"testLeaveTeamNotOnTeam", testLeaveTeamNotOnTeam},
		{"testLeaveTeamValid", testLeaveTeamValid},

		/*		{"testBuyOneValid", testBuyOneValid},
				{"testBuyMultValid", testBuyMultValid},
				{"testSellOneValid", testSellOneValid},
				{"testSellMultValid", testSellMultValid},
				{"testBuySellMultValid", testBuySellMultValid},

				{"testBuyMultWithInvalid", testBuyMultWithInvalid},
				{"testSellMultWithInvalid", testSellMultWithInvalid},
				{"testBuySellMultWithInvalidBuy", testBuySellMultWithInvalidBuy},
				{"testBuySellMultWithInvalidSell", testBuySellMultWithInvalidSell},

				{"testGetPortfolioInvalidTeam", testGetPortfolioInvalidTeam},
				{"testGetPortfolioValid", testGetPortfolioValid},

				{"testDoActionsNotLoggedIn", testDoActionsNotLoggedIn},
				{"testDoActionsInvalidSession", testDoActionsInvalidSession},*/
	}

	_ = tests

	flag.Parse()
	if flag.NArg() < 1 {
		LOGE.Fatalln("Usage: stocktest -host=<hostport>")
	}

	if err := initStockClient(*hostport); err != nil {
		LOGE.Fatalln("Failed to connect to StockServer: ", err)
	}

	// Run tests.
	for _, t := range tests {
		t.f()
	}

	LOGE.Println("Finished tests")
}
