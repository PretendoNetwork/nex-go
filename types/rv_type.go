// Package types provides types used in Quazal Rendez-Vous/NEX
package types

// RVType represents a Quazal Rendez-Vous/NEX type.
// This includes primitives and custom types.
type RVType interface {
	WriteTo(writable Writable)
	Copy() RVType
	Equals(other RVType) bool
}

// RVTypePtr represents a pointer to an RVType.
// User to separate pointer receivers for easier type checking.
type RVTypePtr interface {
	RVType
	ExtractFrom(readable Readable) error
}
