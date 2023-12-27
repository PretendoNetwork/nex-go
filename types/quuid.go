package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// QUUID represents a QRV qUUID type. This type encodes a UUID in little-endian byte order
type QUUID struct {
	Data []byte
}

// WriteTo writes the QUUID to the given writable
func (qu *QUUID) WriteTo(writable Writable) {
	writable.Write(qu.Data)
}

// ExtractFrom extracts the QUUID from the given readable
func (qu *QUUID) ExtractFrom(readable Readable) error {
	if readable.Remaining() < uint64(16) {
		return errors.New("Not enough data left to read qUUID")
	}

	qu.Data, _ = readable.Read(16)

	return nil
}

// Copy returns a new copied instance of qUUID
func (qu *QUUID) Copy() RVType {
	copied := NewQUUID()

	copied.Data = make([]byte, len(qu.Data))

	copy(copied.Data, qu.Data)

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (qu *QUUID) Equals(o RVType) bool {
	if _, ok := o.(*QUUID); !ok {
		return false
	}

	return qu.GetStringValue() == (o.(*QUUID)).GetStringValue()
}

// String returns a string representation of the struct
func (qu *QUUID) String() string {
	return qu.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (qu *QUUID) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("qUUID{\n")
	b.WriteString(fmt.Sprintf("%sUUID: %s\n", indentationValues, qu.GetStringValue()))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// GetStringValue returns the UUID encoded in the qUUID
func (qu *QUUID) GetStringValue() string {
	// * Create copy of the data since slices.Reverse modifies the slice in-line
	data := make([]byte, len(qu.Data))
	copy(data, qu.Data)

	if len(data) != 16 {
		// * Default dummy UUID as found in WATCH_DOGS
		return "00000000-0000-0000-0000-000000000002"
	}

	section1 := data[0:4]
	section2 := data[4:6]
	section3 := data[6:8]
	section4 := data[8:10]
	section5_1 := data[10:12]
	section5_2 := data[12:14]
	section5_3 := data[14:16]

	slices.Reverse(section1)
	slices.Reverse(section2)
	slices.Reverse(section3)
	slices.Reverse(section4)
	slices.Reverse(section5_1)
	slices.Reverse(section5_2)
	slices.Reverse(section5_3)

	var b strings.Builder

	b.WriteString(hex.EncodeToString(section1))
	b.WriteString("-")
	b.WriteString(hex.EncodeToString(section2))
	b.WriteString("-")
	b.WriteString(hex.EncodeToString(section3))
	b.WriteString("-")
	b.WriteString(hex.EncodeToString(section4))
	b.WriteString("-")
	b.WriteString(hex.EncodeToString(section5_1))
	b.WriteString(hex.EncodeToString(section5_2))
	b.WriteString(hex.EncodeToString(section5_3))

	return b.String()
}

// FromString converts a UUID string to a qUUID
func (qu *QUUID) FromString(uuid string) error {
	sections := strings.Split(uuid, "-")
	if len(sections) != 5 {
		return fmt.Errorf("Invalid UUID. Not enough sections. Expected 5, got %d", len(sections))
	}

	data := make([]byte, 0, 16)

	var appendSection = func(section string, expectedSize int) error {
		sectionBytes, err := hex.DecodeString(section)
		if err != nil {
			return err
		}

		if len(sectionBytes) != expectedSize {
			return fmt.Errorf("Unexpected section size. Expected %d, got %d", expectedSize, len(sectionBytes))
		}

		data = append(data, sectionBytes...)

		return nil
	}

	if err := appendSection(sections[0], 4); err != nil {
		return fmt.Errorf("Failed to read UUID section 1. %s", err.Error())
	}

	if err := appendSection(sections[1], 2); err != nil {
		return fmt.Errorf("Failed to read UUID section 2. %s", err.Error())
	}

	if err := appendSection(sections[2], 2); err != nil {
		return fmt.Errorf("Failed to read UUID section 3. %s", err.Error())
	}

	if err := appendSection(sections[3], 2); err != nil {
		return fmt.Errorf("Failed to read UUID section 4. %s", err.Error())
	}

	if err := appendSection(sections[4], 6); err != nil {
		return fmt.Errorf("Failed to read UUID section 5. %s", err.Error())
	}

	slices.Reverse(data[0:4])
	slices.Reverse(data[4:6])
	slices.Reverse(data[6:8])
	slices.Reverse(data[8:10])
	slices.Reverse(data[10:12])
	slices.Reverse(data[12:14])
	slices.Reverse(data[14:16])

	qu.Data = make([]byte, 0, 16)

	copy(qu.Data, data)

	return nil
}

// NewQUUID returns a new qUUID
func NewQUUID() *QUUID {
	return &QUUID{
		Data: make([]byte, 0, 16),
	}
}
