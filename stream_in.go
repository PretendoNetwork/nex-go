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

// Remaining reads a bool
func (stream *StreamIn) Remaining() int {
	return len(stream.Bytes()[stream.ByteOffset():])
}

// ReadBool reads a bool
func (stream *StreamIn) ReadBool() (bool, error) {
	if stream.Remaining() < 1 {
		return false, errors.New("Not enough data to read bool")
	}

	return stream.ReadByteNext() == 1, nil
}

// ReadUInt8 reads a uint8
func (stream *StreamIn) ReadUInt8() (uint8, error) {
	if stream.Remaining() < 1 {
		return 0, errors.New("Not enough data to read uint8")
	}

	return uint8(stream.ReadByteNext()), nil
}

// ReadInt8 reads a uint8
func (stream *StreamIn) ReadInt8() (int8, error) {
	if stream.Remaining() < 1 {
		return 0, errors.New("Not enough data to read int8")
	}

	return int8(stream.ReadByteNext()), nil
}

// ReadUInt16LE reads a Little-Endian encoded uint16
func (stream *StreamIn) ReadUInt16LE() (uint16, error) {
	if stream.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return stream.ReadU16LENext(1)[0], nil
}

// ReadUInt16BE reads a Big-Endian encoded uint16
func (stream *StreamIn) ReadUInt16BE() (uint16, error) {
	if stream.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return stream.ReadU16BENext(1)[0], nil
}

// ReadInt16LE reads a Little-Endian encoded int16
func (stream *StreamIn) ReadInt16LE() (int16, error) {
	if stream.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(stream.ReadU16LENext(1)[0]), nil
}

// ReadInt16BE reads a Big-Endian encoded int16
func (stream *StreamIn) ReadInt16BE() (int16, error) {
	if stream.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(stream.ReadU16BENext(1)[0]), nil
}

// ReadUInt32LE reads a Little-Endian encoded uint32
func (stream *StreamIn) ReadUInt32LE() (uint32, error) {
	if stream.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return stream.ReadU32LENext(1)[0], nil
}

// ReadUInt32BE reads a Big-Endian encoded uint32
func (stream *StreamIn) ReadUInt32BE() (uint32, error) {
	if stream.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return stream.ReadU32BENext(1)[0], nil
}

// ReadInt32LE reads a Little-Endian encoded int32
func (stream *StreamIn) ReadInt32LE() (int32, error) {
	if stream.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(stream.ReadU32LENext(1)[0]), nil
}

// ReadInt32BE reads a Big-Endian encoded int32
func (stream *StreamIn) ReadInt32BE() (int32, error) {
	if stream.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(stream.ReadU32BENext(1)[0]), nil
}

// ReadUInt64LE reads a Little-Endian encoded uint64
func (stream *StreamIn) ReadUInt64LE() (uint64, error) {
	if stream.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return stream.ReadU64LENext(1)[0], nil
}

// ReadUInt64BE reads a Big-Endian encoded uint64
func (stream *StreamIn) ReadUInt64BE() (uint64, error) {
	if stream.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return stream.ReadU64BENext(1)[0], nil
}

// ReadInt64LE reads a Little-Endian encoded int64
func (stream *StreamIn) ReadInt64LE() (int64, error) {
	if stream.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(stream.ReadU64LENext(1)[0]), nil
}

// ReadInt64BE reads a Big-Endian encoded int64
func (stream *StreamIn) ReadInt64BE() (int64, error) {
	if stream.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(stream.ReadU64BENext(1)[0]), nil
}

// ReadFloat32LE reads a Little-Endian encoded float32
func (stream *StreamIn) ReadFloat32LE() (float32, error) {
	if stream.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return stream.ReadF32LENext(1)[0], nil
}

// ReadFloat32BE reads a Big-Endian encoded float32
func (stream *StreamIn) ReadFloat32BE() (float32, error) {
	if stream.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return stream.ReadF32BENext(1)[0], nil
}

// ReadFloat64LE reads a Little-Endian encoded float64
func (stream *StreamIn) ReadFloat64LE() (float64, error) {
	if stream.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return stream.ReadF64LENext(1)[0], nil
}

// ReadFloat64BE reads a Big-Endian encoded float64
func (stream *StreamIn) ReadFloat64BE() (float64, error) {
	if stream.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return stream.ReadF64BENext(1)[0], nil
}

// ReadString reads and returns a nex string type
func (stream *StreamIn) ReadString() (string, error) {
	length, err := stream.ReadUInt16LE()
	if err != nil {
		return "", fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	if stream.Remaining() < int(length) {
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

	if stream.Remaining() < int(length) {
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

	if stream.Remaining() < int(length) {
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

	if nexVersion.GreaterOrEqual("3.5.0") {
		version, err := stream.ReadUInt8()
		if err != nil {
			return nil, fmt.Errorf("Failed to read NEX Structure version. %s", err.Error())
		}

		structureLength, err := stream.ReadUInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read NEX Structure content length. %s", err.Error())
		}

		if stream.Remaining() < int(structureLength) {
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

	typeReader := func(function interface{}) (interface{}, error) {
		var value interface{}
		var err error

		switch function.(type) {
		case func() (string, error):
			value, err = stream.ReadString()
		case func() (*Variant, error):
			value, err = stream.ReadVariant()
		default:
			value = nil
			err = errors.New("Unsupported type in ReadMap")
		}

		return value, err
	}

	newMap := make(map[interface{}]interface{})

	for i := 0; i < int(length); i++ {
		key, err := typeReader(keyFunction)
		if err != nil {
			return nil, fmt.Errorf("Failed to read Map key. %s", err.Error())
		}

		value, err := typeReader(valueFunction)
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

	if stream.Remaining() < int(length) {
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

// ReadListInt8 reads a list of int8 types
func (stream *StreamIn) ReadListInt8() ([]int8, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int8> length. %s", err.Error())
	}

	if stream.Remaining() < int(length) {
		return nil, errors.New("NEX List<int8> length longer than data size")
	}

	list := make([]int8, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt8()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int8> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt16LE reads a list of Little-Endian encoded uint16 types
func (stream *StreamIn) ReadListUInt16LE() ([]uint16, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint16> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*2) {
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

// ReadListUInt16BE reads a list of Big-Endian encoded uint16 types
func (stream *StreamIn) ReadListUInt16BE() ([]uint16, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint16> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<uint16> length longer than data size")
	}

	list := make([]uint16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt16BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt16LE reads a list of Little-Endian encoded int16 types
func (stream *StreamIn) ReadListInt16LE() ([]int16, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int16> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<int16> length longer than data size")
	}

	list := make([]int16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt16LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt16BE reads a list of Big-Endian encoded uint16 types
func (stream *StreamIn) ReadListInt16BE() ([]int16, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int16> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<int16> length longer than data size")
	}

	list := make([]int16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt16BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt32LE reads a list of Little-Endian encoded uint32 types
func (stream *StreamIn) ReadListUInt32LE() ([]uint32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint32> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
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

// ReadListUInt32BE reads a list of Big-Endian encoded uint32 types
func (stream *StreamIn) ReadListUInt32BE() ([]uint32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint32> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<uint32> length longer than data size")
	}

	list := make([]uint32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt32BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt32LE reads a list of Little-Endian encoded int32 types
func (stream *StreamIn) ReadListInt32LE() ([]int32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int32> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<int32> length longer than data size")
	}

	list := make([]int32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt32BE reads a list of Big-Endian encoded int32 types
func (stream *StreamIn) ReadListInt32BE() ([]int32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int32> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<int32> length longer than data size")
	}

	list := make([]int32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt32BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt64LE reads a list of Little-Endian encoded uint64 types
func (stream *StreamIn) ReadListUInt64LE() ([]uint64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint64> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*8) {
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

// ReadListUInt64BE reads a list of Big-Endian encoded uint64 types
func (stream *StreamIn) ReadListUInt64BE() ([]uint64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint64> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<uint64> length longer than data size")
	}

	list := make([]uint64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadUInt64BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt64LE reads a list of Little-Endian encoded int64 types
func (stream *StreamIn) ReadListInt64LE() ([]int64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int64> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<int64> length longer than data size")
	}

	list := make([]int64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt64LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt64BE reads a list of Big-Endian encoded int64 types
func (stream *StreamIn) ReadListInt64BE() ([]int64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int64> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<int64> length longer than data size")
	}

	list := make([]int64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadInt64BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat32LE reads a list of Little-Endian encoded float32 types
func (stream *StreamIn) ReadListFloat32LE() ([]float32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float32> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float32> length longer than data size")
	}

	list := make([]float32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadFloat32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat32BE reads a list of Big-Endian encoded float32 types
func (stream *StreamIn) ReadListFloat32BE() ([]float32, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float32> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float32> length longer than data size")
	}

	list := make([]float32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadFloat32BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat64LE reads a list of Little-Endian encoded float64 types
func (stream *StreamIn) ReadListFloat64LE() ([]float64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float64> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float64> length longer than data size")
	}

	list := make([]float64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadFloat64LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat64BE reads a list of Big-Endian encoded float64 types
func (stream *StreamIn) ReadListFloat64BE() ([]float64, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float64> length. %s", err.Error())
	}

	if stream.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float64> length longer than data size")
	}

	list := make([]float64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadFloat64BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float64> value at index %d. %s", i, err.Error())
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

// ReadListBuffer reads a list of NEX Buffer types
func (stream *StreamIn) ReadListBuffer() ([][]byte, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<Buffer> length. %s", err.Error())
	}

	list := make([][]byte, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadBuffer()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<Buffer> value at index %d. %s", i, err.Error())
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

// ReadListDataHolder reads a list of NEX DataHolder types
func (stream *StreamIn) ReadListDataHolder() ([]*DataHolder, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<DataHolder> length. %s", err.Error())
	}

	list := make([]*DataHolder, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := stream.ReadDataHolder()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<DataHolder> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server *Server) *StreamIn {
	return &StreamIn{
		Buffer: crunch.NewBuffer(data),
		Server: server,
	}
}
