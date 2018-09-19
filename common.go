package nex

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
)

type Counter struct { // we could potentially use this in the client to auto-increment SequenceIDOut
	value uint16
}

func (counter Counter) Value() uint16 {
	return counter.value
}
func (counter *Counter) Increment() uint16 {
	counter.value++
	return counter.Value()
}

var Types = make(map[string]uint16, 5)
var Flags = make(map[string]uint16, 5)

func init() {
	Types["Syn"] = 0
	Types["Connect"] = 1
	Types["Data"] = 2
	Types["Disconnect"] = 3
	Types["Ping"] = 4

	Flags["Ack"] = 0x001
	Flags["Reliable"] = 0x002
	Flags["NeedAck"] = 0x004
	Flags["HasSize"] = 0x008
	Flags["MultiAck"] = 0x200
}

func readInt(data []byte, endianness binary.ByteOrder) (ret int) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, endianness, &ret)
	return
}

func readUInt16(data []byte, endianness binary.ByteOrder) (ret uint16) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, endianness, &ret)
	return
}

func readUInt32(data []byte, endianness binary.ByteOrder) (ret uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, endianness, &ret)
	return
}

// Sum calculates the sum of the input
func sum(a interface{}) int {
	var (
		va = reflect.ValueOf(a)
		r  = float64(0)
		vb reflect.Value
	)

	if va.Kind() != reflect.Slice {
		panic(fmt.Sprintf("a %s is not a slice!", va.Kind().String()))
	}

	for i := 0; i < va.Len(); i++ {
		vb = va.Index(i)

		switch vb.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			r += float64(vb.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			r += float64(vb.Uint())
		case reflect.Float32, reflect.Float64:
			r += vb.Float()
		default:
			panic(fmt.Sprintf("a %s is not a summable type!", vb.Kind().String()))
		}
	}

	return int(r)
}

func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
