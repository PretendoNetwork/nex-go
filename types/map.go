package types

import (
	"fmt"
	"reflect"
	"strings"
)

// Map represents a Quazal Rendez-Vous/NEX Map type.
// Type alias of map[K]V.
// There is not an official type in either the rdv or nn::nex namespaces.
// May have any RVType as both a key and value.
type Map[K RVComparable, V RVType] map[K]V

// WriteTo writes the Map to the given writable
func (m Map[K, V]) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(uint32(len(m)))

	for key, value := range m {
		key.WriteTo(writable)
		value.WriteTo(writable)
	}
}

func (m Map[K, V]) newKeyType() K {
	var k K
	kType := reflect.TypeOf(k).Elem()
	return reflect.New(kType).Interface().(K)
}

func (m Map[K, V]) newValueType() V {
	var v V
	vType := reflect.TypeOf(v).Elem()
	return reflect.New(vType).Interface().(V)
}

// ExtractFrom extracts the Map from the given readable
func (m *Map[K, V]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	for i := 0; i < int(length); i++ {
		key := m.newKeyType()
		if err := key.ExtractFrom(readable); err != nil {
			return err
		}

		value := m.newValueType()
		if err := value.ExtractFrom(readable); err != nil {
			return err
		}

		(*m)[key] = value
	}

	return nil
}

// Copy returns a pointer to a copy of the Map. Requires type assertion when used
func (m Map[K, V]) Copy() RVType {
	copied := make(Map[K, V])

	for key, value := range m {
		copied[key.Copy().(K)] = value.Copy().(V)
	}

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (m Map[K, V]) Equals(o RVType) bool {
	if _, ok := o.(*Map[K, V]); !ok {
		return false
	}

	other := *o.(*Map[K, V])

	if len(m) != len(other) {
		return false
	}

	for key, value := range m {
		if otherValue, ok := other[key]; !ok || !value.Equals(otherValue) {
			return false
		}
	}

	return true
}

// String returns a string representation of the struct
func (m *Map[K, V]) String() string {
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
func NewMap[K RVComparable, V RVType]() *Map[K, V] {
	m := make(Map[K, V])
	return &m
}
