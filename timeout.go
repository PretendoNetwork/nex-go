package nex

import (
	"context"
	"time"
)

// Timeout is an implementation of rdv::Timeout.
// Used to hold state related to resend timeouts on a packet
type Timeout struct {
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
}

// SetRTO sets the timeout field on this instance
func (t *Timeout) SetRTO(timeout time.Duration) {
	t.timeout = timeout
}

// GetRTO gets the timeout field of this instance
func (t *Timeout) RTO() time.Duration {
	return t.timeout
}

// NewTimeout creates a new Timeout
func NewTimeout() *Timeout {
	return &Timeout{}
}
