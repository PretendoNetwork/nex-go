package main

import (
	"sync"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
)

var wg sync.WaitGroup

var authenticationServerAccount *nex.Account
var secureServerAccount *nex.Account
var testUserAccount *nex.Account

func accountDetailsByPID(pid types.PID) (*nex.Account, *nex.Error) {
	if pid.Equals(authenticationServerAccount.PID) {
		return authenticationServerAccount, nil
	}

	if pid.Equals(secureServerAccount.PID) {
		return secureServerAccount, nil
	}

	if pid.Equals(testUserAccount.PID) {
		return testUserAccount, nil
	}

	return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidPID, "Invalid PID")
}

func accountDetailsByUsername(username string) (*nex.Account, *nex.Error) {
	if username == authenticationServerAccount.Username {
		return authenticationServerAccount, nil
	}

	if username == secureServerAccount.Username {
		return secureServerAccount, nil
	}

	if username == testUserAccount.Username {
		return testUserAccount, nil
	}

	return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidPID, "Invalid username")
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
