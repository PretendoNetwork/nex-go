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
}
