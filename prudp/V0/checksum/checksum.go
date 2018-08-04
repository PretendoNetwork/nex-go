package Checksum

import (
	"encoding/binary" // Used only for the display
	"errors"
	"fmt"
	"os"
	"strconv"
)

func main() {
	ACCESS_KEY := "9f2b4678"
	packet := "a1af90000000000000000065d9e3340000"

	checksum, err := CalculateChecksum(ACCESS_KEY, []byte(packet)) // Returns as an int
	if err != nil {

		fmt.Printf("[err]: unable to calculate checksum for the packet...\n")
		fmt.Printf("       %v\n", err)
		os.Exit(1)

	}

	fmt.Println(strconv.FormatInt(int64(checksum), 16)) // Print it as a HEX string
}

// CalculateChecksum calculates the checksum of a prudpv0 packet
func CalculateChecksum(key string, packet []byte) (int, error) {

	// create the initial checksum variable
	checksum, err := sum([]byte(key))
	if err != nil {

		return 0, err

	}

	// position in the packet
	pos := 0

	// calculate the number of sections and make an array with that many indexes
	sections := len(packet) / 4
	chunks := make([]uint32, 0, sections)

	// loop over the array and add the chunks into it
	for i := 0; i < sections; i++ {

		// get the chunk and add it into the chunk array
		chunk := binary.LittleEndian.Uint32(packet[pos : pos+4])
		chunks = append(chunks, chunk)

		// increment the position counter by four
		pos += 4

	}

	// get the sum of the chunk array
	temp1, err := sum(chunks)
	if err != nil {

		return 0, err

	}
	temp := temp1 & 0xFFFFFFFF

	// make a byte buffer and put the uint32 of the sum into it
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(temp))

	// generate the sum
	tempSum, err := sum(packet[len(packet) & ^3:])
	if err != nil {

		return 0, err

	}
	checksum += tempSum
	tempSum, err = sum(buff)
	if err != nil {

		return 0, err

	}
	checksum += tempSum

	// return the checksum
	return (checksum & 0xFF), nil

}

// CalcChecksum calculates the checksum, returning only ONE item
func CalcChecksum(key string, packet []byte) int {

	// create the initial checksum variable
	checksum, err := sum([]byte(key))
	if err != nil {

		fmt.Println(err)
		os.Exit(1)
		return 0

	}

	// position in the packet
	pos := 0

	// calculate the number of sections and make an array with that many indexes
	sections := len(packet) / 4
	chunks := make([]uint32, 0, sections)

	// loop over the array and add the chunks into it
	for i := 0; i < sections; i++ {

		// get the chunk and add it into the chunk array
		chunk := binary.LittleEndian.Uint32(packet[pos : pos+4])
		chunks = append(chunks, chunk)

		// increment the position counter by four
		pos += 4

	}

	// get the sum of the chunk array
	temp1, err := sum(chunks)
	if err != nil {

		fmt.Println(err)
		os.Exit(1)
		return 0

	}
	temp := temp1 & 0xFFFFFFFF

	// make a byte buffer and put the uint32 of the sum into it
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(temp))

	// generate the sum
	tempSum, err := sum(packet[len(packet) & ^3:])
	if err != nil {

		fmt.Println(err)
		os.Exit(1)
		return 0

	}
	checksum += tempSum
	tempSum, err = sum(buff)
	if err != nil {

		fmt.Println(err)
		os.Exit(1)
		return 0

	}
	checksum += tempSum

	// return the checksum
	return (checksum & 0xFF)

}

func sum(sliceNotAsserted interface{}) (int, error) {

	sum := 0

	var slice []int

	// assert the type of the slice
	switch v := sliceNotAsserted.(type) {
	case []int:
		slice = v
	case []int8:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []int16:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []int32:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []int64:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []uint:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []uint8:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []uint16:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []uint32:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	case []uint64:
		slice = []int{}
		for _, val := range v {

			slice = append(slice, int(val))

		}
	default:
		return 0, errors.New("not a supported type")
	}

	// loop over the slice
	for _, val := range slice {

		// update the sum
		sum += val

	}

	return sum, nil

}
