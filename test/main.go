package main

import (
	"sync"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-go/types"
)

var wg sync.WaitGroup

var authenticationServerAccount *nex.Account
var secureServerAccount *nex.Account
var testUserAccount *nex.Account

func accountDetailsByPID(pid *types.PID) (*nex.Account, uint32) {
	if pid.Equals(authenticationServerAccount.PID) {
		return authenticationServerAccount, 0
	}

	if pid.Equals(secureServerAccount.PID) {
		return secureServerAccount, 0
	}

	if pid.Equals(testUserAccount.PID) {
		return testUserAccount, 0
	}

	return nil, nex.ResultCodes.RendezVous.InvalidPID
}

func accountDetailsByUsername(username string) (*nex.Account, uint32) {
	if username == authenticationServerAccount.Username {
		return authenticationServerAccount, 0
	}

	if username == secureServerAccount.Username {
		return secureServerAccount, 0
	}

	if username == testUserAccount.Username {
		return testUserAccount, 0
	}

	return nil, nex.ResultCodes.RendezVous.InvalidUsername
}

func main() {
	authenticationServerAccount = nex.NewAccount(types.NewPID(1), "Quazal Authentication", "authpassword")
	secureServerAccount = nex.NewAccount(types.NewPID(2), "Quazal Rendez-Vous", "securepassword")
	testUserAccount = nex.NewAccount(types.NewPID(1800000000), "1800000000", "nexuserpassword")

	wg.Add(3)

	go startAuthenticationServer()
	go startSecureServer()
	go startHPPServer()

	wg.Wait()
}
