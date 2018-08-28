package prudp

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

// CalculateV0PacketChecksum calculates the checksum of a prudpv0 packet
func CalculateV0PacketChecksum(checksum int, packet []byte) int {
	pos := 0

	sections := len(packet) / 4
	chunks := make([]uint32, 0, sections)

	for i := 0; i < sections; i++ {
		chunk := binary.LittleEndian.Uint32(packet[pos : pos+4])
		chunks = append(chunks, chunk)

		pos += 4
	}

	temp1 := Sum(chunks)
	temp := temp1 & 0xFFFFFFFF

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(temp))

	tempSum := Sum(packet[len(packet) & ^3:])

	checksum += tempSum
	tempSum = Sum(buff)
	checksum += tempSum

	return (checksum & 0xFF)
}

// Sum calculates the sum of the input
func Sum(a interface{}) int {
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

/*
func main() {
	ACCESS_KEY := "ridfebb9"
	//packet := "a1af90000000000000000065d9e3340000"

	base := Sum([]byte(ACCESS_KEY))
	checksum := CalculateV0PacketChecksum(base, []byte{161, 175, 16, 0, 0, 0, 0}) // Returns as an int

	fmt.Println(hex.EncodeToString([]byte{161, 175, 144, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	fmt.Println(strconv.FormatInt(int64(checksum), 16)) // Print it as a HEX string
	fmt.Println(checksum)
}
*/
