package nex

import (
	"bytes"
	"strings"
	"time"
)

// String represents a NEX formatted string
// Length: Length of the null-terminated string
// String: Null-terminated string
type String struct {
	Length uint16
	String string
}

// StationURL represents a NEX Station URL
// URL: String of the station URL, containing the protocol and station options
type StationURL struct {
	URL String
}

// DateTime represents a NEX date timestamp
type DateTime struct {
	Second int
	Minute int
	Hour   int
	Day    int
	Month  int
	Year   int

	TimeStamp int
}

func NewString(str string) String {
	str = str + "\x00"
	Length := len(str)

	return String{uint16(Length), str}
}

func NewStationURL(protocol string, JSON map[string]string) String {
	var URLBuffer bytes.Buffer

	URLBuffer.WriteString(protocol + ":/")

	for key, value := range JSON {
		option := key + "=" + value + ";"
		URLBuffer.WriteString(option)
	}

	URL := URLBuffer.String()
	URL = strings.TrimRight(URL, ";")

	return NewString(URL)
}

func NewDateTime(current time.Time) DateTime {
	second := current.Second()
	minute := current.Minute()
	hour := current.Hour()
	day := current.Day()
	month := int(current.Month())
	year := current.Year()

	TimeStamp := (second | (minute << 6) | (hour << 12) | (day << 17) | (month << 22) | (year << 26))

	datetime := DateTime{}

	datetime.TimeStamp = TimeStamp

	datetime.Second = TimeStamp & 63
	datetime.Minute = (TimeStamp >> 6) & 63
	datetime.Hour = (TimeStamp >> 12) & 31
	datetime.Day = (TimeStamp >> 17) & 31
	datetime.Month = (TimeStamp >> 22) & 15
	datetime.Year = TimeStamp >> 26

	return datetime
}

/*
func main() {
	JSONBuffer := []byte(`{
		"stream":  "10",
		"type": "2",
		"PID": "2",
		"port": "60091",
		"address": "35.162.205.114",
		"sid": "1",
		"CID": "1"
	}`)

	var JSON map[string]string
	err := json.Unmarshal(JSONBuffer, &JSON)

	if err != nil {
		panic(err)
	}

	fmt.Println(NewStationURL("prudp", JSON))

	now := time.Now()

	fmt.Println(now.Second())
	fmt.Println(now.Minute())
	fmt.Println(now.Hour())
	fmt.Println(now.Day())
	fmt.Println(now.Month())
	fmt.Println(now.Year())

	datet := NewDateTime(now)

	fmt.Println(datet)

	date := "%d-%d-%d %d:%02d:%02d\n"

	fmt.Printf(date, datet.Day, datet.Month, datet.Year, datet.Hour, datet.Minute, datet.Second)
}
*/
