package types

import (
	"fmt"
	"strings"
)

// PID represents a unique number to identify a user.
// Type alias of uint64.
// The true size of this value depends on the client version.
// Legacy clients (Wii U/3DS) use a uint32, whereas modern clients (Nintendo Switch) use a uint64.
type PID uint64

// WriteTo writes the PID to the given writable
func (p PID) WriteTo(writable Writable) {
	if writable.PIDSize() == 8 {
		writable.WriteUInt64LE(uint64(p))
	} else {
		writable.WriteUInt32LE(uint32(p))
	}
}

// ExtractFrom extracts the PID from the given readable
func (p *PID) ExtractFrom(readable Readable) error {
	var pid uint64
	var err error

	if readable.PIDSize() == 8 {
		pid, err = readable.ReadUInt64LE()
	} else {
		p, e := readable.ReadUInt32LE()

		pid = uint64(p)
		err = e
	}

	if err != nil {
		return err
	}

	*p = PID(pid)
	return nil
}

// Copy returns a pointer to a copy of the PID. Requires type assertion when used
func (p PID) Copy() RVType {
	return NewPID(uint64(p))
}

// Equals checks if the input is equal in value to the current instance
func (p PID) Equals(o RVType) bool {
	if _, ok := o.(PID); !ok {
		return false
	}

	return p == o.(PID)
}

// CopyRef copies the current value of the PID
// and returns a pointer to the new copy
func (p PID) CopyRef() RVTypePtr {
	return &p
}

// Deref takes a pointer to the PID
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (p *PID) Deref() RVType {
	return *p
}

// String returns a string representation of the struct
func (p PID) String() string {
	return p.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (p PID) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("PID{\n")
	b.WriteString(fmt.Sprintf("%spid: %d\n", indentationValues, p))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewPID returns a PID instance. The real size of PID depends on the client version
func NewPID(input uint64) PID {
	p := PID(input)
	return p
}
