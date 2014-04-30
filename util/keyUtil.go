package util


// This file contains utility functions for creating lookup keys for the storage server.
// Each function is preceded by a comment specifying the structure of the result that is returned

// pre:post
func CreateKey(pre, post string) string { return pre + ":" + post }

// path/route
func AppendRoute(path, route string) string { return path + "/" + route }

// user:id --> bool (true if exists)
func CreateUserKey(id string) string { return CreateKey("user", id) }

// user:id/pw --> hash(pw)
func CreateUserPassKey(id string) string { return AppendRoute(CreateUserKey(id), "pw") }

// team:id --> list of members of team
func CreateTeamKey(id string) string { return CreateKey("team", id) }

// team:id/pw --> hash(pw)
func CreateTeamPassKey(id string) string { return AppendRoute(CreateTeamKey(id), "pw") }

// team:id/balance --> team balance (int64) in cents
func CreateTeamBalanceKey(id string) string { return AppendRoute(CreateTeamKey(id), "balance") }

// team:id/ticker --> number of shares of ticker held by team
func CreateTeamHoldingKey(teamID, ticker string) string {
	return AppendRoute(CreateTeamKey(teamID), ticker)
}

// team:id/holdings --> []holdings
func CreateTeamHoldingsKey(id string) string { return AppendRoute(CreateTeamKey(id), "holdings") }

// ticker:id --> share price (int64) in cents
func CreateTickerKey(id string) string { return CreateKey("ticker", id) }



