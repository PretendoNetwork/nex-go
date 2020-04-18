package nex

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

type CompressionScheme interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
}

// DummyCompression represents no compression
type DummyCompression struct{}

// NewDummyCompression returns a new dummy compression scheme
func NewDummyCompression() *DummyCompression {
	return &DummyCompression{}
}

// Compress returns the data as-is
func (c *DummyCompression) Compress(data []byte) ([]byte, error) {
	return data, nil
}

// Decompress returns the data as-is
func (c *DummyCompression) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

// ZLibCompression represents ZLib compression
type ZLibCompression int

func NewZLibCompression(level int) *ZLibCompression {
	z := ZLibCompression(level)
	return &z
}

// Compress returns the data as-is (needs to be updated to return ZLib compressed data)
func (c *ZLibCompression) Compress(data []byte) ([]byte, error) {
	b := bytes.NewBuffer([]byte{})
	w, err := zlib.NewWriterLevel(b, int(*c))
	if err != nil {
		return []byte{}, err
	}

	_, err = w.Write(data)
	if err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

// Decompress returns the data as-is (needs to be updated to return ZLib decompressed data)
func (c *ZLibCompression) Decompress(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return []byte{}, err
	}
	defer r.Close()

	return ioutil.ReadAll(r)
}
