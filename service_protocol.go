package nex

// ServiceProtocol represents a NEX service capable of handling PRUDP/HPP packets
type ServiceProtocol interface {
	HandlePacket(packet PacketInterface)
	Endpoint() EndpointInterface
	SetEndpoint(endpoint EndpointInterface)
}
