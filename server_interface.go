package nex

// ServerInterface defines all the methods a server should have regardless of type
type ServerInterface interface {
	AccessKey() string
	SetAccessKey(accessKey string)
	SetProtocolMinorVersion(protocolMinorVersion uint32)
	ProtocolMinorVersion() uint32
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
	OnReliableData(handler func(PacketInterface))
}
