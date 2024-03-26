package compression

// Dummy does no compression. Payloads are returned as-is
type Dummy struct{}

// Compress does nothing
func (d *Dummy) Compress(payload []byte) ([]byte, error) {
	return payload, nil
}

// Decompress does nothing
func (d *Dummy) Decompress(payload []byte) ([]byte, error) {
	return payload, nil
}

// Copy returns a copy of the algorithm
func (d *Dummy) Copy() Algorithm {
	return NewDummyCompression()
}

// NewDummyCompression returns a new instance of the Dummy compression
func NewDummyCompression() *Dummy {
	return &Dummy{}
}
