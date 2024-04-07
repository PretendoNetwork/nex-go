package constants

// StationURLType is an implementation of the nn::nex::StationURL::URLType enum.
//
// StationURLType is used to indicate the type of connection to use when contacting a station.
type StationURLType uint8

const (
	// UnknownStationURLType indicates an unknown URL type
	UnknownStationURLType StationURLType = iota

	// StationURLPRUDP indicates the station should be contacted with a standard PRUDP connection
	StationURLPRUDP

	// StationURLPRUDPS indicates the station should be contacted with a secure PRUDP connection
	StationURLPRUDPS

	// StationURLUDP indicates the station should be contacted with raw UDP data. Used for custom protocols
	StationURLUDP
)
