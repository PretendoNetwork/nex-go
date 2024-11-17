// Package types provides types used in Quazal Rendez-Vous/NEX
package types

// RVType represents a Quazal Rendez-Vous/NEX type.
// This includes primitives and custom types.
type RVType interface {
	WriteTo(writable Writable) // Writes the type data to the given Writable stream
	Copy() RVType              // Returns a non-pointer copy of the type data. Complex types are deeply copied
	CopyRef() RVTypePtr        // Returns a pointer to a copy of the type data. Complex types are deeply copied. Useful for obtaining a pointer without reflection, though limited to copies
	Equals(other RVType) bool  // Checks if the input type is strictly equal to the current type
}

// RVTypePtr represents a pointer to an RVType.
// Used to separate pointer receivers for easier type checking.
type RVTypePtr interface {
	RVType
	ExtractFrom(readable Readable) error // Reads the type data to the given Readable stream
	Deref() RVType                       // Returns the raw type data from a pointer. Useful for ensuring you have raw data without reflection
}
