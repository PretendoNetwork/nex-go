package constants

// NATMappingProperties is an implementation of the nn::nex::NATProperties::MappingProperties enum.
//
// NATMappingProperties is used to indicate the NAT mapping properties of the users router.
//
// See https://datatracker.ietf.org/doc/html/rfc4787 for more details
type NATMappingProperties uint8

const (
	// UnknownNATMapping indicates the NAT type could not be identified
	UnknownNATMapping NATMappingProperties = iota

	// EIMNATMapping indicates endpoint-independent mapping
	EIMNATMapping

	// EDMNATMapping indicates endpoint-dependent mapping
	EDMNATMapping
)
