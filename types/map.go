package types

import (
	"fmt"
	"strings"
)

// Map represents a Quazal Rendez-Vous/NEX Map type.
// Type alias of map[K]V.
// There is not an official type in either the rdv or nn::nex namespaces.
// May have any RVType as both a key and value. If either they key or
// value types are not an RVType, they are ignored.
//
// Incompatible with RVType pointers!
type Map[K comparable, V RVType] map[K]V

func (m Map[K, V]) writeType(t any, writable Writable) {
	// * This just makes Map.WriteTo() a bit cleaner
	// * since it doesn't have to type check
	if rvt, ok := t.(interface{ WriteTo(writable Writable) }); ok {
		rvt.WriteTo(writable)
	}
}

// WriteTo writes the Map to the given writable
func (m Map[K, V]) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(len(m)))

	for key, value := range m {
		m.writeType(key, writable)
		m.writeType(value, writable)
	}
}

func (m Map[K, V]) extractType(t any, readable Readable) error {
	// * This just makes Map.ExtractFrom() a bit cleaner
	// * since it doesn't have to type check
	if ptr, ok := t.(RVTypePtr); ok {
		return ptr.ExtractFrom(readable)
	}

	// * Maybe support other types..?

	return fmt.Errorf("Unsupported Map type %T", t)
}

// ExtractFrom extracts the Map from the given readable
func (m *Map[K, V]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	extracted := make(Map[K, V])

	for i := 0; i < int(length); i++ {
		var key K
		if err := m.extractType(&key, readable); err != nil {
			return err
		}

		var value V
		if err := m.extractType(&value, readable); err != nil {
			return err
		}

		extracted[key] = value
	}

	*m = extracted

	return nil
}

func (m Map[K, V]) copyType(t any) RVType {
	// * This just makes Map.Copy() a bit cleaner
	// * since it doesn't have to type check
	if rvt, ok := t.(RVType); ok {
		return rvt.Copy()
	}

	// TODO - Improve this, this isn't safe
	return nil
}

// Copy returns a pointer to a copy of the Map. Requires type assertion when used
func (m Map[K, V]) Copy() RVType {
	copied := make(Map[K, V])

	for key, value := range m {
		copied[m.copyType(key).(K)] = value.Copy().(V)
	}

	return copied
}

func (m Map[K, V]) typesEqual(t1, t2 any) bool {
	// * This just makes Map.Equals() a bit cleaner
	// * since it doesn't have to type check
	if rvt1, ok := t1.(RVType); ok {
		if rvt2, ok := t2.(RVType); ok {
			return rvt1.Equals(rvt2)
		}
	}

	return false
}

// Equals checks if the input is equal in value to the current instance
func (m Map[K, V]) Equals(o RVType) bool {
	if _, ok := o.(Map[K, V]); !ok {
		return false
	}

	other := o.(Map[K, V])

	if len(m) != len(other) {
		return false
	}

	for key, value := range m {
		if otherValue, ok := other[key]; !ok || m.typesEqual(&value, &otherValue) {
			return false
		}
	}

	return true
}

// CopyRef copies the current value of the Map
// and returns a pointer to the new copy
func (m Map[K, V]) CopyRef() RVTypePtr {
	copied := m.Copy().(Map[K, V])
	return &copied
}

// Deref takes a pointer to the Map
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (m *Map[K, V]) Deref() RVType {
	return *m
}

// String returns a string representation of the struct
func (m Map[K, V]) String() string {
	return m.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (m Map[K, V]) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	if len(m) == 0 {
		b.WriteString("{}\n")
	} else {
		b.WriteString("{\n")

		for key, value := range m {
			// TODO - Special handle the the last item to not add the comma on last item
			b.WriteString(fmt.Sprintf("%s%v: %v,\n", indentationValues, key, value))
		}

		b.WriteString(fmt.Sprintf("%s}\n", indentationEnd))
	}

	return b.String()
}

// NewMap returns a new Map of the provided type
func NewMap[K comparable, V RVType]() Map[K, V] {
	return make(Map[K, V])
}
