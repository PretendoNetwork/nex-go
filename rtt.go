package nex

import (
	"math"
	"sync"
	"time"
)

const (
	alpha float64 = 1.0 / 8.0
	beta  float64 = 1.0 / 4.0
	k     float64 = 4.0
)

// RTT is an implementation of rdv::RTT.
// Used to calculate the average round trip time of reliable packets
type RTT struct {
	sync.Mutex
	lastRTT     float64
	average     float64
	variance    float64
	initialized bool
}

// Adjust updates the average RTT with the new value
func (rtt *RTT) Adjust(next time.Duration) {
	// * This calculation comes from the RFC6298 which defines RTT calculation for TCP packets
	rtt.Lock()
	if rtt.initialized {
		rtt.variance = (1.0-beta)*rtt.variance + beta*math.Abs(rtt.variance-float64(next))
		rtt.average = (1.0-alpha)*rtt.average + alpha*float64(next)
	} else {
		rtt.lastRTT = float64(next)
		rtt.variance = float64(next) / 2
		rtt.average = float64(next) + k*rtt.variance
		rtt.initialized = true
	}
	rtt.Unlock()
}

// GetRTTSmoothedAvg returns the smoothed average of this RTT, it is used in calls to the custom
// RTO calculation function set on `PRUDPEndpoint::SetCalcRetransmissionTimeoutCallback`
func (rtt *RTT) GetRTTSmoothedAvg() float64 {
	return rtt.average / 16
}

// GetRTTSmoothedDev returns the smoothed standard deviation of this RTT, it is used in calls to the custom
// RTO calculation function set on `PRUDPEndpoint::SetCalcRetransmissionTimeoutCallback`
func (rtt *RTT) GetRTTSmoothedDev() float64 {
	return rtt.variance / 8
}

// Initialized returns a bool indicating whether this RTT has been initialized
func (rtt *RTT) Initialized() bool {
	return rtt.initialized
}

// GetRTO returns the current average
func (rtt *RTT) Average() time.Duration {
	return time.Duration(rtt.average)
}

// NewRTT returns a new RTT based on the first value
func NewRTT() *RTT {
	return &RTT{}
}
