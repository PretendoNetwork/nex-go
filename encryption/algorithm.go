// Package encryption provides a set of encryption algorithms found
// in several versions of Rendez-Vous for encrypting payloads
package encryption

// Algorithm defines all the methods a compression algorithm should have
type Algorithm interface {
	Key() []byte
	SetKey(key []byte) error
	Encrypt(payload []byte) ([]byte, error)
	Decrypt(payload []byte) ([]byte, error)
	Copy() Algorithm
}
