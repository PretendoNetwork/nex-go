package nex

// Counter represents an incremental counter
type Counter struct {
	value uint32
}

// Value returns the counters current value
func (counter Counter) Value() uint32 {
	return counter.value
}

// Increment increments the counter by 1 and returns the value
func (counter *Counter) Increment() uint32 {
	counter.value++
	return counter.Value()
}

// NewCounter returns a new Counter, with a starting number
func NewCounter(start uint32) *Counter {
	counter := &Counter{value: start}

	return counter
}
