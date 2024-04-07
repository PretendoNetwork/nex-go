package compression

import (
	"bytes"
	"compress/zlib"
	"fmt"
)

// Zlib implements packet payload compression using zlib
type Zlib struct{}

// Compress compresses the payload using zlib
func (z *Zlib) Compress(payload []byte) ([]byte, error) {
	compressed := bytes.Buffer{}

	zlibWriter := zlib.NewWriter(&compressed)

	_, err := zlibWriter.Write(payload)
	if err != nil {
		return []byte{}, err
	}

	err = zlibWriter.Close()
	if err != nil {
		return []byte{}, err
	}

	compressedBytes := compressed.Bytes()

	compressionRatio := len(payload)/len(compressedBytes) + 1

	result := make([]byte, len(compressedBytes)+1)

	result[0] = byte(compressionRatio)

	copy(result[1:], compressedBytes)

	return result, nil
}

// Decompress decompresses the payload using zlib
func (z *Zlib) Decompress(payload []byte) ([]byte, error) {
	compressionRatio := payload[0]
	compressed := payload[1:]

	if compressionRatio == 0 {
		// * Compression ratio of 0 means no compression
		return compressed, nil
	}

	reader := bytes.NewReader(compressed)
	decompressed := bytes.Buffer{}

	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return []byte{}, err
	}

	_, err = decompressed.ReadFrom(zlibReader)
	if err != nil {
		return []byte{}, err
	}

	err = zlibReader.Close()
	if err != nil {
		return []byte{}, err
	}

	decompressedBytes := decompressed.Bytes()

	ratioCheck := len(decompressedBytes)/len(compressed) + 1

	if ratioCheck != int(compressionRatio) {
		return []byte{}, fmt.Errorf("Failed to decompress payload. Got bad ratio. Expected %d, got %d", compressionRatio, ratioCheck)
	}

	return decompressedBytes, nil
}

// Copy returns a copy of the algorithm
func (z *Zlib) Copy() Algorithm {
	return NewZlibCompression()
}

// NewZlibCompression returns a new instance of the Zlib compression
func NewZlibCompression() *Zlib {
	return &Zlib{}
}
