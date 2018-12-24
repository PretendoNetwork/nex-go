package nex

import (
	"bytes"
	"encoding/binary"
)

// InputStream represents a readable NEX stream
type InputStream struct {
	data []byte
	pos  int
}

// OutputStream represents a writeable NEX stream
type OutputStream struct {
	data []byte
}

// NewInputStream returns a new InputStream
func NewInputStream(data []byte) InputStream {
	return InputStream{
		data: data,
		pos:  0,
	}
}

// NewOutputStream returns a new OutputStream
func NewOutputStream() OutputStream {
	return OutputStream{
		data: []byte{},
	}
}

/*
	InputStream methods
*/

// Seek sets the stream pointer to the given position
func (stream *InputStream) Seek(pos int) {
	stream.pos = pos
}

// Skip skips the given number of bytes
func (stream *InputStream) Skip(len int) {
	stream.pos += len
}

// Read returns a slice of the given length starting from the pointer position
func (stream *InputStream) Read(len int) []byte {
	data := stream.data[stream.pos : stream.pos+len]
	stream.pos += len

	return data
}

// Bytes reads the given number of bytes
func (stream *InputStream) Bytes(len int) []byte {
	return stream.Read(len)
}

// Byte returns a single byte
func (stream *InputStream) Byte() []byte {
	return stream.Read(1)
}

// UInt8 reads an unsigned 1 byte integer
func (stream *InputStream) UInt8() (ret uint8) {
	data := stream.Byte()
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// UInt16LE reads an unsigned 2 byte integer in little endian
func (stream *InputStream) UInt16LE() (ret uint16) {
	data := stream.Bytes(2)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// UInt32LE reads an unsigned 4 byte integer in little endian
func (stream *InputStream) UInt32LE() (ret uint32) {
	data := stream.Bytes(4)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// UInt64LE reads an unsigned 8 byte integer in little endian
func (stream *InputStream) UInt64LE() (ret uint64) {
	data := stream.Bytes(8)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// Int8 reads a signed 1 byte integer
func (stream *InputStream) Int8() (ret int8) {
	data := stream.Byte()
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// Int16LE reads a signed 2 byte integer in little endian
func (stream *InputStream) Int16LE() (ret int16) {
	data := stream.Bytes(2)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// Int32LE reads a signed 4 byte integer in little endian
func (stream *InputStream) Int32LE() (ret int32) {
	data := stream.Bytes(4)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// Int64LE reads an signed 8 byte integer in little endian
func (stream *InputStream) Int64LE() (ret int64) {
	data := stream.Bytes(8)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// Float32LE reads an signed 4 byte float in little endian
func (stream *InputStream) Float32LE() (ret float32) {
	data := stream.Bytes(4)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// Float64LE reads a 8 byte float in little endian
func (stream *InputStream) Float64LE() (ret float64) {
	data := stream.Bytes(8)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

// String reads a NEX string
func (stream *InputStream) String() string {
	length := stream.UInt16LE()
	str := stream.Bytes(int(length))

	return string(str[:])[:length-1]
}

// Buffer reads a NEX buffer
func (stream *InputStream) Buffer() []byte {
	return stream.Bytes(int(stream.UInt32LE()))
}

// QBuffer reads a NEX qBuffer
func (stream *InputStream) QBuffer() []byte {
	return stream.Bytes(int(stream.UInt16LE()))
}

// DataHolder returns a NEX DataHolder
func (stream *InputStream) DataHolder() DataHolder {
	data := DataHolder{
		Name:       stream.String(),
		Length:     stream.UInt32LE(),
		DataLength: stream.UInt32LE(),
	}
	data.Data = stream.Bytes(int(data.DataLength))

	return data
}

func (stream *InputStream) Bool() bool {
	return stream.UInt8() != 0
}

// TODO: Variant is fucking cursed
func (stream *InputStream) Variant() (ret Variant) {
	// fuck you do nothing
	return
}

/*
	OutputStream methods
*/

// Bytes returns the streams raw bytes
func (stream *OutputStream) Bytes() []byte {
	return stream.data
}

// Write directly writes to the buffer
func (stream *OutputStream) Write(data []byte) {
	stream.data = append(stream.data, data...)
}

// UInt8 writes a single unsigned byte to to the buffer
func (stream *OutputStream) UInt8(val uint8) {
	stream.Write([]byte{val})
}

// UInt16LE writes a 2 byte unsigned integer in little endian
func (stream *OutputStream) UInt16LE(val uint16) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, val)

	stream.Write(data)
}

// UInt32LE writes a 4 byte unsigned integer in little endian
func (stream *OutputStream) UInt32LE(val uint32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, val)

	stream.Write(data)
}

// UInt64LE writes a 8 byte unsigned integer in little endian
func (stream *OutputStream) UInt64LE(val uint64) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, val)

	stream.Write(data)
}

// Int8 writes a single signed byte to to the buffer
func (stream *OutputStream) Int8(val int8) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, val)
	stream.Write(data.Bytes())
}

// Int16LE writes a 2 byte unsigned integer in little endian
func (stream *OutputStream) Int16LE(val int16) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, val)
	stream.Write(data.Bytes())
}

// Int32LE writes a 4 byte unsigned integer in little endian
func (stream *OutputStream) Int32LE(val int32) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, val)
	stream.Write(data.Bytes())
}

// Int64LE writes a 8 byte unsigned integer in little endian
func (stream *OutputStream) Int64LE(val int64) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, val)
	stream.Write(data.Bytes())
}

// Float32LE reads an signed 4 byte float in little endian
func (stream *OutputStream) Float32LE(val float32) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, val)
	stream.Write(data.Bytes())
}

// Float64LE reads a 8 byte float in little endian
func (stream *OutputStream) Float64LE(val float64) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, val)
	stream.Write(data.Bytes())
}

// TODO: Variant is fucking cursed
func (stream *OutputStream) Variant(val Variant) {
	// fuck you do nothing
}

// String writes a NEX string
func (stream *OutputStream) String(str string) {
	if len(str) <= 0 {
		stream.UInt16LE(uint16(0))
	} else {
		str = str + "\000"
		length := len(str)

		stream.UInt16LE(uint16(length))

		data := []byte(str)

		stream.Write(data)
	}
}

// Buffer writes a NEX buffer
func (stream *OutputStream) Buffer(buffer []byte) {
	length := len(buffer)

	stream.UInt32LE(uint32(length))
	stream.Write(buffer)
}

// QBuffer writes a NEX qBuffer
func (stream *OutputStream) QBuffer(buffer []byte) {
	length := len(buffer)

	stream.UInt16LE(uint16(length))
	stream.Write(buffer)
}

func (stream *OutputStream) Bool(in bool) {
	if in {
		stream.UInt8(1)
	} else {
		stream.UInt8(0)
	}
}