package nex

import (
	"errors"
	"fmt"
	"strings"

	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamIn is an input stream abstraction of github.com/superwhiskers/crunch/v3 with nex type support
type StreamIn struct {
	*crunch.Buffer
	Server ServerInterface
}

// Remaining returns the amount of data left to be read in the buffer
func (s *StreamIn) Remaining() int {
	return len(s.Bytes()[s.ByteOffset():])
}

// ReadRemaining reads all the data left to be read in the buffer
func (s *StreamIn) ReadRemaining() []byte {
	// TODO - Should we do a bounds check here? Or just allow empty slices?
	return s.ReadBytesNext(int64(s.Remaining()))
}

// ReadUInt8 reads a uint8
func (s *StreamIn) ReadUInt8() (uint8, error) {
	if s.Remaining() < 1 {
		return 0, errors.New("Not enough data to read uint8")
	}

	return uint8(s.ReadByteNext()), nil
}

// ReadInt8 reads a uint8
func (s *StreamIn) ReadInt8() (int8, error) {
	if s.Remaining() < 1 {
		return 0, errors.New("Not enough data to read int8")
	}

	return int8(s.ReadByteNext()), nil
}

// ReadUInt16LE reads a Little-Endian encoded uint16
func (s *StreamIn) ReadUInt16LE() (uint16, error) {
	if s.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return s.ReadU16LENext(1)[0], nil
}

// ReadUInt16BE reads a Big-Endian encoded uint16
func (s *StreamIn) ReadUInt16BE() (uint16, error) {
	if s.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return s.ReadU16BENext(1)[0], nil
}

// ReadInt16LE reads a Little-Endian encoded int16
func (s *StreamIn) ReadInt16LE() (int16, error) {
	if s.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(s.ReadU16LENext(1)[0]), nil
}

// ReadInt16BE reads a Big-Endian encoded int16
func (s *StreamIn) ReadInt16BE() (int16, error) {
	if s.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(s.ReadU16BENext(1)[0]), nil
}

// ReadUInt32LE reads a Little-Endian encoded uint32
func (s *StreamIn) ReadUInt32LE() (uint32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return s.ReadU32LENext(1)[0], nil
}

// ReadUInt32BE reads a Big-Endian encoded uint32
func (s *StreamIn) ReadUInt32BE() (uint32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return s.ReadU32BENext(1)[0], nil
}

// ReadInt32LE reads a Little-Endian encoded int32
func (s *StreamIn) ReadInt32LE() (int32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(s.ReadU32LENext(1)[0]), nil
}

// ReadInt32BE reads a Big-Endian encoded int32
func (s *StreamIn) ReadInt32BE() (int32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(s.ReadU32BENext(1)[0]), nil
}

// ReadUInt64LE reads a Little-Endian encoded uint64
func (s *StreamIn) ReadUInt64LE() (uint64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return s.ReadU64LENext(1)[0], nil
}

// ReadUInt64BE reads a Big-Endian encoded uint64
func (s *StreamIn) ReadUInt64BE() (uint64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return s.ReadU64BENext(1)[0], nil
}

// ReadInt64LE reads a Little-Endian encoded int64
func (s *StreamIn) ReadInt64LE() (int64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(s.ReadU64LENext(1)[0]), nil
}

// ReadInt64BE reads a Big-Endian encoded int64
func (s *StreamIn) ReadInt64BE() (int64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(s.ReadU64BENext(1)[0]), nil
}

// ReadFloat32LE reads a Little-Endian encoded float32
func (s *StreamIn) ReadFloat32LE() (float32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return s.ReadF32LENext(1)[0], nil
}

// ReadFloat32BE reads a Big-Endian encoded float32
func (s *StreamIn) ReadFloat32BE() (float32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return s.ReadF32BENext(1)[0], nil
}

// ReadFloat64LE reads a Little-Endian encoded float64
func (s *StreamIn) ReadFloat64LE() (float64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return s.ReadF64LENext(1)[0], nil
}

// ReadFloat64BE reads a Big-Endian encoded float64
func (s *StreamIn) ReadFloat64BE() (float64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return s.ReadF64BENext(1)[0], nil
}

// ReadBool reads a bool
func (s *StreamIn) ReadBool() (bool, error) {
	if s.Remaining() < 1 {
		return false, errors.New("Not enough data to read bool")
	}

	return s.ReadByteNext() == 1, nil
}

// ReadPID reads a PID. The size depends on the server version
func (s *StreamIn) ReadPID() (*PID, error) {
	if s.Server.LibraryVersion().GreaterOrEqual("4.0.0") {
		if s.Remaining() < 8 {
			return nil, errors.New("Not enough data to read PID")
		}

		pid, _ := s.ReadUInt64LE()

		return NewPID(pid), nil
	} else {
		if s.Remaining() < 4 {
			return nil, errors.New("Not enough data to read legacy PID")
		}

		pid, _ := s.ReadUInt32LE()

		return NewPID(pid), nil
	}
}

// ReadString reads and returns a nex string type
func (s *StreamIn) ReadString() (string, error) {
	var length int64
	var err error

	// TODO - These variable names kinda suck?
	if s.Server == nil {
		l, e := s.ReadUInt16LE()
		length = int64(l)
		err = e
	} else if s.Server.StringLengthSize() == 4 {
		l, e := s.ReadUInt32LE()
		length = int64(l)
		err = e
	} else {
		l, e := s.ReadUInt16LE()
		length = int64(l)
		err = e
	}

	if err != nil {
		return "", fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	if s.Remaining() < int(length) {
		return "", errors.New("NEX string length longer than data size")
	}

	stringData := s.ReadBytesNext(length)
	str := string(stringData)

	return strings.TrimRight(str, "\x00"), nil
}

// ReadBuffer reads a nex Buffer type
func (s *StreamIn) ReadBuffer() ([]byte, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to read NEX buffer length. %s", err.Error())
	}

	if s.Remaining() < int(length) {
		return []byte{}, errors.New("NEX buffer length longer than data size")
	}

	data := s.ReadBytesNext(int64(length))

	return data, nil
}

// ReadQBuffer reads a nex qBuffer type
func (s *StreamIn) ReadQBuffer() ([]byte, error) {
	length, err := s.ReadUInt16LE()
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to read NEX qBuffer length. %s", err.Error())
	}

	if s.Remaining() < int(length) {
		return []byte{}, errors.New("NEX qBuffer length longer than data size")
	}

	data := s.ReadBytesNext(int64(length))

	return data, nil
}

// ReadVariant reads a Variant type. This type can hold 7 different types
func (s *StreamIn) ReadVariant() (*Variant, error) {
	variant := NewVariant()

	err := variant.ExtractFromStream(s)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Variant. %s", err.Error())
	}

	return variant, nil
}

// ReadDateTime reads a DateTime type
func (s *StreamIn) ReadDateTime() (*DateTime, error) {
	value, err := s.ReadUInt64LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read DateTime value. %s", err.Error())
	}

	return NewDateTime(value), nil
}

// ReadDataHolder reads a DataHolder type
func (s *StreamIn) ReadDataHolder() (*DataHolder, error) {
	dataHolder := NewDataHolder()
	err := dataHolder.ExtractFromStream(s)
	if err != nil {
		return nil, fmt.Errorf("Failed to read DateHolder. %s", err.Error())
	}

	return dataHolder, nil
}

// ReadStationURL reads a StationURL type
func (s *StreamIn) ReadStationURL() (*StationURL, error) {
	stationString, err := s.ReadString()
	if err != nil {
		return nil, fmt.Errorf("Failed to read StationURL. %s", err.Error())
	}

	return NewStationURL(stationString), nil
}

// ReadQUUID reads a qUUID type
func (s *StreamIn) ReadQUUID() (*QUUID, error) {
	qUUID := NewQUUID()

	err := qUUID.ExtractFromStream(s)
	if err != nil {
		return nil, fmt.Errorf("Failed to read qUUID. %s", err.Error())
	}

	return qUUID, nil
}

// ReadListUInt8 reads a list of uint8 types
func (s *StreamIn) ReadListUInt8() ([]uint8, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint8> length. %s", err.Error())
	}

	if s.Remaining() < int(length) {
		return nil, errors.New("NEX List<uint8> length longer than data size")
	}

	list := make([]uint8, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt8()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint8> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt8 reads a list of int8 types
func (s *StreamIn) ReadListInt8() ([]int8, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int8> length. %s", err.Error())
	}

	if s.Remaining() < int(length) {
		return nil, errors.New("NEX List<int8> length longer than data size")
	}

	list := make([]int8, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt8()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int8> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt16LE reads a list of Little-Endian encoded uint16 types
func (s *StreamIn) ReadListUInt16LE() ([]uint16, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint16> length. %s", err.Error())
	}

	if s.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<uint16> length longer than data size")
	}

	list := make([]uint16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt16LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt16BE reads a list of Big-Endian encoded uint16 types
func (s *StreamIn) ReadListUInt16BE() ([]uint16, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint16> length. %s", err.Error())
	}

	if s.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<uint16> length longer than data size")
	}

	list := make([]uint16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt16BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt16LE reads a list of Little-Endian encoded int16 types
func (s *StreamIn) ReadListInt16LE() ([]int16, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int16> length. %s", err.Error())
	}

	if s.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<int16> length longer than data size")
	}

	list := make([]int16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt16LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt16BE reads a list of Big-Endian encoded uint16 types
func (s *StreamIn) ReadListInt16BE() ([]int16, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int16> length. %s", err.Error())
	}

	if s.Remaining() < int(length*2) {
		return nil, errors.New("NEX List<int16> length longer than data size")
	}

	list := make([]int16, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt16BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int16> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt32LE reads a list of Little-Endian encoded uint32 types
func (s *StreamIn) ReadListUInt32LE() ([]uint32, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint32> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<uint32> length longer than data size")
	}

	list := make([]uint32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt32BE reads a list of Big-Endian encoded uint32 types
func (s *StreamIn) ReadListUInt32BE() ([]uint32, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint32> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<uint32> length longer than data size")
	}

	list := make([]uint32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt32BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt32LE reads a list of Little-Endian encoded int32 types
func (s *StreamIn) ReadListInt32LE() ([]int32, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int32> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<int32> length longer than data size")
	}

	list := make([]int32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt32BE reads a list of Big-Endian encoded int32 types
func (s *StreamIn) ReadListInt32BE() ([]int32, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int32> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<int32> length longer than data size")
	}

	list := make([]int32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt32BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt64LE reads a list of Little-Endian encoded uint64 types
func (s *StreamIn) ReadListUInt64LE() ([]uint64, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint64> length. %s", err.Error())
	}

	if s.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<uint64> length longer than data size")
	}

	list := make([]uint64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt64LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListUInt64BE reads a list of Big-Endian encoded uint64 types
func (s *StreamIn) ReadListUInt64BE() ([]uint64, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<uint64> length. %s", err.Error())
	}

	if s.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<uint64> length longer than data size")
	}

	list := make([]uint64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadUInt64BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<uint64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt64LE reads a list of Little-Endian encoded int64 types
func (s *StreamIn) ReadListInt64LE() ([]int64, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int64> length. %s", err.Error())
	}

	if s.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<int64> length longer than data size")
	}

	list := make([]int64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt64LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListInt64BE reads a list of Big-Endian encoded int64 types
func (s *StreamIn) ReadListInt64BE() ([]int64, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<int64> length. %s", err.Error())
	}

	if s.Remaining() < int(length*8) {
		return nil, errors.New("NEX List<int64> length longer than data size")
	}

	list := make([]int64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadInt64BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<int64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat32LE reads a list of Little-Endian encoded float32 types
func (s *StreamIn) ReadListFloat32LE() ([]float32, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float32> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float32> length longer than data size")
	}

	list := make([]float32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadFloat32LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat32BE reads a list of Big-Endian encoded float32 types
func (s *StreamIn) ReadListFloat32BE() ([]float32, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float32> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float32> length longer than data size")
	}

	list := make([]float32, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadFloat32BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float32> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat64LE reads a list of Little-Endian encoded float64 types
func (s *StreamIn) ReadListFloat64LE() ([]float64, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float64> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float64> length longer than data size")
	}

	list := make([]float64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadFloat64LE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListFloat64BE reads a list of Big-Endian encoded float64 types
func (s *StreamIn) ReadListFloat64BE() ([]float64, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<float64> length. %s", err.Error())
	}

	if s.Remaining() < int(length*4) {
		return nil, errors.New("NEX List<float64> length longer than data size")
	}

	list := make([]float64, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadFloat64BE()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<float64> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListPID reads a list of NEX PIDs
func (s *StreamIn) ReadListPID() ([]*PID, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<String> length. %s", err.Error())
	}

	list := make([]*PID, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadPID()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<PID> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListString reads a list of NEX String types
func (s *StreamIn) ReadListString() ([]string, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<String> length. %s", err.Error())
	}

	list := make([]string, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadString()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<String> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListBuffer reads a list of NEX Buffer types
func (s *StreamIn) ReadListBuffer() ([][]byte, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<Buffer> length. %s", err.Error())
	}

	list := make([][]byte, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadBuffer()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<Buffer> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListQBuffer reads a list of NEX qBuffer types
func (s *StreamIn) ReadListQBuffer() ([][]byte, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<qBuffer> length. %s", err.Error())
	}

	list := make([][]byte, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadQBuffer()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<qBuffer> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListStationURL reads a list of NEX Station URL types
func (s *StreamIn) ReadListStationURL() ([]*StationURL, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<StationURL> length. %s", err.Error())
	}

	list := make([]*StationURL, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadStationURL()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<StationURL> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListDataHolder reads a list of NEX DataHolder types
func (s *StreamIn) ReadListDataHolder() ([]*DataHolder, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<DataHolder> length. %s", err.Error())
	}

	list := make([]*DataHolder, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadDataHolder()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<DataHolder> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// ReadListQUUID reads a list of NEX qUUID types
func (s *StreamIn) ReadListQUUID() ([]*QUUID, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<QUUID> length. %s", err.Error())
	}

	list := make([]*QUUID, 0, length)

	for i := 0; i < int(length); i++ {
		value, err := s.ReadQUUID()
		if err != nil {
			return nil, fmt.Errorf("Failed to read List<QUUID> value at index %d. %s", i, err.Error())
		}

		list = append(list, value)
	}

	return list, nil
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server ServerInterface) *StreamIn {
	return &StreamIn{
		Buffer: crunch.NewBuffer(data),
		Server: server,
	}
}

// StreamReadStructure reads a Structure type from a StreamIn
//
// Implemented as a separate function to utilize generics
func StreamReadStructure[T StructureInterface](stream *StreamIn, structure T) (T, error) {
	if structure.ParentType() != nil {
		//_, err := s.ReadStructure(structure.ParentType())
		_, err := StreamReadStructure(stream, structure.ParentType())
		if err != nil {
			return structure, fmt.Errorf("Failed to read structure parent. %s", err.Error())
		}
	}

	useStructureHeader := false

	if stream.Server != nil {
		switch server := stream.Server.(type) {
		case *PRUDPServer: // * Support QRV versions
			useStructureHeader = server.PRUDPMinorVersion >= 3
		default:
			useStructureHeader = server.LibraryVersion().GreaterOrEqual("3.5.0")
		}
	}

	if useStructureHeader {
		version, err := stream.ReadUInt8()
		if err != nil {
			return structure, fmt.Errorf("Failed to read NEX Structure version. %s", err.Error())
		}

		structureLength, err := stream.ReadUInt32LE()
		if err != nil {
			return structure, fmt.Errorf("Failed to read NEX Structure content length. %s", err.Error())
		}

		if stream.Remaining() < int(structureLength) {
			return structure, errors.New("NEX Structure content length longer than data size")
		}

		structure.SetStructureVersion(version)
	}

	err := structure.ExtractFromStream(stream)
	if err != nil {
		return structure, fmt.Errorf("Failed to read structure from s. %s", err.Error())
	}

	return structure, nil
}

// StreamReadListStructure reads and returns a list of structure types from a StreamIn
//
// Implemented as a separate function to utilize generics
func StreamReadListStructure[T StructureInterface](stream *StreamIn, structure T) ([]T, error) {
	length, err := stream.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read List<Structure> length. %s", err.Error())
	}

	structures := make([]T, 0, int(length))

	for i := 0; i < int(length); i++ {
		newStructure := structure.Copy()

		extracted, err := StreamReadStructure[T](stream, newStructure.(T))
		if err != nil {
			return nil, err
		}

		structures = append(structures, extracted)
	}

	return structures, nil
}

// StreamReadMap reads a Map type with the given key and value types from a StreamIn
//
// Implemented as a separate function to utilize generics
func StreamReadMap[K comparable, V any](s *StreamIn, keyReader func() (K, error), valueReader func() (V, error)) (map[K]V, error) {
	length, err := s.ReadUInt32LE()
	if err != nil {
		return nil, fmt.Errorf("Failed to read Map length. %s", err.Error())
	}

	m := make(map[K]V)

	for i := 0; i < int(length); i++ {
		key, err := keyReader()
		if err != nil {
			return nil, err
		}

		value, err := valueReader()
		if err != nil {
			return nil, err
		}

		m[key] = value
	}

	return m, nil
}
