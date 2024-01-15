package encryption

// Dummy does no encryption. Payloads are returned as-is
type Dummy struct {
	key []byte
}

// Key returns the crypto key
func (d *Dummy) Key() []byte {
	return d.key
}

// SetKey sets the crypto key
func (d *Dummy) SetKey(key []byte) error {
	d.key = key

	return nil
}

// Encrypt does nothing
func (d *Dummy) Encrypt(payload []byte) ([]byte, error) {
	return payload, nil
}

// Decrypt does nothing
func (d *Dummy) Decrypt(payload []byte) ([]byte, error) {
	return payload, nil
}

// Copy returns a copy of the algorithm while retaining it's state
func (d *Dummy) Copy() Algorithm {
	copied := NewDummyEncryption()

	copied.key = d.key

	return copied
}

// NewDummyEncryption returns a new instance of the Dummy encryption
func NewDummyEncryption() *Dummy {
	return &Dummy{
		key: make([]byte, 0),
	}
}
