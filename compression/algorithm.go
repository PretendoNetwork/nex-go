// Package compression provides a set of compression algorithms found
// in several versions of Rendez-Vous for compressing large payloads
package compression

// Algorithm defines all the methods a compression algorithm should have
type Algorithm interface {
	Compress(payload []byte) ([]byte, error)
	Decompress(payload []byte) ([]byte, error)
	Copy() Algorithm
}
