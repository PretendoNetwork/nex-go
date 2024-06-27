package nex

import "sync"

// MutexMap implements a map type with go routine safe accessors through mutex locks. Embeds sync.RWMutex
type MutexMap[K comparable, V any] struct {
	*sync.RWMutex
	real map[K]V
}

// Set sets a key to a given value
func (m *MutexMap[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()

	m.real[key] = value
}

// Get returns the given key value and a bool if found
func (m *MutexMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()

	value, ok := m.real[key]

	return value, ok
}

// GetOrSetDefault returns the value for the given key if it exists. If it does not exist, it creates a default
// with the provided function and sets it for that key
func (m *MutexMap[K, V]) GetOrSetDefault(key K, createDefault func() V) V {
	m.Lock()
	defer m.Unlock()

	value, ok := m.real[key]

	if !ok {
		value = createDefault()
		m.real[key] = value
	}

	return value
}

// Has checks if a key exists in the map
func (m *MutexMap[K, V]) Has(key K) bool {
	m.RLock()
	defer m.RUnlock()

	_, ok := m.real[key]
	return ok
}

// Delete removes a key from the internal map
func (m *MutexMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()

	delete(m.real, key)
}

// DeleteIf deletes every element if the predicate returns true.
// Returns the amount of elements deleted.
func (m *MutexMap[K, V]) DeleteIf(predicate func(key K, value V) bool) int {
	m.Lock()
	defer m.Unlock()

	amount := 0
	for key, value := range m.real {
		if predicate(key, value) {
			delete(m.real, key)
			amount++
		}
	}

	return amount
}

// RunAndDelete runs a callback and removes the key afterwards
func (m *MutexMap[K, V]) RunAndDelete(key K, callback func(key K, value V)) {
	m.Lock()
	defer m.Unlock()

	if value, ok := m.real[key]; ok {
		callback(key, value)
		delete(m.real, key)
	}
}

// Size returns the length of the internal map
func (m *MutexMap[K, V]) Size() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.real)
}

// Each runs a callback function for every item in the map
// The map should not be modified inside the callback function
// Returns true if the loop was terminated early
func (m *MutexMap[K, V]) Each(callback func(key K, value V) bool) bool {
	m.RLock()
	defer m.RUnlock()

	for key, value := range m.real {
		if callback(key, value) {
			return true
		}
	}

	return false
}

// Clear removes all items from the `real` map
// Accepts an optional callback function ran for every item before it is deleted
func (m *MutexMap[K, V]) Clear(callback func(key K, value V)) {
	m.Lock()
	defer m.Unlock()

	for key, value := range m.real {
		if callback != nil {
			callback(key, value)
		}
		delete(m.real, key)
	}
}

// NewMutexMap returns a new instance of MutexMap with the provided key/value types
func NewMutexMap[K comparable, V any]() *MutexMap[K, V] {
	return &MutexMap[K, V]{
		RWMutex: &sync.RWMutex{},
		real:    make(map[K]V),
	}
}
