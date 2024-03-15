package constants

// StationURLFlag is an enum of flags used by the StationURL "type" parameter.
type StationURLFlag uint8

const (
	// StationURLFlagBehindNAT indicates the user is behind NAT
	StationURLFlagBehindNAT StationURLFlag = iota + 1

	// StationURLFlagPublic indicates the station is a public address
	StationURLFlagPublic
)
