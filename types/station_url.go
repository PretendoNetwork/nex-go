package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PretendoNetwork/nex-go/v2/constants"
)

// StationURL is an implementation of rdv::StationURL.
//
// Contains location of a station to connect to, with data about how to connect.
type StationURL struct {
	urlType constants.StationURLType
	flags   uint8
	params  map[string]string
}

func (s *StationURL) numberParamValue(name string, bits int) (uint64, bool) {
	valueString, ok := s.ParamValue(name)
	if !ok {
		return 0, false
	}

	value, err := strconv.ParseUint(valueString, 10, bits)
	if err != nil {
		return 0, false
	}

	return value, true
}

func (s *StationURL) uint8ParamValue(name string) (uint8, bool) {
	value, ok := s.numberParamValue(name, 8)
	if !ok {
		return 0, false
	}

	return uint8(value), true
}

func (s *StationURL) uint16ParamValue(name string) (uint16, bool) {
	value, ok := s.numberParamValue(name, 16)
	if !ok {
		return 0, false
	}

	return uint16(value), true
}

func (s *StationURL) uint32ParamValue(name string) (uint32, bool) {
	value, ok := s.numberParamValue(name, 32)
	if !ok {
		return 0, false
	}

	return uint32(value), true
}

func (s *StationURL) uint64ParamValue(name string) (uint64, bool) {
	return s.numberParamValue(name, 64)
}

func (s *StationURL) boolParamValue(name string) bool {
	valueString, ok := s.ParamValue(name)
	if !ok {
		return false
	}

	return valueString == "1"
}

// WriteTo writes the StationURL to the given writable
func (s *StationURL) WriteTo(writable Writable) {
	str := NewString(s.EncodeToString())

	str.WriteTo(writable)
}

// ExtractFrom extracts the StationURL from the given readable
func (s *StationURL) ExtractFrom(readable Readable) error {
	str := NewString("")

	if err := str.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read StationURL. %s", err.Error())
	}

	s.FromString(string(*str))

	return nil
}

// Copy returns a new copied instance of StationURL
func (s *StationURL) Copy() RVType {
	return NewStationURL(s.EncodeToString())
}

// Equals checks if the input is equal in value to the current instance
func (s *StationURL) Equals(o RVType) bool {
	if _, ok := o.(*StationURL); !ok {
		return false
	}

	other := o.(*StationURL)

	if s.urlType != other.urlType {
		return false
	}

	if s.flags != other.flags {
		return false
	}

	if len(s.params) != len(other.params) {
		return false
	}

	for key, value1 := range s.params {
		value2, ok := other.params[key]
		if !ok || value1 != value2 {
			return false
		}
	}

	return true
}

// SetParamValue sets a StationURL parameter
func (s *StationURL) SetParamValue(name, value string) {
	s.params[name] = value
}

// RemoveParam removes a StationURL parameter.
//
// Originally called nn::nex::StationURL::Remove
func (s *StationURL) RemoveParam(name string) {
	delete(s.params, name)
}

// ParamValue returns the value of the requested param.
//
// Returns the string value and a bool indicating if the value existed or not.
//
// Originally called nn::nex::StationURL::GetParamValue
func (s *StationURL) ParamValue(name string) (string, bool) {
	if value, ok := s.params[name]; ok {
		return value, true
	}

	return "", false
}

// SetAddress sets the stations IP address
func (s *StationURL) SetAddress(address string) {
	s.SetParamValue("address", address)
}

// Address gets the stations IP address.
//
// Originally called nn::nex::StationURL::GetAddress
func (s *StationURL) Address() (string, bool) {
	return s.ParamValue("address")
}

// SetPortNumber sets the stations port
func (s *StationURL) SetPortNumber(port uint16) {
	s.SetParamValue("port", strconv.FormatUint(uint64(port), 10))
}

// PortNumber gets the stations port.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetPortNumber
func (s *StationURL) PortNumber() (uint16, bool) {
	return s.uint16ParamValue("port")
}

// SetURLType sets the stations URL scheme type
func (s *StationURL) SetURLType(urlType constants.StationURLType) {
	s.urlType = urlType
}

// URLType returns the stations scheme type
//
// Originally called nn::nex::StationURL::GetURLType
func (s *StationURL) URLType() constants.StationURLType {
	return s.urlType
}

// SetStreamID sets the stations stream ID
//
// See VirtualPort
func (s *StationURL) SetStreamID(streamID uint8) {
	s.SetParamValue("sid", strconv.FormatUint(uint64(streamID), 10))
}

// StreamID gets the stations stream ID.
//
// See VirtualPort.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetStreamID
func (s *StationURL) StreamID() (uint8, bool) {
	return s.uint8ParamValue("sid")
}

// SetStreamType sets the stations stream type
//
// See VirtualPort
func (s *StationURL) SetStreamType(streamType constants.StreamType) {
	s.SetParamValue("stream", strconv.FormatUint(uint64(streamType), 10))
}

// StreamType gets the stations stream type.
//
// See VirtualPort.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetStreamType
func (s *StationURL) StreamType() (constants.StreamType, bool) {
	streamType, ok := s.uint8ParamValue("stream")

	// TODO - Range check on the enum?

	return constants.StreamType(streamType), ok
}

// SetNodeID sets the stations node ID
//
// Originally called nn::nex::StationURL::SetNodeId
func (s *StationURL) SetNodeID(nodeID uint16) {
	s.SetParamValue("NodeID", strconv.FormatUint(uint64(nodeID), 10))
}

// NodeID gets the stations node ID.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetNodeId
func (s *StationURL) NodeID() (uint16, bool) {
	return s.uint16ParamValue("NodeID")
}

// SetPrincipalID sets the stations target PID
func (s *StationURL) SetPrincipalID(pid *PID) {
	s.SetParamValue("PID", strconv.FormatUint(pid.Value(), 10))
}

// PrincipalID gets the stations target PID.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetPrincipalID
func (s *StationURL) PrincipalID() (*PID, bool) {
	pid, ok := s.uint64ParamValue("PID")
	if !ok {
		return nil, false
	}

	return NewPID(pid), true
}

// SetConnectionID sets the stations connection ID
//
// Unsure how this differs from the Rendez-Vous connection ID
func (s *StationURL) SetConnectionID(connectionID uint32) {
	s.SetParamValue("CID", strconv.FormatUint(uint64(connectionID), 10))
}

// ConnectionID gets the stations connection ID.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetConnectionID
func (s *StationURL) ConnectionID() (uint32, bool) {
	return s.uint32ParamValue("CID")
}

// SetRVConnectionID sets the stations Rendez-Vous connection ID
//
// Unsure how this differs from the connection ID
func (s *StationURL) SetRVConnectionID(connectionID uint32) {
	s.SetParamValue("RVCID", strconv.FormatUint(uint64(connectionID), 10))
}

// RVConnectionID gets the stations Rendez-Vous connection ID.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetRVConnectionID
func (s *StationURL) RVConnectionID() (uint32, bool) {
	return s.uint32ParamValue("RVCID")
}

// SetProbeRequestID sets the probe request ID
func (s *StationURL) SetProbeRequestID(probeRequestID uint32) {
	s.SetParamValue("PRID", strconv.FormatUint(uint64(probeRequestID), 10))
}

// ProbeRequestID gets the probe request ID.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetProbeRequestID
func (s *StationURL) ProbeRequestID() (uint32, bool) {
	return s.uint32ParamValue("PRID")
}

// SetFastProbeResponse sets whether fast probe response should be enabled or not
func (s *StationURL) SetFastProbeResponse(fast bool) {
	if fast {
		s.SetParamValue("fastproberesponse", "1")
	} else {
		s.SetParamValue("fastproberesponse", "0")
	}
}

// IsFastProbeResponseEnabled checks if fast probe response is enabled
//
// Originally called nn::nex::StationURL::GetFastProbeResponse
func (s *StationURL) IsFastProbeResponseEnabled() bool {
	return s.boolParamValue("fastproberesponse")
}

// SetNATMapping sets the clients NAT mapping properties
func (s *StationURL) SetNATMapping(mapping constants.NATMappingProperties) {
	s.SetParamValue("natm", strconv.FormatUint(uint64(mapping), 10))
}

// NATMapping gets the clients NAT mapping properties.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetNATMapping
func (s *StationURL) NATMapping() (constants.NATMappingProperties, bool) {
	natm, ok := s.uint8ParamValue("natm")

	// TODO - Range check on the enum?

	return constants.NATMappingProperties(natm), ok
}

// SetNATFiltering sets the clients NAT filtering properties
func (s *StationURL) SetNATFiltering(filtering constants.NATFilteringProperties) {
	s.SetParamValue("natf", strconv.FormatUint(uint64(filtering), 10))
}

// NATFiltering gets the clients NAT filtering properties.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetNATFiltering
func (s *StationURL) NATFiltering() (constants.NATFilteringProperties, bool) {
	natf, ok := s.uint8ParamValue("natf")

	// TODO - Range check on the enum?

	return constants.NATFilteringProperties(natf), ok
}

// SetProbeRequestInitiation sets whether probing should begin or not
func (s *StationURL) SetProbeRequestInitiation(probeinit bool) {
	if probeinit {
		s.SetParamValue("probeinit", "1")
	} else {
		s.SetParamValue("probeinit", "0")
	}
}

// IsProbeRequestInitiationEnabled checks wheteher probing should be initiated.
//
// Originally called nn::nex::StationURL::GetProbeRequestInitiation
func (s *StationURL) IsProbeRequestInitiationEnabled() bool {
	return s.boolParamValue("probeinit")
}

// SetUPnPSupport sets whether UPnP should be enabled or not
func (s *StationURL) SetUPnPSupport(supported bool) {
	if supported {
		s.SetParamValue("upnp", "1")
	} else {
		s.SetParamValue("upnp", "0")
	}
}

// IsUPnPSupported checks whether UPnP is enabled on the station.
//
// Originally called nn::nex::StationURL::GetUPnPSupport
func (s *StationURL) IsUPnPSupported() bool {
	return s.boolParamValue("upnp")
}

// SetNATPMPSupport sets whether PMP should be enabled or not.
//
// Originally called nn::nex::StationURL::SetNatPMPSupport
func (s *StationURL) SetNATPMPSupport(supported bool) {
	if supported {
		s.SetParamValue("pmp", "1")
	} else {
		s.SetParamValue("pmp", "0")
	}
}

// IsNATPMPSupported checks whether PMP is enabled on the station.
//
// Originally called nn::nex::StationURL::GetNatPMPSupport
func (s *StationURL) IsNATPMPSupported() bool {
	return s.boolParamValue("pmp")
}

// SetType sets the stations type flags
func (s *StationURL) SetType(flags uint8) {
	s.flags = flags // * This normally isn't done, but makes IsPublic and IsBehindNAT simpler
	s.SetParamValue("type", strconv.FormatUint(uint64(flags), 10))
}

// Type gets the stations type flags.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetType
func (s *StationURL) Type() (uint8, bool) {
	return s.uint8ParamValue("type")
}

// SetRelayServerAddress sets the address for the relay server
func (s *StationURL) SetRelayServerAddress(address string) {
	s.SetParamValue("Rsa", address)
}

// RelayServerAddress gets the address for the relay server
//
// Originally called nn::nex::StationURL::GetRelayServerAddress
func (s *StationURL) RelayServerAddress() (string, bool) {
	return s.ParamValue("Rsa")
}

// SetRelayServerPort sets the port for the relay server
func (s *StationURL) SetRelayServerPort(port uint16) {
	s.SetParamValue("Rsp", strconv.FormatUint(uint64(port), 10))
}

// RelayServerPort gets the stations relay server port.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetRelayServerPort
func (s *StationURL) RelayServerPort() (uint16, bool) {
	return s.uint16ParamValue("Rsp")
}

// SetRelayAddress gets the address for the relay
func (s *StationURL) SetRelayAddress(address string) {
	s.SetParamValue("Ra", address)
}

// RelayAddress gets the address for the relay
//
// Originally called nn::nex::StationURL::GetRelayAddress
func (s *StationURL) RelayAddress() (string, bool) {
	return s.ParamValue("Ra")
}

// SetRelayPort sets the port for the relay
func (s *StationURL) SetRelayPort(port uint16) {
	s.SetParamValue("Rp", strconv.FormatUint(uint64(port), 10))
}

// RelayPort gets the stations relay port.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetRelayPort
func (s *StationURL) RelayPort() (uint16, bool) {
	return s.uint16ParamValue("Rp")
}

// SetUseRelayServer sets whether or not a relay server should be used
func (s *StationURL) SetUseRelayServer(useRelayServer bool) {
	if useRelayServer {
		s.SetParamValue("R", "1")
	} else {
		s.SetParamValue("R", "0")
	}
}

// IsRelayServerEnabled checks whether the connection should use a relay server.
//
// Originally called nn::nex::StationURL::GetUseRelayServer
func (s *StationURL) IsRelayServerEnabled() bool {
	return s.boolParamValue("R")
}

// SetPlatformType sets the stations platform type
func (s *StationURL) SetPlatformType(platformType uint8) {
	// * This is likely to change based on the target platforms, so no enum
	// * 2 = Wii U (Seen in Minecraft)
	// * 1 = 3DS? Assumed based on Wii U
	s.SetParamValue("Pl", strconv.FormatUint(uint64(platformType), 10))
}

// PlatformType gets the stations target platform. Legal values vary by developer and platforms.
//
// Returns a bool indicating if the parameter existed or not.
//
// Originally called nn::nex::StationURL::GetPlatformType
func (s *StationURL) PlatformType() (uint8, bool) {
	return s.uint8ParamValue("Pl")
}

// IsPublic checks if the station is a public address
func (s *StationURL) IsPublic() bool {
	return s.flags&uint8(constants.StationURLFlagPublic) == uint8(constants.StationURLFlagPublic)
}

// IsBehindNAT checks if the user is behind NAT
func (s *StationURL) IsBehindNAT() bool {
	return s.flags&uint8(constants.StationURLFlagBehindNAT) == uint8(constants.StationURLFlagBehindNAT)
}

// FromString parses the StationURL data from a string
func (s *StationURL) FromString(str string) {
	if str == "" {
		return
	}

	parts := strings.Split(str, ":/")
	parametersString := ""

	// * Unknown schemes seem to be supported based on
	// * Format__Q3_2nn3nex10StationURLFv
	if len(parts) == 1 {
		parametersString = parts[0]
		s.SetURLType(constants.UnknownStationURLType)
	} else if len(parts) == 2 {
		scheme := parts[0]
		parametersString = parts[1]

		if scheme == "prudp" {
			s.SetURLType(constants.StationURLPRUDP)
		} else if scheme == "prudps" {
			s.SetURLType(constants.StationURLPRUDPS)
		} else if scheme == "udp" {
			s.SetURLType(constants.StationURLUDP)
		} else {
			s.SetURLType(constants.UnknownStationURLType)
		}
	} else {
		// * Badly formatted station
		return
	}

	// * Return if there are no fields
	if parametersString == "" {
		return
	}

	parameters := strings.Split(parametersString, ";")

	for i := 0; i < len(parameters); i++ {
		// TODO - StationURL parameters support extra data through the # delimiter. What is that? Need to support it somehow
		name, value, _ := strings.Cut(parameters[i], "=")

		s.SetParamValue(name, value)
	}

	if flags, ok := s.uint8ParamValue("type"); ok {
		s.flags = flags
	}
}

// EncodeToString encodes the StationURL into a string
func (s *StationURL) EncodeToString() string {
	scheme := ""

	// * Unknown schemes seem to be supported based on
	// * Format__Q3_2nn3nex10StationURLFv
	if s.urlType == constants.StationURLPRUDP {
		scheme = "prudp:/"
	} else if s.urlType == constants.StationURLPRUDPS {
		scheme = "prudps:/"
	} else if s.urlType == constants.StationURLUDP {
		scheme = "udp:/"
	}

	fields := make([]string, 0)

	for key, value := range s.params {
		// TODO - StationURL parameters support extra data through the # delimiter. What is that? Need to support it somehow
		fields = append(fields, fmt.Sprintf("%s=%s", key, value))
	}

	return scheme + strings.Join(fields, ";")
}

// String returns a string representation of the struct
func (s *StationURL) String() string {
	return s.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (s *StationURL) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("StationURL{\n")
	b.WriteString(fmt.Sprintf("%surl: %q\n", indentationValues, s.EncodeToString()))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewStationURL returns a new StationURL
func NewStationURL(str string) *StationURL {
	stationURL := &StationURL{
		params: make(map[string]string),
	}

	stationURL.FromString(str)

	return stationURL
}
