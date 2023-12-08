package nex

const (
	// VirtualStreamTypeDO represents the DO PRUDP virtual connection stream type
	VirtualStreamTypeDO uint8 = 1

	// VirtualStreamTypeRV represents the RV PRUDP virtual connection stream type
	VirtualStreamTypeRV uint8 = 2

	// VirtualStreamTypeOldRVSec represents the OldRVSec PRUDP virtual connection stream type
	VirtualStreamTypeOldRVSec uint8 = 3

	// VirtualStreamTypeSBMGMT represents the SBMGMT PRUDP virtual connection stream type
	VirtualStreamTypeSBMGMT uint8 = 4

	// VirtualStreamTypeNAT represents the NAT PRUDP virtual connection stream type
	VirtualStreamTypeNAT uint8 = 5

	// VirtualStreamTypeSessionDiscovery represents the SessionDiscovery PRUDP virtual connection stream type
	VirtualStreamTypeSessionDiscovery uint8 = 6

	// VirtualStreamTypeNATEcho represents the NATEcho PRUDP virtual connection stream type
	VirtualStreamTypeNATEcho uint8 = 7

	// VirtualStreamTypeRouting represents the Routing PRUDP virtual connection stream type
	VirtualStreamTypeRouting uint8 = 8

	// VirtualStreamTypeGame represents the Game PRUDP virtual connection stream type
	VirtualStreamTypeGame uint8 = 9

	// VirtualStreamTypeRVSecure represents the RVSecure PRUDP virtual connection stream type
	VirtualStreamTypeRVSecure uint8 = 10

	// VirtualStreamTypeRelay represents the Relay PRUDP virtual connection stream type
	VirtualStreamTypeRelay uint8 = 11
)

// PRUDPVirtualStream represents a PRUDP virtual stream
type PRUDPVirtualStream struct {
	clients *MutexMap[string, *PRUDPClient]
}

// PRUDPVirtualPort represents a PRUDP virtual connections virtual port
type PRUDPVirtualPort struct {
	streams *MutexMap[uint8, *PRUDPVirtualStream]
}

func (vp *PRUDPVirtualPort) init() {
	vp.initStream(VirtualStreamTypeDO)
	vp.initStream(VirtualStreamTypeRV)
	vp.initStream(VirtualStreamTypeOldRVSec)
	vp.initStream(VirtualStreamTypeSBMGMT)
	vp.initStream(VirtualStreamTypeNAT)
	vp.initStream(VirtualStreamTypeSessionDiscovery)
	vp.initStream(VirtualStreamTypeNATEcho)
	vp.initStream(VirtualStreamTypeRouting)
	vp.initStream(VirtualStreamTypeGame)
	vp.initStream(VirtualStreamTypeRVSecure)
	vp.initStream(VirtualStreamTypeRelay)
}

func (vp *PRUDPVirtualPort) initStream(streamType uint8) *PRUDPVirtualStream {
	virtualStream := &PRUDPVirtualStream{
		clients: NewMutexMap[string, *PRUDPClient](),
	}

	vp.streams.Set(streamType, virtualStream)

	return virtualStream
}

// PRUDPVirtualConnectionManager manages virtual ports used by PRUDP connections
//
// PRUDP uses a single UDP connection to establish multiple "connections" through virtual ports
type PRUDPVirtualConnectionManager struct {
	ports *MutexMap[uint8, *PRUDPVirtualPort]
}

func (vcm *PRUDPVirtualConnectionManager) init(numberOfPorts uint8) {
	for i := 0; i < int(numberOfPorts); i++ {
		vcm.createVirtualPort(uint8(i))
	}
}

func (vcm *PRUDPVirtualConnectionManager) createVirtualPort(port uint8) *PRUDPVirtualPort {
	virtualPort := &PRUDPVirtualPort{
		streams: NewMutexMap[uint8, *PRUDPVirtualStream](),
	}

	virtualPort.init()

	vcm.ports.Set(port, virtualPort)

	return virtualPort
}

// Get returns PRUDPVirtualStream for the given port and stream type.
// If either the virtual port or stream type do not exist, new ones are created
func (vcm *PRUDPVirtualConnectionManager) Get(port, streamType uint8) *PRUDPVirtualStream {
	virtualPort, ok := vcm.ports.Get(port)
	if !ok {
		// * Just force the port to exist
		virtualPort = vcm.createVirtualPort(port)
		logger.Warningf("Invalid virtual port %d trying to be accessed. Creating new one to prevent crash", port)
	}

	virtualStream, ok := virtualPort.streams.Get(streamType)
	if !ok {
		// * Just force the stream to exist
		virtualStream = virtualPort.initStream(streamType)
		logger.Warningf("Invalid virtual stream type %d trying to be accessed. Creating new one to prevent crash", streamType)
	}

	return virtualStream
}

// NewPRUDPVirtualConnectionManager creates a new PRUDPVirtualConnectionManager with the given number of virtual ports
func NewPRUDPVirtualConnectionManager(numberOfPorts uint8) *PRUDPVirtualConnectionManager {
	virtualConnectionManager := &PRUDPVirtualConnectionManager{
		ports: NewMutexMap[uint8, *PRUDPVirtualPort](),
	}

	virtualConnectionManager.init(numberOfPorts)

	return virtualConnectionManager
}
