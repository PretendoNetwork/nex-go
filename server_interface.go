package nex

import "github.com/PretendoNetwork/nex-go/types"

// ServerInterface defines all the methods a server should have regardless of type
type ServerInterface interface {
	AccessKey() string
	SetAccessKey(accessKey string)
	LibraryVersion() *LibraryVersion
	DataStoreProtocolVersion() *LibraryVersion
	MatchMakingProtocolVersion() *LibraryVersion
	RankingProtocolVersion() *LibraryVersion
	Ranking2ProtocolVersion() *LibraryVersion
	MessagingProtocolVersion() *LibraryVersion
	UtilityProtocolVersion() *LibraryVersion
	NATTraversalProtocolVersion() *LibraryVersion
	SetDefaultLibraryVersion(version *LibraryVersion)
	Send(packet PacketInterface)
	PasswordFromPID(pid *types.PID) (string, uint32)
	SetPasswordFromPIDFunction(handler func(pid *types.PID) (string, uint32))
	ByteStreamSettings() *ByteStreamSettings
	SetByteStreamSettings(settings *ByteStreamSettings)
}
