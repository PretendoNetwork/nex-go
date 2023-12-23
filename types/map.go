package types

// Map represents a Quazal Rendez-Vous/NEX Map type
type Map[K RVType, V RVType] struct {
	// * Rendez-Vous/NEX MapMap types can have ANY value for the key, but Go requires
	// * map keys to implement the "comparable" constraint. This is not possible with
	// * RVTypes. We have to either break spec and only allow primitives as Map keys,
	// * or store the key/value types indirectly
	keys      []K
	values    []V
	KeyType   K
	ValueType V
}

// WriteTo writes the bool to the given writable
func (m *Map[K, V]) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(uint32(m.Size()))

	for i := 0; i < len(m.keys); i++ {
		m.keys[i].WriteTo(writable)
		m.values[i].WriteTo(writable)
	}
}

// ExtractFrom extracts the bool to the given readable
func (m *Map[K, V]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	keys := make([]K, 0, length)
	values := make([]V, 0, length)

	for i := 0; i < int(length); i++ {
		key := m.KeyType.Copy()
		if err := key.ExtractFrom(readable); err != nil {
			return err
		}

		value := m.ValueType.Copy()
		if err := value.ExtractFrom(readable); err != nil {
			return err
		}

		keys = append(keys, value.(K))
		values = append(values, value.(V))
	}

	m.keys = keys
	m.values = values

	return nil
}

// Copy returns a pointer to a copy of the Map[K, V]. Requires type assertion when used
func (m Map[K, V]) Copy() RVType {
	copied := NewMap[K, V]()
	copied.keys = make([]K, len(m.keys))
	copied.values = make([]V, len(m.values))
	copied.KeyType = m.KeyType.Copy().(K)
	copied.ValueType = m.ValueType.Copy().(V)

	for i := 0; i < len(m.keys); i++ {
		copied.keys[i] = m.keys[i].Copy().(K)
		copied.values[i] = m.values[i].Copy().(V)
	}

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (m *Map[K, V]) Equals(o RVType) bool {
	if _, ok := o.(*Map[K, V]); !ok {
		return false
	}

	other := o.(*Map[K, V])

	if len(m.keys) != len(other.keys) {
		return false
	}

	if len(m.values) != len(other.values) {
		return false
	}

	for i := 0; i < len(m.keys); i++ {
		if !m.keys[i].Equals(other.keys[i]) {
			return false
		}

		if !m.values[i].Equals(other.values[i]) {
			return false
		}
	}

	return true
}

// Set sets an element to the Map internal slices
func (m *Map[K, V]) Set(key K, value V) {
	var index int = -1

	for i := 0; i < len(m.keys); i++ {
		if m.keys[i].Equals(key) {
			index = i
			break
		}
	}

	// * Replace the element if exists, otherwise push new
	if index != -1 {
		m.keys[index] = key
		m.values[index] = value
	} else {
		m.keys = append(m.keys, key)
		m.values = append(m.values, value)
	}
}

// Get returns an element from the Map. If not found, "ok" is false
func (m *Map[K, V]) Get(key K) (V, bool) {
	var index int = -1

	for i := 0; i < len(m.keys); i++ {
		if m.keys[i].Equals(key) {
			index = i
			break
		}
	}

	if index != -1 {
		return m.values[index], true
	}

	return m.ValueType.Copy().(V), false
}

// Size returns the length of the Map
func (m *Map[K, V]) Size() int {
	return len(m.keys)
}

// NewMap returns a new Map of the provided type
func NewMap[K RVType, V RVType]() *Map[K, V] {
	return &Map[K, V]{
		keys:   make([]K, 0),
		values: make([]V, 0),
	}
}
