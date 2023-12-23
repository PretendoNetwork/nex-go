package types

// StructureInterface implements all Structure methods
type StructureInterface interface {
	SetParentType(parentType StructureInterface)
	ParentType() StructureInterface
	SetStructureVersion(version uint8)
	StructureVersion() uint8
	Copy() StructureInterface
	Equals(other StructureInterface) bool
	FormatToString(indentationLevel int) string
}

// Structure represents a Quazal Rendez-Vous/NEX Structure (custom class) base struct
type Structure struct {
	parentType       StructureInterface
	structureVersion uint8
	StructureInterface
}

// SetParentType sets the Structures parent type
func (s *Structure) SetParentType(parentType StructureInterface) {
	s.parentType = parentType
}

// ParentType returns the Structures parent type. nil if the Structure does not inherit another Structure
func (s *Structure) ParentType() StructureInterface {
	return s.parentType
}

// SetStructureVersion sets the structures version. Only used in NEX 3.5+
func (s *Structure) SetStructureVersion(version uint8) {
	s.structureVersion = version
}

// StructureVersion returns the structures version. Only used in NEX 3.5+
func (s *Structure) StructureVersion() uint8 {
	return s.structureVersion
}
