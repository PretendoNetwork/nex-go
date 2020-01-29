package nex

import (
	"strings"

	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamIn is an input stream extension of github.com/superwhiskers/crunch with NEX type support
type StreamIn struct {
	*crunch.Buffer
	server *Server
}

// ReadStringNext reads and returns a NEX string type
func (stream *StreamIn) ReadStringNext() string {
	length := stream.ReadU16LENext(1)[0]
	stringData := stream.ReadBytesNext(int64(length))
	str := string(stringData)

	return strings.TrimRight(str, "\x00")
}

// ReadBufferNext reads a NEX Buffer type
func (stream *StreamIn) ReadBufferNext() []byte {
	length := stream.ReadU32LENext(1)[0]
	data := stream.ReadBytesNext(int64(length))

	return data
}

// ReadQBufferNext reads a NEX qBuffer type
func (stream *StreamIn) ReadQBufferNext() []byte {
	length := stream.ReadU16LENext(1)[0]
	data := stream.ReadBytesNext(int64(length))

	return data
}

func (stream *StreamIn) ReadStructureNext(structure StructureInterface) StructureInterface {
	hierarchy := structure.GetHierarchy()

	for _, class := range hierarchy {
		stream.ReadStructureNext(class)
	}

	if stream.server.GetNexMinorVersion() >= 3 {
		// skip the new struct header as we don't really need the data there
		_ = stream.ReadByteNext()   // structure header version
		_ = stream.ReadU32LENext(1) // structure content length
	}

	structure.ExtractFromStream(stream)

	return structure
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server *Server) *StreamIn {
	buff := crunch.NewBuffer(data)

	return &StreamIn{
		Buffer: buff,
		server: server,
	}
}
