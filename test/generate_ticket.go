package main

import (
	"crypto/rand"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
)

func generateTicket(source *nex.Account, target *nex.Account, sessionKeyLength int) []byte {
	sourceKey := nex.DeriveKerberosKey(source.PID, []byte(source.Password))
	targetKey := nex.DeriveKerberosKey(target.PID, []byte(target.Password))
	sessionKey := make([]byte, sessionKeyLength)

	_, err := rand.Read(sessionKey)
	if err != nil {
		panic(err)
	}

	ticketInternalData := nex.NewKerberosTicketInternalData(authServer)
	serverTime := types.NewDateTime(0).Now()

	ticketInternalData.Issued = serverTime
	ticketInternalData.SourcePID = source.PID
	ticketInternalData.SessionKey = sessionKey

	encryptedTicketInternalData, _ := ticketInternalData.Encrypt(targetKey, nex.NewByteStreamOut(authServer.LibraryVersions, authServer.ByteStreamSettings))

	ticket := nex.NewKerberosTicket()
	ticket.SessionKey = sessionKey
	ticket.TargetPID = target.PID
	ticket.InternalData = types.NewBuffer(encryptedTicketInternalData)

	encryptedTicket, _ := ticket.Encrypt(sourceKey, nex.NewByteStreamOut(authServer.LibraryVersions, authServer.ByteStreamSettings))

	return encryptedTicket
}
