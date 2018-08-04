package converter

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

func ByteSliceToHexString(slice []byte) string {
	ret := hex.EncodeToString(slice)
	if len(ret)%2 != 0 {
		ret = "0" + ret
	}
	return ret
}

func HexStringToByteSlice(hexadecimal string) []byte {
	byteslice, err := hex.DecodeString(hexadecimal)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return byteslice
}

func ByteSliceToInt(slice []byte) int {
	_int, err := strconv.ParseInt(ByteSliceToHexString(slice), 16, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return int(_int)
}

func IntToByteSlice(number int) []byte {
	return HexStringToByteSlice(strconv.FormatInt(int64(number), 16))
}

func IntToHexString(number int) string {
	ret := strconv.FormatInt(int64(number), 16)
	if len(ret)%2 != 0 {
		ret = "0" + ret
	}
	return ret
}

func HexStringToInt(hexadecimal string) int {
	_int, err := strconv.ParseInt(hexadecimal, 16, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return int(_int)
}
