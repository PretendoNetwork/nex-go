package nex

import (
	"golang.org/x/exp/constraints"
)

type numeric interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

// Counter represents an incremental counter of a specific numeric type
type Counter[T numeric] struct {
	Value T
}

// Next increments the counter by 1 and returns the new value
func (c *Counter[T]) Next() T {
	c.Value++
	return c.Value
}

// NewCounter returns a new Counter, with a starting number
func NewCounter[T numeric](start T) *Counter[T] {
	return &Counter[T]{Value: start}
}
