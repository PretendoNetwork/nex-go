package encryption

import (
	"crypto/rc4"
)

// QuazalRC4 encrypts data with RC4. Each iteration uses a new cipher instance. The key is always CD&ML
type QuazalRC4 struct {
	key             []byte
	cipher          *rc4.Cipher
	decipher        *rc4.Cipher
	cipheredCount   uint64
	decipheredCount uint64
}

// Key returns the crypto key
func (r *QuazalRC4) Key() []byte {
	return r.key
}

// SetKey sets the crypto key and updates the ciphers
func (r *QuazalRC4) SetKey(key []byte) error {
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

// Encrypt encrypts the payload with the outgoing QuazalRC4 stream
func (r *QuazalRC4) Encrypt(payload []byte) ([]byte, error) {
	r.SetKey([]byte("CD&ML"))

	ciphered := make([]byte, len(payload))

	r.cipher.XORKeyStream(ciphered, payload)

	r.cipheredCount += uint64(len(payload))

	return ciphered, nil
}

// Decrypt decrypts the payload with the incoming QuazalRC4 stream
func (r *QuazalRC4) Decrypt(payload []byte) ([]byte, error) {
	r.SetKey([]byte("CD&ML"))

	deciphered := make([]byte, len(payload))

	r.decipher.XORKeyStream(deciphered, payload)

	r.decipheredCount += uint64(len(payload))

	return deciphered, nil
}

// Copy returns a copy of the algorithm while retaining it's state
func (r *QuazalRC4) Copy() Algorithm {
	copied := NewQuazalRC4Encryption()

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

// NewQuazalRC4Encryption returns a new instance of the QuazalRC4 encryption
func NewQuazalRC4Encryption() *QuazalRC4 {
	return &QuazalRC4{
		key: make([]byte, 0),
	}
}
