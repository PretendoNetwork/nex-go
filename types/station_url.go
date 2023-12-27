package types

import (
	"fmt"
	"strings"
)

// StationURL contains the data for a NEX station URL
type StationURL struct {
	local  bool // * Not part of the data structure. Used for easier lookups elsewhere
	public bool // * Not part of the data structure. Used for easier lookups elsewhere
	Scheme string
	Fields map[string]string
}

// WriteTo writes the StationURL to the given writable
func (s *StationURL) WriteTo(writable Writable) {
	str := NewString(s.EncodeToString())

	str.WriteTo(writable)
}

// ExtractFrom extracts the StationURL from the given readable
func (s *StationURL) ExtractFrom(readable Readable) error {
	str := NewString("")

	if err := str.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read StationURL. %s", err.Error())
	}

	s.FromString(str.Value)

	return nil
}

// Copy returns a new copied instance of StationURL
func (s *StationURL) Copy() RVType {
	return NewStationURL(s.EncodeToString())
}

// Equals checks if the input is equal in value to the current instance
func (s *StationURL) Equals(o RVType) bool {
	if _, ok := o.(*StationURL); !ok {
		return false
	}

	other := o.(*StationURL)

	if s.local != other.local {
		return false
	}

	if s.public != other.public {
		return false
	}

	if s.Scheme != other.Scheme {
		return false
	}

	if len(s.Fields) != len(other.Fields) {
		return false
	}

	for key, value1 := range s.Fields {
		value2, ok := other.Fields[key]
		if !ok || value1 != value2 {
			return false
		}
	}

	return true
}

// SetLocal marks the StationURL as an local URL
func (s *StationURL) SetLocal() {
	s.local = true
	s.public = false
}

// SetPublic marks the StationURL as an public URL
func (s *StationURL) SetPublic() {
	s.local = false
	s.public = true
}

// IsLocal checks if the StationURL is a local URL
func (s *StationURL) IsLocal() bool {
	return s.local
}

// IsPublic checks if the StationURL is a public URL
func (s *StationURL) IsPublic() bool {
	return s.public
}

// FromString parses the StationURL data from a string
func (s *StationURL) FromString(str string) {
	if str == "" {
		return
	}

	split := strings.Split(str, ":/")

	s.Scheme = split[0]

	// * Return if there are no fields
	if split[1] == "" {
		return
	}

	fields := strings.Split(split[1], ";")

	for i := 0; i < len(fields); i++ {
		field := strings.Split(fields[i], "=")

		key := field[0]
		value := field[1]

		s.Fields[key] = value
	}
}

// EncodeToString encodes the StationURL into a string
func (s *StationURL) EncodeToString() string {
	// * Don't return anything if no scheme is set
	if s.Scheme == "" {
		return ""
	}

	fields := make([]string, 0)

	for key, value := range s.Fields {
		fields = append(fields, fmt.Sprintf("%s=%s", key, value))
	}

	return s.Scheme + ":/" + strings.Join(fields, ";")
}

// String returns a string representation of the struct
func (s *StationURL) String() string {
	return s.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (s *StationURL) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("StationURL{\n")
	b.WriteString(fmt.Sprintf("%surl: %q\n", indentationValues, s.EncodeToString()))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewStationURL returns a new StationURL
func NewStationURL(str string) *StationURL {
	stationURL := &StationURL{
		Fields: make(map[string]string),
	}

	stationURL.FromString(str)

	return stationURL
}
