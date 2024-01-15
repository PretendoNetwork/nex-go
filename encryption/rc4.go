package encryption

import (
	"crypto/rc4"
)

// RC4 does no encryption. Payloads are returned as-is
type RC4 struct {
	key             []byte
	cipher          *rc4.Cipher
	decipher        *rc4.Cipher
	cipheredCount   uint64
	decipheredCount uint64
}

// Key returns the crypto key
func (r *RC4) Key() []byte {
	return r.key
}

// SetKey sets the crypto key and updates the ciphers
func (r *RC4) SetKey(key []byte) error {
	r.key = key

	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return err
	}

	decipher, err := rc4.NewCipher(key)
	if err != nil {
		return err
	}

	r.cipher = cipher
	r.decipher = decipher

	return nil
}

// Encrypt encrypts the payload with the outgoing RC4 stream
func (r *RC4) Encrypt(payload []byte) ([]byte, error) {
	ciphered := make([]byte, len(payload))

	r.cipher.XORKeyStream(ciphered, payload)

	r.cipheredCount += uint64(len(payload))

	return ciphered, nil
}

// Decrypt decrypts the payload with the incoming RC4 stream
func (r *RC4) Decrypt(payload []byte) ([]byte, error) {
	deciphered := make([]byte, len(payload))

	r.decipher.XORKeyStream(deciphered, payload)

	r.decipheredCount += uint64(len(payload))

	return deciphered, nil
}

// Copy returns a copy of the algorithm while retaining it's state
func (r *RC4) Copy() Algorithm {
	copied := NewRC4Encryption()

	copied.SetKey(r.key)

	// * crypto/rc4 does not expose a way to directly copy streams and retain their state.
	// * This just discards the number of iterations done in the original ciphers to sync
	// * the copied ciphers states to the original
	for i := 0; i < int(r.cipheredCount); i++ {
		copied.cipher.XORKeyStream([]byte{0}, []byte{0})
	}

	for i := 0; i < int(r.decipheredCount); i++ {
		copied.decipher.XORKeyStream([]byte{0}, []byte{0})
	}

	copied.cipheredCount = r.cipheredCount
	copied.decipheredCount = r.decipheredCount

	return copied
}

// NewRC4Encryption returns a new instance of the RC4 encryption
func NewRC4Encryption() *RC4 {
	encryption := &RC4{
		key: make([]byte, 0),
	}

	encryption.SetKey([]byte("CD&ML")) // TODO - Make this configurable?

	return encryption
}
