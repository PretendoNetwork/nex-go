package nex

import "github.com/PretendoNetwork/nex-go/v2/types"

// Account represents a game server account.
//
// Game server accounts are separate from other accounts, like Uplay, Nintendo Accounts and NNIDs.
// These exist only on the game server. Account passwords are used as part of the servers Kerberos
// authentication. There are also a collection of non-user, special, accounts. These include a
// guest account, an account which represents the authentication server, and one which represents
// the secure server. See https://nintendo-wiki.pretendo.network/docs/nex/kerberos for more information.
type Account struct {
	PID               types.PID // * The PID of the account. PIDs are unique IDs per account. NEX PIDs start at 1800000000 and decrement with each new account.
	Username          string    // * The username for the account. For NEX user accounts this is the same as the accounts PID.
	Password          string    // * The password for the account. For NEX accounts this is always 16 characters long using seemingly any ASCII character.
	RequiresTokenAuth bool      // * If the account requires token authentication. Always false for special accounts or user accounts pre-Switch.
}

// NewAccount returns a new instance of Account.
// This does not register an account, only creates a new
// struct instance.
func NewAccount(pid types.PID, username, password string, requiresTokenAuth bool) *Account {
	return &Account{
		PID:               pid,
		Username:          username,
		Password:          password,
		RequiresTokenAuth: requiresTokenAuth,
	}
}
