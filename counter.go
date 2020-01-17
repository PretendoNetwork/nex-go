package nex

// Counter represents an incremental counter
type Counter struct {
	value uint64
}

// Value returns the counters current value
func (counter Counter) Value() uint64 {
	return counter.value
}

// Increment increments the counter by 1 and returns the value
func (counter *Counter) Increment() uint64 {
	counter.value++
	return counter.Value()
}

func NewCounter(start uint64) *Counter {
	counter := &Counter{value: start}

	return counter
}