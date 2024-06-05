package types

import (
	"fmt"
	"strings"
	"time"
)

// DateTime is an implementation of rdv::DateTime.
// Type alias of uint64.
// The underlying value is a uint64 bit field containing date and time information.
type DateTime uint64

// WriteTo writes the DateTime to the given writable
func (dt DateTime) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt64LE(uint64(dt))
}

// ExtractFrom extracts the DateTime from the given readable
func (dt *DateTime) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt64LE()
	if err != nil {
		return fmt.Errorf("Failed to read DateTime value. %s", err.Error())
	}

	*dt = DateTime(value)
	return nil
}

// Copy returns a new copied instance of DateTime
func (dt DateTime) Copy() RVType {
	return NewDateTime(uint64(dt))
}

// Equals checks if the input is equal in value to the current instance
func (dt DateTime) Equals(o RVType) bool {
	if _, ok := o.(DateTime); !ok {
		return false
	}

	return dt == o.(DateTime)
}

// Make initilizes a DateTime with the input data
func (dt *DateTime) Make(year, month, day, hour, minute, second int) DateTime {
	*dt = DateTime(second | (minute << 6) | (hour << 12) | (day << 17) | (month << 22) | (year << 26))

	return *dt
}

// FromTimestamp converts a Time timestamp into a NEX DateTime
func (dt DateTime) FromTimestamp(timestamp time.Time) DateTime {
	year := timestamp.Year()
	month := int(timestamp.Month())
	day := timestamp.Day()
	hour := timestamp.Hour()
	minute := timestamp.Minute()
	second := timestamp.Second()

	return dt.Make(year, month, day, hour, minute, second)
}

// Now returns a NEX DateTime value of the current UTC time
func (dt DateTime) Now() DateTime {
	return dt.FromTimestamp(time.Now().UTC())
}

// Second returns the seconds value stored in the DateTime
func (dt DateTime) Second() int {
	return int(dt & 63)
}

// Minute returns the minutes value stored in the DateTime
func (dt DateTime) Minute() int {
	return int((dt >> 6) & 63)
}

// Hour returns the hours value stored in the DateTime
func (dt DateTime) Hour() int {
	return int((dt >> 12) & 31)
}

// Day returns the day value stored in the DateTime
func (dt DateTime) Day() int {
	return int((dt >> 17) & 31)
}

// Month returns the month value stored in the DateTime
func (dt DateTime) Month() time.Month {
	return time.Month((dt >> 22) & 15)
}

// Year returns the year value stored in the DateTime
func (dt DateTime) Year() int {
	return int(dt >> 26)
}

// Standard returns the DateTime as a standard time.Time
func (dt DateTime) Standard() time.Time {
	return time.Date(
		dt.Year(),
		dt.Month(),
		dt.Day(),
		dt.Hour(),
		dt.Minute(),
		dt.Second(),
		0,
		time.UTC,
	)
}

// String returns a string representation of the struct
func (dt DateTime) String() string {
	return dt.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (dt DateTime) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("DateTime{\n")
	b.WriteString(fmt.Sprintf("%svalue: %d (%s)\n", indentationValues, dt, dt.Standard().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewDateTime returns a new DateTime instance
func NewDateTime(input uint64) DateTime {
	dt := DateTime(input)
	return dt
}
