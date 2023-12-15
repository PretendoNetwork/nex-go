package nex

import (
	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamOut is an abstraction of github.com/superwhiskers/crunch with nex type support
type StreamOut struct {
	*crunch.Buffer
	Server ServerInterface
}

// WriteUInt8 writes a uint8
func (stream *StreamOut) WriteUInt8(u8 uint8) {
	stream.Grow(1)
	stream.WriteByteNext(byte(u8))
}

// WriteInt8 writes a int8
func (stream *StreamOut) WriteInt8(s8 int8) {
	stream.Grow(1)
	stream.WriteByteNext(byte(s8))
}

// WriteUInt16LE writes a uint16 as LE
func (stream *StreamOut) WriteUInt16LE(u16 uint16) {
	stream.Grow(2)
	stream.WriteU16LENext([]uint16{u16})
}

// WriteUInt16BE writes a uint16 as BE
func (stream *StreamOut) WriteUInt16BE(u16 uint16) {
	stream.Grow(2)
	stream.WriteU16BENext([]uint16{u16})
}

// WriteInt16LE writes a uint16 as LE
func (stream *StreamOut) WriteInt16LE(s16 int16) {
	stream.Grow(2)
	stream.WriteU16LENext([]uint16{uint16(s16)})
}

// WriteInt16BE writes a uint16 as BE
func (stream *StreamOut) WriteInt16BE(s16 int16) {
	stream.Grow(2)
	stream.WriteU16BENext([]uint16{uint16(s16)})
}

// WriteUInt32LE writes a uint32 as LE
func (stream *StreamOut) WriteUInt32LE(u32 uint32) {
	stream.Grow(4)
	stream.WriteU32LENext([]uint32{u32})
}

// WriteUInt32BE writes a uint32 as BE
func (stream *StreamOut) WriteUInt32BE(u32 uint32) {
	stream.Grow(4)
	stream.WriteU32BENext([]uint32{u32})
}

// WriteInt32LE writes a int32 as LE
func (stream *StreamOut) WriteInt32LE(s32 int32) {
	stream.Grow(4)
	stream.WriteU32LENext([]uint32{uint32(s32)})
}

// WriteInt32BE writes a int32 as BE
func (stream *StreamOut) WriteInt32BE(s32 int32) {
	stream.Grow(4)
	stream.WriteU32BENext([]uint32{uint32(s32)})
}

// WriteUInt64LE writes a uint64 as LE
func (stream *StreamOut) WriteUInt64LE(u64 uint64) {
	stream.Grow(8)
	stream.WriteU64LENext([]uint64{u64})
}

// WriteUInt64BE writes a uint64 as BE
func (stream *StreamOut) WriteUInt64BE(u64 uint64) {
	stream.Grow(8)
	stream.WriteU64BENext([]uint64{u64})
}

// WriteInt64LE writes a int64 as LE
func (stream *StreamOut) WriteInt64LE(s64 int64) {
	stream.Grow(8)
	stream.WriteU64LENext([]uint64{uint64(s64)})
}

// WriteInt64BE writes a int64 as BE
func (stream *StreamOut) WriteInt64BE(s64 int64) {
	stream.Grow(8)
	stream.WriteU64BENext([]uint64{uint64(s64)})
}

// WriteFloat32LE writes a float32 as LE
func (stream *StreamOut) WriteFloat32LE(f32 float32) {
	stream.Grow(4)
	stream.WriteF32LENext([]float32{f32})
}

// WriteFloat32BE writes a float32 as BE
func (stream *StreamOut) WriteFloat32BE(f32 float32) {
	stream.Grow(4)
	stream.WriteF32BENext([]float32{f32})
}

// WriteFloat64LE writes a float64 as LE
func (stream *StreamOut) WriteFloat64LE(f64 float64) {
	stream.Grow(8)
	stream.WriteF64LENext([]float64{f64})
}

// WriteFloat64BE writes a float64 as BE
func (stream *StreamOut) WriteFloat64BE(f64 float64) {
	stream.Grow(8)
	stream.WriteF64BENext([]float64{f64})
}

// WriteBool writes a bool
func (stream *StreamOut) WriteBool(b bool) {
	var bVar uint8
	if b {
		bVar = 1
	}
	stream.Grow(1)
	stream.WriteByteNext(byte(bVar))
}

// WritePID writes a NEX PID. The size depends on the server version
func (stream *StreamOut) WritePID(pid *PID) {
	if stream.Server.LibraryVersion().GreaterOrEqual("4.0.0") {
		stream.WriteUInt64LE(pid.pid)
	} else {
		stream.WriteUInt32LE(uint32(pid.pid))
	}
}

// WriteString writes a NEX string type
func (stream *StreamOut) WriteString(str string) {
	str = str + "\x00"
	strLength := len(str)

	if stream.Server == nil {
		stream.WriteUInt16LE(uint16(strLength))
	} else if stream.Server.StringLengthSize() == 4 {
		stream.WriteUInt32LE(uint32(strLength))
	} else {
		stream.WriteUInt16LE(uint16(strLength))
	}

	stream.Grow(int64(strLength))
	stream.WriteBytesNext([]byte(str))
}

// WriteBuffer writes a NEX Buffer type
func (stream *StreamOut) WriteBuffer(data []byte) {
	dataLength := len(data)

	stream.WriteUInt32LE(uint32(dataLength))

	if dataLength > 0 {
		stream.Grow(int64(dataLength))
		stream.WriteBytesNext(data)
	}
}

// WriteQBuffer writes a NEX qBuffer type
func (stream *StreamOut) WriteQBuffer(data []byte) {
	dataLength := len(data)

	stream.WriteUInt16LE(uint16(dataLength))

	if dataLength > 0 {
		stream.Grow(int64(dataLength))
		stream.WriteBytesNext(data)
	}
}

// WriteResult writes a NEX Result type
func (stream *StreamOut) WriteResult(result *Result) {
	stream.WriteUInt32LE(result.Code)
}

// WriteStructure writes a nex Structure type
func (stream *StreamOut) WriteStructure(structure StructureInterface) {
	if structure.ParentType() != nil {
		stream.WriteStructure(structure.ParentType())
	}

	content := structure.Bytes(NewStreamOut(stream.Server))

	useStructures := false

	if stream.Server != nil {
		switch server := stream.Server.(type) {
		case *PRUDPServer: // * Support QRV versions
			useStructures = server.PRUDPMinorVersion >= 3
		default:
			useStructures = server.LibraryVersion().GreaterOrEqual("3.5.0")
		}
	}

	if useStructures {
		stream.WriteUInt8(structure.StructureVersion())
		stream.WriteUInt32LE(uint32(len(content)))
	}

	stream.Grow(int64(len(content)))
	stream.WriteBytesNext(content)
}

// WriteStationURL writes a StationURL type
func (stream *StreamOut) WriteStationURL(stationURL *StationURL) {
	stream.WriteString(stationURL.EncodeToString())
}

// WriteListUInt8 writes a list of uint8 types
func (stream *StreamOut) WriteListUInt8(list []uint8) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt8(list[i])
	}
}

// WriteListInt8 writes a list of int8 types
func (stream *StreamOut) WriteListInt8(list []int8) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt8(list[i])
	}
}

// WriteListUInt16LE writes a list of Little-Endian encoded uint16 types
func (stream *StreamOut) WriteListUInt16LE(list []uint16) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt16LE(list[i])
	}
}

// WriteListUInt16BE writes a list of Big-Endian encoded uint16 types
func (stream *StreamOut) WriteListUInt16BE(list []uint16) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt16BE(list[i])
	}
}

// WriteListInt16LE writes a list of Little-Endian encoded int16 types
func (stream *StreamOut) WriteListInt16LE(list []int16) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt16LE(list[i])
	}
}

// WriteListInt16BE writes a list of Big-Endian encoded int16 types
func (stream *StreamOut) WriteListInt16BE(list []int16) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt16BE(list[i])
	}
}

// WriteListUInt32LE writes a list of Little-Endian encoded uint32 types
func (stream *StreamOut) WriteListUInt32LE(list []uint32) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt32LE(list[i])
	}
}

// WriteListUInt32BE writes a list of Big-Endian encoded uint32 types
func (stream *StreamOut) WriteListUInt32BE(list []uint32) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt32BE(list[i])
	}
}

// WriteListInt32LE writes a list of Little-Endian encoded int32 types
func (stream *StreamOut) WriteListInt32LE(list []int32) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt32LE(list[i])
	}
}

// WriteListInt32BE writes a list of Big-Endian encoded int32 types
func (stream *StreamOut) WriteListInt32BE(list []int32) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt32BE(list[i])
	}
}

// WriteListUInt64LE writes a list of Little-Endian encoded uint64 types
func (stream *StreamOut) WriteListUInt64LE(list []uint64) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt64LE(list[i])
	}
}

// WriteListUInt64BE writes a list of Big-Endian encoded uint64 types
func (stream *StreamOut) WriteListUInt64BE(list []uint64) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteUInt64BE(list[i])
	}
}

// WriteListInt64LE writes a list of Little-Endian encoded int64 types
func (stream *StreamOut) WriteListInt64LE(list []int64) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt64LE(list[i])
	}
}

// WriteListInt64BE writes a list of Big-Endian encoded int64 types
func (stream *StreamOut) WriteListInt64BE(list []int64) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteInt64BE(list[i])
	}
}

// WriteListFloat32LE writes a list of Little-Endian encoded float32 types
func (stream *StreamOut) WriteListFloat32LE(list []float32) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteFloat32LE(list[i])
	}
}

// WriteListFloat32BE writes a list of Big-Endian encoded float32 types
func (stream *StreamOut) WriteListFloat32BE(list []float32) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteFloat32BE(list[i])
	}
}

// WriteListFloat64LE writes a list of Little-Endian encoded float64 types
func (stream *StreamOut) WriteListFloat64LE(list []float64) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteFloat64LE(list[i])
	}
}

// WriteListFloat64BE writes a list of Big-Endian encoded float64 types
func (stream *StreamOut) WriteListFloat64BE(list []float64) {
	stream.WriteUInt32LE(uint32(len(list)))

	for i := 0; i < len(list); i++ {
		stream.WriteFloat64BE(list[i])
	}
}

// WriteListPID writes a list of NEX PIDs
func (stream *StreamOut) WriteListPID(pids []*PID) {
	length := len(pids)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WritePID(pids[i])
	}
}

// WriteListString writes a list of NEX String types
func (stream *StreamOut) WriteListString(strings []string) {
	length := len(strings)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WriteString(strings[i])
	}
}

// WriteListBuffer writes a list of NEX Buffer types
func (stream *StreamOut) WriteListBuffer(buffers [][]byte) {
	length := len(buffers)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WriteBuffer(buffers[i])
	}
}

// WriteListQBuffer writes a list of NEX qBuffer types
func (stream *StreamOut) WriteListQBuffer(buffers [][]byte) {
	length := len(buffers)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WriteQBuffer(buffers[i])
	}
}

// WriteListResult writes a list of NEX Result types
func (stream *StreamOut) WriteListResult(results []*Result) {
	length := len(results)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WriteResult(results[i])
	}
}

// WriteListStationURL writes a list of NEX StationURL types
func (stream *StreamOut) WriteListStationURL(stationURLs []*StationURL) {
	length := len(stationURLs)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WriteString(stationURLs[i].EncodeToString())
	}
}

// WriteListDataHolder writes a NEX DataHolder type
func (stream *StreamOut) WriteListDataHolder(dataholders []*DataHolder) {
	length := len(dataholders)

	stream.WriteUInt32LE(uint32(length))

	for i := 0; i < length; i++ {
		stream.WriteDataHolder(dataholders[i])
	}
}

// WriteDataHolder writes a NEX DataHolder type
func (stream *StreamOut) WriteDataHolder(dataholder *DataHolder) {
	content := dataholder.Bytes(NewStreamOut(stream.Server))
	stream.Grow(int64(len(content)))
	stream.WriteBytesNext(content)
}

// WriteDateTime writes a NEX DateTime type
func (stream *StreamOut) WriteDateTime(datetime *DateTime) {
	stream.WriteUInt64LE(datetime.value)
}

// WriteVariant writes a Variant type
func (stream *StreamOut) WriteVariant(variant *Variant) {
	content := variant.Bytes(NewStreamOut(stream.Server))
	stream.Grow(int64(len(content)))
	stream.WriteBytesNext(content)
}

// NewStreamOut returns a new nex output stream
func NewStreamOut(server ServerInterface) *StreamOut {
	return &StreamOut{
		Buffer: crunch.NewBuffer(),
		Server: server,
	}
}

// StreamWriteListStructure writes a list of structure types to a StreamOut
//
// Implemented as a separate function to utilize generics
func StreamWriteListStructure[T StructureInterface](stream *StreamOut, structures []T) {
	count := len(structures)

	stream.WriteUInt32LE(uint32(count))

	for i := 0; i < count; i++ {
		stream.WriteStructure(structures[i])
	}
}

func mapTypeWriter[T any](stream *StreamOut, t T) {
	// * Map types in NEX can have any type for the
	// * key and value. So we need to just check the
	// * type each time and call the right function
	switch v := any(t).(type) {
	case uint8:
		stream.WriteUInt8(v)
	case int8:
		stream.WriteInt8(v)
	case uint16:
		stream.WriteUInt16LE(v)
	case int16:
		stream.WriteInt16LE(v)
	case uint32:
		stream.WriteUInt32LE(v)
	case int32:
		stream.WriteInt32LE(v)
	case uint64:
		stream.WriteUInt64LE(v)
	case int64:
		stream.WriteInt64LE(v)
	case float32:
		stream.WriteFloat32LE(v)
	case float64:
		stream.WriteFloat64LE(v)
	case string:
		stream.WriteString(v)
	case bool:
		stream.WriteBool(v)
	case []byte:
		// * This actually isn't a good situation, since a byte slice can be either
		// * a Buffer or qBuffer. The only known official case is a qBuffer, inside
		// * UserAccountManagement::LookupSceNpIds, which is why it's implemented
		// * as a qBuffer
		stream.WriteQBuffer(v) // TODO - Maybe we should make Buffer and qBuffer real types?
	case StructureInterface:
		stream.WriteStructure(v)
	case *Variant:
		stream.WriteVariant(v)
	default:
		// * Writer functions don't return errors so just log here.
		// * The client will disconnect but the server won't die,
		// * that way other clients stay connected, but we still
		// * have a log of what the error was
		logger.Warningf("Unsupported Map type trying to be written: %T\n", v)
	}
}

// StreamWriteMap writes a Map type to a StreamOut
//
// Implemented as a separate function to utilize generics
func StreamWriteMap[K comparable, V any](stream *StreamOut, m map[K]V) {
	count := len(m)

	stream.WriteUInt32LE(uint32(count))

	for key, value := range m {
		mapTypeWriter(stream, key)
		mapTypeWriter(stream, value)
	}
}
