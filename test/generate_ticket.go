package main

import (
	"crypto/rand"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-go/types"
)

func generateTicket(userPID *types.PID, targetPID *types.PID) []byte {
	userKey := nex.DeriveKerberosKey(userPID, []byte("z5sykuHnX0q5SCJN"))
	targetKey := nex.DeriveKerberosKey(targetPID, []byte("password"))
	sessionKey := make([]byte, authServer.KerberosKeySize())

	_, err := rand.Read(sessionKey)
	if err != nil {
		panic(err)
	}

	ticketInternalData := nex.NewKerberosTicketInternalData()
	serverTime := types.NewDateTime(0).Now()

	ticketInternalData.Issued = serverTime
	ticketInternalData.SourcePID = userPID
	ticketInternalData.SessionKey = sessionKey

	encryptedTicketInternalData, _ := ticketInternalData.Encrypt(targetKey, nex.NewByteStreamOut(authServer))

	ticket := nex.NewKerberosTicket()
	ticket.SessionKey = sessionKey
	ticket.TargetPID = targetPID
	ticket.InternalData = types.NewBuffer(encryptedTicketInternalData)

	encryptedTicket, _ := ticket.Encrypt(userKey, nex.NewByteStreamOut(authServer))

	return encryptedTicket
}
