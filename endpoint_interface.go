package nex

// EndpointInterface defines all the methods an endpoint should have regardless of type
type EndpointInterface interface {
	AccessKey() string
	SetAccessKey(accessKey string)
	Send(packet PacketInterface)
	LibraryVersions() *LibraryVersions
	ByteStreamSettings() *ByteStreamSettings
	SetByteStreamSettings(settings *ByteStreamSettings)
	UseVerboseRMC() bool // TODO - Move this to a RMCSettings struct?
	EnableVerboseRMC(enabled bool)
	EmitError(err *Error)
}
