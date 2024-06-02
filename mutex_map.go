package nex

import "sync"

// RawMap implements a map type with helper functions to operate it.
type RawMap[K comparable, V any] struct {
	real map[K]V
}

// MutexMap implements a map type with go routine safe accessors through mutex locks. Embeds sync.RWMutex
type MutexMap[K comparable, V any] struct {
	mutex  *sync.RWMutex
	rawMap RawMap[K, V]
}

// Set sets a key to a given value
func (m *RawMap[K, V]) Set(key K, value V) {
	m.real[key] = value
}

// Get returns the given key value and a bool if found
func (m *RawMap[K, V]) Get(key K) (V, bool) {
	value, ok := m.real[key]

	return value, ok
}

// Has checks if a key exists in the map
func (m *RawMap[K, V]) Has(key K) bool {
	_, ok := m.real[key]
	return ok
}

// Delete removes a key from the internal map
func (m *RawMap[K, V]) Delete(key K) {
	delete(m.real, key)
}

// DeleteIf deletes every element if the predicate returns true.
// Returns the amount of elements deleted.
func (m *RawMap[K, V]) DeleteIf(predicate func(key K, value V) bool) int {
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
func (m *RawMap[K, V]) RunAndDelete(key K, callback func(key K, value V)) {
	if value, ok := m.real[key]; ok {
		callback(key, value)
		delete(m.real, key)
	}
}

// Size returns the length of the internal map
func (m *RawMap[K, V]) Size() int {
	return len(m.real)
}

// Each runs a callback function for every item in the map
// The map should not be modified inside the callback function
// Returns true if the loop was terminated early
func (m *RawMap[K, V]) Each(callback func(key K, value V) bool) bool {
	for key, value := range m.real {
		if callback(key, value) {
			return true
		}
	}

	return false
}

// Clear removes all items from the map
// Accepts an optional callback function ran for every item before it is deleted
func (m *RawMap[K, V]) Clear(callback func(key K, value V)) {
	for key, value := range m.real {
		if callback != nil {
			callback(key, value)
		}
		delete(m.real, key)
	}
}

// Set sets a key to a given value
func (m *MutexMap[K, V]) Set(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.rawMap.Set(key, value)
}

// Get returns the given key value and a bool if found
func (m *MutexMap[K, V]) Get(key K) (V, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	value, ok := m.rawMap.Get(key)

	return value, ok
}

// Has checks if a key exists in the map
func (m *MutexMap[K, V]) Has(key K) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.rawMap.Has(key)
}

// Delete removes a key from the internal map
func (m *MutexMap[K, V]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.rawMap.Delete(key)
}

// DeleteIf deletes every element if the predicate returns true.
// Returns the amount of elements deleted.
func (m *MutexMap[K, V]) DeleteIf(predicate func(key K, value V) bool) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.rawMap.DeleteIf(predicate)
}

// RunAndDelete runs a callback and removes the key afterwards
func (m *MutexMap[K, V]) RunAndDelete(key K, callback func(key K, value V)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.rawMap.RunAndDelete(key, callback)
}

// Size returns the length of the internal map
func (m *MutexMap[K, V]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.rawMap.Size()
}

// Each runs a callback function for every item in the map
// The map should not be modified inside the callback function
// Returns true if the loop was terminated early
func (m *MutexMap[K, V]) Each(callback func(key K, value V) bool) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.rawMap.Each(callback)
}

// Clear removes all items from the map
// Accepts an optional callback function ran for every item before it is deleted
func (m *MutexMap[K, V]) Clear(callback func(key K, value V)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.rawMap.Clear(callback)
}

// RMutexSection read-locks the map and runs the provided callback. The map will
// be unlocked after the callback returns. Useful for critical sections where
// multiple map operations are needed. Do not perform write operations to the map.
func (m *MutexMap[K, V]) RMutexSection(callback func(realMap *RawMap[K, V])) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	callback(&m.rawMap)
}

// MutexSection write-locks the map and runs the provided callback. The map will
// be unlocked after the callback returns. Useful for critical sections where
// multiple map operations are needed.
func (m *MutexMap[K, V]) MutexSection(callback func(mapInterface *RawMap[K, V])) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	callback(&m.rawMap)
}

// NewMutexMap returns a new instance of MutexMap with the provided key/value types
func NewMutexMap[K comparable, V any]() *MutexMap[K, V] {
	return &MutexMap[K, V]{
		mutex: &sync.RWMutex{},
		rawMap: RawMap[K, V]{
			real: make(map[K]V),
		},
	}
}
