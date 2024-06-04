// Package types provides types used in Quazal Rendez-Vous/NEX
package types

// RVType represents a Quazal Rendez-Vous/NEX type.
// This includes primitives and custom types.
type RVType interface {
	WriteTo(writable Writable)
	ExtractFrom(readable Readable) error
	Copy() RVType
	Equals(other RVType) bool
}

// RVType represents a Quazal Rendez-Vous/NEX type which
// implements the "comparable" type constraint.
type RVComparable interface {
	comparable
	RVType
}
