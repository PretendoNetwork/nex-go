package constants

// NATFilteringProperties is an implementation of the nn::nex::NATProperties::FilteringProperties enum.
//
// NATFilteringProperties is used to indicate the NAT filtering properties of the users router.
//
// See https://datatracker.ietf.org/doc/html/rfc4787 for more details
type NATFilteringProperties uint8

const (
	// UnknownNATFiltering indicates the NAT type could not be identified
	UnknownNATFiltering NATFilteringProperties = iota

	// PIFNATFiltering indicates port-independent filtering
	PIFNATFiltering

	// PDFNATFiltering indicates port-dependent filtering
	PDFNATFiltering
)
