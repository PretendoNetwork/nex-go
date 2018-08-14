package prudp

import (
	"encoding/binary"
	"errors"
)

// CalculateChecksum calculates the checksum of a prudpv0 packet
// checksum: The base of the checksum.
// packet  : The clients connection state
func CalculateChecksum(checksum int, packet []byte) (int, error) {
	pos := 0

	sections := len(packet) / 4
	chunks := make([]uint32, 0, sections)

	for i := 0; i < sections; i++ {
		chunk := binary.LittleEndian.Uint32(packet[pos : pos+4])
		chunks = append(chunks, chunk)

		pos += 4
	}

	temp1, err := sum(chunks)
	if err != nil {
		return 0, err
	}
	temp := temp1 & 0xFFFFFFFF

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(temp))

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

	return (checksum & 0xFF), nil
}

func sum(SliceNotAsserted interface{}) (int, error) {
	sum := 0

	var slice []int

	switch v := SliceNotAsserted.(type) {
	case []byte:
		slice = []int{}
		for _, val := range v {
			slice = append(slice, int(val))
		}
	case []uint32:
		slice = []int{}
		for _, val := range v {
			slice = append(slice, int(val))
		}
	default:
		return 0, errors.New("not a supported type")
	}

	for _, val := range slice {
		sum += val
	}

	return sum, nil
}

/*
func main() {
	ACCESS_KEY := "9f2b4678"
	packet := "a1af90000000000000000065d9e3340000"

	base, _ := sum(ACCESS_KEY)
	checksum, err := CalculateChecksum(base, []byte(packet)) // Returns as an int
	if err != nil {
		fmt.Printf("[err]: unable to calculate checksum for the packet...\n")
		fmt.Printf("       %v\n", err)
		os.Exit(1)
	}

	fmt.Println(strconv.FormatInt(int64(checksum), 16)) // Print it as a HEX string
}
*/
