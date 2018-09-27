package nex

import (
	"bytes"
	"strings"
)

// DataHolder represents a generic data holder
type DataHolder struct {
	Name       string
	Length     uint32
	DataLength uint32
	Data       []byte
}

// NewStationURL returns a new station URL string
func NewStationURL(protocol string, JSON map[string]string) string {
	var URLBuffer bytes.Buffer

	URLBuffer.WriteString(protocol + ":/")

	for key, value := range JSON {
		option := key + "=" + value + ";"
		URLBuffer.WriteString(option)
	}

	URL := URLBuffer.String()
	URL = strings.TrimRight(URL, ";")

	return URL
}
