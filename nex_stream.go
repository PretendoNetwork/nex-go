package nex

import (
	"bytes"
	"encoding/binary"
)

type InputStream struct {
	data []byte
	pos  int
}

type OutputStream struct {
	data []byte
}

func NewInputStream(data []byte) InputStream {
	return InputStream{
		data: data,
		pos:  0,
	}
}

func NewOutputStream() OutputStream {
	return OutputStream{
		data: []byte{},
	}
}

/*
	InputStream methods
*/

func (stream *InputStream) Seek(pos int) {
	stream.pos = pos
}

func (stream *InputStream) Skip(len int) {
	stream.pos += len
}

func (stream *InputStream) Read(len int) []byte {
	data := stream.data[stream.pos : stream.pos+len]
	stream.pos += len

	return data
}

func (stream *InputStream) Bytes(len int) []byte {
	return stream.Read(len)
}

func (stream *InputStream) Byte() []byte {
	return stream.Read(1)
}

func (stream *InputStream) UInt8() (ret uint8) {
	data := stream.Byte()
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func (stream *InputStream) UInt16LE() (ret uint16) {
	data := stream.Bytes(2)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func (stream *InputStream) UInt32LE() (ret uint32) {
	data := stream.Bytes(4)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func (stream *InputStream) UInt64LE() (ret uint64) {
	data := stream.Bytes(8)
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func (stream *InputStream) String() string {
	length := stream.UInt16LE()
	str := stream.Bytes(int(length))

	return string(str[:])[:length-1]
}

func (stream *InputStream) Buffer() []byte {
	return stream.Bytes(int(stream.UInt32LE()))
}

func (stream *InputStream) QBuffer() []byte {
	return stream.Bytes(int(stream.UInt16LE()))
}

func (stream *InputStream) DataHolder() DataHolder {
	data := DataHolder{
		Name:       stream.String(),
		Length:     stream.UInt32LE(),
		DataLength: stream.UInt32LE(),
	}
	data.Data = stream.Bytes(int(data.DataLength))

	return data
}

/*
	OutputStream methods
*/

func (stream *OutputStream) Bytes() []byte {
	return stream.data
}

func (stream *OutputStream) UInt16LE(val uint16) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, val)

	stream.data = append(stream.data, data...)
}

func (stream *OutputStream) UInt32LE(val uint32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, val)

	stream.data = append(stream.data, data...)
}

func (stream *OutputStream) UInt64LE(val uint64) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, val)

	stream.data = append(stream.data, data...)
}

func (stream *OutputStream) String(str string) {
	if len(str) <= 0 {
		stream.UInt16LE(uint16(0))
	} else {
		str = str + "\000"
		length := len(str)

		stream.UInt16LE(uint16(length))

		data := []byte(str)

		stream.data = append(stream.data, data...)
	}
}

func (stream *OutputStream) Buffer(buffer []byte) {
	length := len(buffer)

	stream.UInt32LE(uint32(length))
	stream.data = append(stream.data, buffer...)
}

func (stream *OutputStream) QBuffer(buffer []byte) {
	length := len(buffer)

	stream.UInt16LE(uint16(length))
	stream.data = append(stream.data, buffer...)
}

/*
data := make([]byte, 0, 4)
buffer := bytes.NewBuffer(data)
binary.Write(buffer, binary.LittleEndian, uint32(0x8068000B))
*/
