package nex

import (
	"time"
	"strings"
)

type DateTime struct {
	value uint64
}

func (datetime *DateTime) Now() uint64 {
	timestamp := time.Now()

	second := timestamp.Second()
	minute := timestamp.Minute()
	hour := timestamp.Hour()
	day := timestamp.Day()
	month := int(timestamp.Month())
	year := timestamp.Year() + 1

	datetime.value = uint64(second | (minute << 6) | (hour << 12) | (day << 17) | (month << 22) | (year << 26))

	return datetime.value
}

func NewDateTime(value uint64) *DateTime {
	return &DateTime{value: value}
}



type StationURL struct {
	// Using pointers to check for nil
	scheme        *string
	address       *string
	port          *string
	stream        *string
	sid           *string
	cid           *string
	pid           *string
	transportType *string
	rvcid         *string
	natm          *string
	natf          *string
	upnp          *string
	pmp           *string
	probeinit     *string
	prid          *string
}

func (station *StationURL) SetAddress(address *string) {
	station.address = address
}

func (station *StationURL) SetPort(port *string) {
	station.port = port
}

func (station *StationURL) SetType(transportType *string) {
	station.transportType = transportType
}

func (station *StationURL) FromString(str string) {
	split := strings.Split(str, ":/")

	station.scheme = &split[0]
	fields := split[1]

	params := strings.Split(fields, ";")

	for i := 0; i < len(params); i++ {
		param := params[i]
		split = strings.Split(param, "=")

		name := split[0]
		value := split[1]

		switch name {
		case "address":
			station.address = &value
		case "port":
			station.port = &value
		case "stream":
			station.stream = &value
		case "sid":
			station.sid = &value
		case "CID":
			station.cid = &value
		case "PID":
			station.pid = &value
		case "type":
			station.transportType = &value
		case "RVCID":
			station.rvcid = &value
		case "natm":
			station.natm = &value
		case "natf":
			station.natf = &value
		case "upnp":
			station.upnp = &value
		case "pmp":
			station.pmp = &value
		case "probeinit":
			station.probeinit = &value
		case "PRID":
			station.prid = &value
		}
	}
}

func (station *StationURL) EncodeToString() string {
	fields := []string{}

	if station.address != nil {
		fields = append(fields, "address="+*station.address)
	}

	if station.port != nil {
		fields = append(fields, "port="+*station.port)
	}

	if station.stream != nil {
		fields = append(fields, "stream="+*station.stream)
	}

	if station.sid != nil {
		fields = append(fields, "sid="+*station.sid)
	}

	if station.cid != nil {
		fields = append(fields, "CID="+*station.cid)
	}

	if station.pid != nil {
		fields = append(fields, "PID="+*station.pid)
	}

	if station.transportType != nil {
		fields = append(fields, "type="+*station.transportType)
	}

	if station.rvcid != nil {
		fields = append(fields, "RVCID="+*station.rvcid)
	}

	if station.natm != nil {
		fields = append(fields, "natm="+*station.natm)
	}

	if station.natf != nil {
		fields = append(fields, "natf="+*station.natf)
	}

	if station.upnp != nil {
		fields = append(fields, "upnp="+*station.upnp)
	}

	if station.pmp != nil {
		fields = append(fields, "pmp="+*station.pmp)
	}

	if station.probeinit != nil {
		fields = append(fields, "probeinit="+*station.probeinit)
	}

	if station.prid != nil {
		fields = append(fields, "PRID="+*station.prid)
	}

	return *station.scheme + ":/" + strings.Join(fields, ";")
}

func NewStationURL(str string) *StationURL {
	station := &StationURL{}

	if str != "" {
		station.FromString(str)
	}

	return station
}