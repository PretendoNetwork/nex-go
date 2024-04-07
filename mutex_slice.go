package nex

import "sync"

// TODO - This currently only properly supports Go native types, due to the use of == for comparisons. Can this be updated to support custom types?

// MutexSlice implements a slice type with go routine safe accessors through mutex locks.
//
// Embeds sync.RWMutex.
type MutexSlice[V comparable] struct {
	*sync.RWMutex
	real []V
}

// Add adds a value to the slice
func (m *MutexSlice[V]) Add(value V) {
	m.Lock()
	defer m.Unlock()

	m.real = append(m.real, value)
}

// Delete removes the first instance of the given value from the slice.
//
// Returns true if the value existed and was deleted, otherwise returns false.
func (m *MutexSlice[V]) Delete(value V) bool {
	m.Lock()
	defer m.Unlock()

	for i, v := range m.real {
		if v == value {
			m.real = append(m.real[:i], m.real[i+1:]...)
			return true
		}
	}

	return false
}

// DeleteAll removes all instances of the given value from the slice.
//
// Returns true if the value existed and was deleted, otherwise returns false.
func (m *MutexSlice[V]) DeleteAll(value V) bool {
	m.Lock()
	defer m.Unlock()

	newSlice := make([]V, 0)
	oldLength := len(m.real)

	for _, v := range m.real {
		if v != value {
			newSlice = append(newSlice, v)
		}
	}

	m.real = newSlice

	return len(newSlice) < oldLength
}

// Has checks if the slice contains the given value.
func (m *MutexSlice[V]) Has(value V) bool {
	m.Lock()
	defer m.Unlock()

	for _, v := range m.real {
		if v == value {
			return true
		}
	}

	return false
}

// GetIndex checks if the slice contains the given value and returns it's index.
//
// Returns -1 if the value does not exist in the slice.
func (m *MutexSlice[V]) GetIndex(value V) int {
	m.Lock()
	defer m.Unlock()

	for i, v := range m.real {
		if v == value {
			return i
		}
	}

	return -1
}

// At returns value at the given index.
//
// Returns a bool indicating if the value was found successfully.
func (m *MutexSlice[V]) At(index int) (V, bool) {
	m.Lock()
	defer m.Unlock()

	if index >= len(m.real) {
		return *new(V), false
	}

	return m.real[index], true
}

// Values returns the internal slice.
func (m *MutexSlice[V]) Values() []V {
	m.Lock()
	defer m.Unlock()

	return m.real
}

// Size returns the length of the internal slice
func (m *MutexSlice[V]) Size() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.real)
}

// Each runs a callback function for every item in the slice.
//
// The slice cannot be modified inside the callback function.
//
// Returns true if the loop was terminated early.
func (m *MutexSlice[V]) Each(callback func(index int, value V) bool) bool {
	m.RLock()
	defer m.RUnlock()

	for i, value := range m.real {
		if callback(i, value) {
			return true
		}
	}

	return false
}

// Clear removes all items from the slice.
func (m *MutexSlice[V]) Clear() {
	m.Lock()
	defer m.Unlock()

	m.real = make([]V, 0)
}

// NewMutexSlice returns a new instance of MutexSlice with the provided value type
func NewMutexSlice[V comparable]() *MutexSlice[V] {
	return &MutexSlice[V]{
		RWMutex: &sync.RWMutex{},
		real:    make([]V, 0),
	}
}
