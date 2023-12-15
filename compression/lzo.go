package compression

import (
	"bytes"
	"fmt"

	"github.com/cyberdelia/lzo"
)

// TODO - Untested. I think this works. Maybe. Verify and remove this comment

// LZO implements packet payload compression using LZO
type LZO struct{}

// Compress compresses the payload using LZO
func (l *LZO) Compress(payload []byte) ([]byte, error) {
	var compressed bytes.Buffer

	lzoWriter := lzo.NewWriter(&compressed)

	_, err := lzoWriter.Write(payload)
	if err != nil {
		return []byte{}, err
	}

	err = lzoWriter.Close()
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

// Decompress decompresses the payload using LZO
func (l *LZO) Decompress(payload []byte) ([]byte, error) {
	compressionRatio := payload[0]
	compressed := payload[1:]

	if compressionRatio == 0 {
		// * Compression ratio of 0 means no compression
		return compressed, nil
	}

	reader := bytes.NewReader(compressed)
	decompressed := bytes.Buffer{}

	lzoReader, err := lzo.NewReader(reader)
	if err != nil {
		return []byte{}, err
	}

	_, err = decompressed.ReadFrom(lzoReader)
	if err != nil {
		return []byte{}, err
	}

	err = lzoReader.Close()
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
