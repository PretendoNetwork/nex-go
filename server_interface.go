package nex

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
	OnData(handler func(packet PacketInterface))
	PasswordFromPID(pid *PID) (string, uint32)
	SetPasswordFromPIDFunction(handler func(pid *PID) (string, uint32))
	StringLengthSize() int
	SetStringLengthSize(size int)
}
