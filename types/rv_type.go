// Package types provides types used in Quazal Rendez-Vous/NEX
package types

// RVType represents a Quazal Rendez-Vous/NEX type.
// This includes primitives and custom types.
type RVType interface {
	WriteTo(writable Writable)
	Copy() RVType
	CopyRef() RVTypePtr
	Equals(other RVType) bool
}

// RVTypePtr represents a pointer to an RVType.
// Used to separate pointer receivers for easier type checking.
type RVTypePtr interface {
	RVType
	ExtractFrom(readable Readable) error
	Deref() RVType
}
