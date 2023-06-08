package nex

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamIn is an input stream abstraction of github.com/superwhiskers/crunch with nex type support
type StreamIn struct {
	*crunch.Buffer
	Server *Server
}

// ReadBool reads a bool
func (stream *StreamIn) ReadBool() (bool, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 1 {
		return false, errors.New("Not enough data to read bool")
	}

	return stream.ReadByteNext() == 1, nil
}

// ReadUInt8 reads a uint8
func (stream *StreamIn) ReadUInt8() (uint8, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 1 {
		return 0, errors.New("Not enough data to read uint8")
	}

	return uint8(stream.ReadByteNext()), nil
}

// ReadUInt16LE reads a uint16
func (stream *StreamIn) ReadUInt16LE() (uint16, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return stream.ReadU16LENext(1)[0], nil
}

// ReadUInt32LE reads a uint32
func (stream *StreamIn) ReadUInt32LE() (uint32, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return stream.ReadU32LENext(1)[0], nil
}

// ReadUInt32BE reads a uint32
func (stream *StreamIn) ReadUInt32BE() (uint32, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return stream.ReadU32BENext(1)[0], nil
}

// ReadInt32LE reads a int32
func (stream *StreamIn) ReadInt32LE() (int32, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(stream.ReadU32LENext(1)[0]), nil
}

// ReadUInt64LE reads a uint64
func (stream *StreamIn) ReadUInt64LE() (uint64, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return stream.ReadU64LENext(1)[0], nil
}

// ReadInt64LE reads a int64
func (stream *StreamIn) ReadInt64LE() (int64, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(stream.ReadU64LENext(1)[0]), nil
}

// ReadUInt64BE reads a uint64
func (stream *StreamIn) ReadUInt64BE() (uint64, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return stream.ReadU64BENext(1)[0], nil
}

// ReadFloat64LE reads a int64
func (stream *StreamIn) ReadFloat64LE() (float64, error) {
	if len(stream.Bytes()[stream.ByteOffset():]) < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return stream.ReadF64LENext(1)[0], nil
}

// ReadString reads and returns a nex string type
func (stream *StreamIn) ReadString() (string, error) {
	length, err := stream.ReadUInt16LE()
	if err != nil {
		return "", fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return "", errors.New("NEX string length longer than data size")
	}

	stringData := stream.ReadBytesNext(int64(length))
	str := string(stringData)

	return strings.TrimRight(str, "\x00"), nil
}

// ReadBuffer reads a nex Buffer type
func (stream *StreamIn) ReadBuffer() ([]byte, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to read NEX buffer length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return []byte{}, errors.New("NEX buffer length longer than data size")
	}

	data := stream.ReadBytesNext(int64(length))

	return data, nil
}

// ReadQBuffer reads a nex qBuffer type
func (stream *StreamIn) ReadQBuffer() ([]byte, error) {
	length, err := stream.ReadUInt16LE()
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to read NEX qBuffer length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return []byte{}, errors.New("NEX qBuffer length longer than data size")
	}

	data := stream.ReadBytesNext(int64(length))

	return data, nil
}

// ReadStructure reads a nex Structure type
func (stream *StreamIn) ReadStructure(structure StructureInterface) (StructureInterface, error) {
	if structure.ParentType() != nil {
		_, err := stream.ReadStructure(structure.ParentType())
		if err != nil {
			return nil, fmt.Errorf("Failed to read structure parent. %s", err.Error())
		}
	}

	nexVersion := stream.Server.NEXVersion()

	if nexVersion.Major >= 3 && nexVersion.Minor >= 5 {
		version, err := stream.ReadUInt8()
		if err != nil {
			return nil, fmt.Errorf("Failed to read NEX Structure version. %s", err.Error())
		}

		structureLength, err := stream.ReadUInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read NEX Structure content length. %s", err.Error())
		}

		if len(stream.Bytes()[stream.ByteOffset():]) < int(structureLength) {
			return nil, errors.New("NEX Structure content length longer than data size")
		}

		structure.SetStructureVersion(version)
	}

	err := structure.ExtractFromStream(stream)
	if err != nil {
		return nil, fmt.Errorf("Failed to read structure from stream. %s", err.Error())
	}

	return structure, nil
}

// ReadVariant reads a Variant type. This type can hold 7 different types
func (stream *StreamIn) ReadVariant() (*Variant, error) {
	variant := NewVariant()

	err := variant.ExtractFromStream(stream)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Variant. %s", err.Error())
	}

	return variant, nil
}

// ReadMap reads a Map type with the given key and value types
func (stream *StreamIn) ReadMap(keyFunction interface{}, valueFunction interface{}) (map[interface{}]interface{}, error) {
	/*
		TODO: Make this not suck

		Map types can have any type as the key and any type as the value
		Due to strict typing we cannot just pass stream functions as these values and call them
		At the moment this just reads what type you want from the interface{} function type
	*/

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read Map length. %s", err.Error())
	}

	newMap := make(map[interface{}]interface{})

	for i := 0; i < int(length); i++ {
		var key interface{}
		var value interface{}
		var err error

		switch keyFunction.(type) {
		case func() (string, error):
			key, err = stream.ReadString()
		}

		if err != nil {
			return nil, fmt.Errorf("Failed to read Map key. %s", err.Error())
		}

		switch valueFunction.(type) {
		case func() *Variant:
			value, err = stream.ReadVariant()
		}

		if err != nil {
			return nil, fmt.Errorf("Failed to read Map value. %s", err.Error())
		}

		newMap[key] = value
	}

	return newMap, nil
}

// ReadDateTime reads a DateTime type
func (stream *StreamIn) ReadDateTime() (*DateTime, error) {
	value, err := stream.ReadUInt64LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read DateTime value. %s", err.Error())
	}

	return NewDateTime(value), nil
}

// ReadDataHolder reads a DataHolder type
func (stream *StreamIn) ReadDataHolder() (*DataHolder, error) {
	dataHolder := NewDataHolder()
	err := dataHolder.ExtractFromStream(stream)
	if err != nil {
		return nil, fmt.Errorf("Failed to read DateHolder. %s", err.Error())
	}

	return dataHolder, nil
}

// ReadStationURL reads a StationURL type
func (stream *StreamIn) ReadStationURL() (*StationURL, error) {
	stationString, err := stream.ReadString()
	if err != nil {
		return nil, fmt.Errorf("Failed to read StationURL. %s", err.Error())
	}

	return NewStationURL(stationString), nil
}

// ReadListUInt8 reads a list of uint8 types
func (stream *StreamIn) ReadListUInt8() ([]uint8, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint8> length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return nil, errors.New("NEX List<uint8> length longer than data size")
	}

	list := make([]uint8, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt8()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint8> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt16LE reads a list of uint16 types
func (stream *StreamIn) ReadListUInt16LE() ([]uint16, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint16> length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length*2) {
		return nil, errors.New("NEX List<uint16> length longer than data size")
	}

	list := make([]uint16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt16LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt32LE reads a list of uint32 types
func (stream *StreamIn) ReadListUInt32LE() ([]uint32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint32> length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length*4) {
		return nil, errors.New("NEX List<uint32> length longer than data size")
	}

	list := make([]uint32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt32LE reads a list of int32 types
func (stream *StreamIn) ReadListInt32LE() ([]int32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint32> length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length*4) {
		return nil, errors.New("NEX List<uint32> length longer than data size")
	}

	list := make([]int32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt64LE reads a list of uint64 types
func (stream *StreamIn) ReadListUInt64LE() ([]uint64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint64> length. %s", err.Error())
	}

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length*8) {
		return nil, errors.New("NEX List<uint64> length longer than data size")
	}

	list := make([]uint64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt64LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListString reads a list of NEX String types
func (stream *StreamIn) ReadListString() ([]string, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<String> length. %s", err.Error())
	}

	list := make([]string, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadString()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<String> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListQBuffer reads a list of NEX qBuffer types
func (stream *StreamIn) ReadListQBuffer() ([][]byte, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<qBuffer> length. %s", err.Error())
	}

	list := make([][]byte, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadQBuffer()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<qBuffer> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListStationURL reads a list of NEX Station URL types
func (stream *StreamIn) ReadListStationURL() ([]*StationURL, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<StationURL> length. %s", err.Error())
	}

	list := make([]*StationURL, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadStationURL()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<StationURL> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListStructure reads and returns a list structure types
func (stream *StreamIn) ReadListStructure(structure StructureInterface) (interface{}, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<Structure> length. %s", err.Error())
	}

	structureType := reflect.TypeOf(structure)
	structureSlice := reflect.MakeSlice(reflect.SliceOf(structureType), 0, int(length))

	for i := 0; i < int(length); i++ {
		newStructure := structure.Copy()

		extractedStructure, err := stream.ReadStructure(newStructure)
		if err != nil {
			return nil, err
		}

		structureSlice = reflect.Append(structureSlice, reflect.ValueOf(extractedStructure))
	}

	return structureSlice.Interface(), nil
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server *Server) *StreamIn {
	return &StreamIn{
		Buffer: crunch.NewBuffer(data),
		Server: server,
	}
}
