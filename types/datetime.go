package types

import (
	"fmt"
	"strings"
	"time"
)

// DateTime represents a NEX DateTime type
type DateTime struct {
	value uint64 // TODO - Replace this with PrimitiveU64?
}

// WriteTo writes the DateTime to the given writable
func (dt *DateTime) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt64LE(dt.value)
}

// ExtractFrom extracts the DateTime to the given readable
func (dt *DateTime) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt64LE()
	if err != nil {
		return fmt.Errorf("Failed to read DateTime value. %s", err.Error())
	}

	dt.value = value

	return nil
}

// Copy returns a new copied instance of DateTime
func (dt DateTime) Copy() RVType {
	return NewDateTime(dt.value)
}

// Equals checks if the input is equal in value to the current instance
func (dt *DateTime) Equals(o RVType) bool {
	if _, ok := o.(*DateTime); !ok {
		return false
	}

	return dt.value == o.(*DateTime).value
}

// Make initilizes a DateTime with the input data
func (dt *DateTime) Make(year, month, day, hour, minute, second int) *DateTime {
	dt.value = uint64(second | (minute << 6) | (hour << 12) | (day << 17) | (month << 22) | (year << 26))

	return dt
}

// FromTimestamp converts a Time timestamp into a NEX DateTime
func (dt *DateTime) FromTimestamp(timestamp time.Time) *DateTime {
	year := timestamp.Year()
	month := int(timestamp.Month())
	day := timestamp.Day()
	hour := timestamp.Hour()
	minute := timestamp.Minute()
	second := timestamp.Second()

	return dt.Make(year, month, day, hour, minute, second)
}

// Now returns a NEX DateTime value of the current UTC time
func (dt *DateTime) Now() *DateTime {
	return dt.FromTimestamp(time.Now().UTC())
}

// Value returns the stored DateTime time
func (dt *DateTime) Value() uint64 {
	return dt.value
}

// Second returns the seconds value stored in the DateTime
func (dt *DateTime) Second() int {
	return int(dt.value & 63)
}

// Minute returns the minutes value stored in the DateTime
func (dt *DateTime) Minute() int {
	return int((dt.value >> 6) & 63)
}

// Hour returns the hours value stored in the DateTime
func (dt *DateTime) Hour() int {
	return int((dt.value >> 12) & 31)
}

// Day returns the day value stored in the DateTime
func (dt *DateTime) Day() int {
	return int((dt.value >> 17) & 31)
}

// Month returns the month value stored in the DateTime
func (dt *DateTime) Month() time.Month {
	return time.Month((dt.value >> 22) & 15)
}

// Year returns the year value stored in the DateTime
func (dt *DateTime) Year() int {
	return int(dt.value >> 26)
}

// Standard returns the DateTime as a standard time.Time
func (dt *DateTime) Standard() time.Time {
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
func (dt *DateTime) String() string {
	return dt.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (dt *DateTime) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("DateTime{\n")
	b.WriteString(fmt.Sprintf("%svalue: %d (%s)\n", indentationValues, dt.value, dt.Standard().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewDateTime returns a new DateTime instance
func NewDateTime(value uint64) *DateTime {
	return &DateTime{value: value}
}
