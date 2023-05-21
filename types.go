package nex

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// StructureInterface implements all Structure methods
type StructureInterface interface {
	SetParentType(parentType StructureInterface)
	ParentType() StructureInterface
	ExtractFromStream(*StreamIn) error
	Bytes(*StreamOut) []byte
}

// Structure represents a nex Structure type
type Structure struct {
	parentType StructureInterface
	StructureInterface
}

// SetParentType returns the Structures parent type. nil if the type does not inherit another Structure
func (structure *Structure) SetParentType(parentType StructureInterface) {
	structure.parentType = parentType
}

// ParentType returns the Structures parent type. nil if the type does not inherit another Structure
func (structure *Structure) ParentType() StructureInterface {
	return structure.parentType
}

// Data represents a structure with no data
type Data struct {
	Structure
}

// ExtractFromStream does nothing for Data
func (data *Data) ExtractFromStream(stream *StreamIn) error {
	// Basically do nothing. Does a relative seek with 0
	stream.SeekByte(0, true)

	return nil
}

// Bytes does nothing for Data
func (data *Data) Bytes(stream *StreamOut) []byte {
	return stream.Bytes()
}

// NewData returns a new Data Structure
func NewData() *Data {
	return &Data{}
}

var dataHolderKnownObjects = make(map[string]StructureInterface)

// RegisterDataHolderType registers a structure to be a valid type in the DataHolder structure
func RegisterDataHolderType(structure StructureInterface) {
	name := reflect.TypeOf(structure).Elem().Name()
	dataHolderKnownObjects[name] = structure
}

// DataHolder represents a structure which can hold any other structure
type DataHolder struct {
	typeName   string
	length1    uint32 // length of data including length2
	length2    uint32 // length of the actual structure
	objectData StructureInterface
}

// TypeName returns the DataHolder type name
func (dataHolder *DataHolder) TypeName() string {
	return dataHolder.typeName
}

// SetTypeName sets the DataHolder type name
func (dataHolder *DataHolder) SetTypeName(typeName string) {
	dataHolder.typeName = typeName
}

// ObjectData returns the DataHolder internal object data
func (dataHolder *DataHolder) ObjectData() StructureInterface {
	return dataHolder.objectData
}

// SetObjectData sets the DataHolder internal object data
func (dataHolder *DataHolder) SetObjectData(objectData StructureInterface) {
	dataHolder.objectData = objectData
}

// ExtractFromStream extracts a DataHolder structure from a stream
func (dataHolder *DataHolder) ExtractFromStream(stream *StreamIn) error {
	// TODO - Error checks
	dataHolder.typeName, _ = stream.ReadString()
	dataHolder.length1 = stream.ReadUInt32LE()
	dataHolder.length2 = stream.ReadUInt32LE()

	dataType := dataHolderKnownObjects[dataHolder.typeName]
	if dataType == nil {
		message := fmt.Sprintf("UNKNOWN DATAHOLDER TYPE: %s", dataHolder.typeName)
		logger.Critical(message)
		return errors.New(message)
	}

	dataHolder.objectData, _ = stream.ReadStructure(dataType)

	return nil
}

// Bytes encodes the DataHolder and returns a byte array
func (dataHolder *DataHolder) Bytes(stream *StreamOut) []byte {
	contentStream := NewStreamOut(stream.Server)
	contentStream.WriteStructure(dataHolder.objectData)
	content := contentStream.Bytes()

	/*
		Technically this way of encoding a DataHolder is "wrong".
		It implies the structure of DataHolder is:

			- Name     (string)
			- Length+4 (uint32)
			- Content  (Buffer)

		However the structure as defined by the official NEX library is:

			- Name     (string)
			- Length+4 (uint32)
			- Length   (uint32)
			- Content  (bytes)

		It is convenient to treat the last 2 fields as a Buffer type, but
		it should be noted that this is not actually the case.
	*/
	stream.WriteString(dataHolder.typeName)
	stream.WriteUInt32LE(uint32(len(content) + 4))
	stream.WriteBuffer(content)

	return stream.Bytes()
}

// NewDataHolder returns a new DataHolder
func NewDataHolder() *DataHolder {
	return &DataHolder{}
}

// RVConnectionData represents a nex RVConnectionData type
type RVConnectionData struct {
	stationURL                 string
	specialProtocols           []byte
	stationURLSpecialProtocols string
	time                       uint64
	Structure
}

// SetStationURL sets the RVConnectionData station URL
func (rvConnectionData *RVConnectionData) SetStationURL(stationURL string) {
	rvConnectionData.stationURL = stationURL
}

// SetSpecialProtocols sets the RVConnectionData special protocol list (unused by Nintendo)
func (rvConnectionData *RVConnectionData) SetSpecialProtocols(specialProtocols []byte) {
	rvConnectionData.specialProtocols = specialProtocols
}

// SetStationURLSpecialProtocols sets the RVConnectionData special station URL (unused by Nintendo)
func (rvConnectionData *RVConnectionData) SetStationURLSpecialProtocols(stationURLSpecialProtocols string) {
	rvConnectionData.stationURLSpecialProtocols = stationURLSpecialProtocols
}

// SetTime sets the RVConnectionData time
func (rvConnectionData *RVConnectionData) SetTime(time uint64) {
	rvConnectionData.time = time
}

// Bytes encodes the RVConnectionData and returns a byte array
func (rvConnectionData *RVConnectionData) Bytes(stream *StreamOut) []byte {
	stream.WriteString(rvConnectionData.stationURL)
	stream.WriteUInt32LE(0) // Always 0
	stream.WriteString(rvConnectionData.stationURLSpecialProtocols)
	stream.WriteUInt64LE(rvConnectionData.time)

	return stream.Bytes()
}

// NewRVConnectionData returns a new RVConnectionData
func NewRVConnectionData() *RVConnectionData {
	rvConnectionData := &RVConnectionData{}

	return rvConnectionData
}

// DateTime represents a NEX DateTime type
type DateTime struct {
	value uint64
}

// Make initilizes a DateTime with the input data
func (datetime *DateTime) Make(year, month, day, hour, minute, second int) uint64 {
	datetime.value = uint64(second | (minute << 6) | (hour << 12) | (day << 17) | (month << 22) | (year << 26))

	return datetime.value
}

// FromTimestamp converts a Time timestamp into a NEX DateTime
func (datetime *DateTime) FromTimestamp(timestamp time.Time) uint64 {
	year := timestamp.Year()
	month := int(timestamp.Month())
	day := timestamp.Day()
	hour := timestamp.Hour()
	minute := timestamp.Minute()
	second := timestamp.Second()

	return datetime.Make(year, month, day, hour, minute, second)
}

// Now converts the current Time timestamp to a NEX DateTime
func (datetime *DateTime) Now() uint64 {
	return datetime.FromTimestamp(time.Now())
}

// UTC returns a NEX DateTime value of the current UTC time
func (datetime *DateTime) UTC() uint64 {
	return datetime.FromTimestamp(time.Now().UTC())
}

// Value returns the stored DateTime time
func (datetime *DateTime) Value() uint64 {
	return datetime.value
}

// Second returns the seconds value stored in the DateTime
func (datetime *DateTime) Second() int {
	return int(datetime.value & 63)
}

// Minute returns the minutes value stored in the DateTime
func (datetime *DateTime) Minute() int {
	return int((datetime.value >> 6) & 63)
}

// Hour returns the hours value stored in the DateTime
func (datetime *DateTime) Hour() int {
	return int((datetime.value >> 12) & 31)
}

// Day returns the day value stored in the DateTime
func (datetime *DateTime) Day() int {
	return int((datetime.value >> 17) & 31)
}

// Month returns the month value stored in the DateTime
func (datetime *DateTime) Month() time.Month {
	return time.Month((datetime.value >> 22) & 15)
}

// Year returns the year value stored in the DateTime
func (datetime *DateTime) Year() int {
	return int(datetime.value >> 26)
}

// Standard returns the DateTime as a standard time.Time
func (datetime *DateTime) Standard() time.Time {
	return time.Date(
		datetime.Year(),
		datetime.Month(),
		datetime.Day(),
		datetime.Hour(),
		datetime.Minute(),
		datetime.Second(),
		0,
		time.UTC,
	)
}

// NewDateTime returns a new DateTime instance
func NewDateTime(value uint64) *DateTime {
	return &DateTime{value: value}
}

// StationURL contains the data for a NEX station URL
type StationURL struct {
	// Using pointers to check for nil
	scheme        string
	address       string
	port          string
	stream        string
	sid           string
	cid           string
	pid           string
	transportType string
	rvcid         string
	natm          string
	natf          string
	upnp          string
	pmp           string
	probeinit     string
	prid          string
}

// SetScheme sets the StationURL scheme
func (station *StationURL) SetScheme(scheme string) {
	station.scheme = scheme
}

// SetAddress sets the StationURL address
func (station *StationURL) SetAddress(address string) {
	station.address = address
}

// SetPort sets the StationURL port
func (station *StationURL) SetPort(port string) {
	station.port = port
}

// SetStream sets the StationURL stream
func (station *StationURL) SetStream(stream string) {
	station.stream = stream
}

// SetSID sets the StationURL SID
func (station *StationURL) SetSID(sid string) {
	station.sid = sid
}

// SetCID sets the StationURL CID
func (station *StationURL) SetCID(cid string) {
	station.cid = cid
}

// SetPID sets the StationURL PID
func (station *StationURL) SetPID(pid string) {
	station.pid = pid
}

// SetType sets the StationURL transportType
func (station *StationURL) SetType(transportType string) {
	station.transportType = transportType
}

// SetRVCID sets the StationURL RVCID
func (station *StationURL) SetRVCID(rvcid string) {
	station.rvcid = rvcid
}

// SetNatm sets the StationURL Natm
func (station *StationURL) SetNatm(natm string) {
	station.natm = natm
}

// SetNatf sets the StationURL Natf
func (station *StationURL) SetNatf(natf string) {
	station.natf = natf
}

// SetUpnp sets the StationURL Upnp
func (station *StationURL) SetUpnp(upnp string) {
	station.upnp = upnp
}

// SetPmp sets the StationURL Pmp
func (station *StationURL) SetPmp(pmp string) {
	station.pmp = pmp
}

// SetProbeInit sets the StationURL ProbeInit
func (station *StationURL) SetProbeInit(probeinit string) {
	station.probeinit = probeinit
}

// SetPRID sets the StationURL PRID
func (station *StationURL) SetPRID(prid string) {
	station.prid = prid
}

// Scheme returns the StationURL scheme type
func (station *StationURL) Scheme() string {
	return station.address
}

// Address returns the StationURL address
func (station *StationURL) Address() string {
	return station.address
}

// Port returns the StationURL port
func (station *StationURL) Port() string {
	return station.port
}

// Stream returns the StationURL stream value
func (station *StationURL) Stream() string {
	return station.stream
}

// SID returns the StationURL SID value
func (station *StationURL) SID() string {
	return station.sid
}

// CID returns the StationURL CID value
func (station *StationURL) CID() string {
	return station.cid
}

// PID returns the StationURL PID value
func (station *StationURL) PID() string {
	return station.pid
}

// Type returns the StationURL type
func (station *StationURL) Type() string {
	return station.transportType
}

// RVCID returns the StationURL RVCID
func (station *StationURL) RVCID() string {
	return station.rvcid
}

// Natm returns the StationURL Natm value
func (station *StationURL) Natm() string {
	return station.natm
}

// Natf returns the StationURL Natf value
func (station *StationURL) Natf() string {
	return station.natf
}

// Upnp returns the StationURL Upnp value
func (station *StationURL) Upnp() string {
	return station.upnp
}

// Pmp returns the StationURL Pmp value
func (station *StationURL) Pmp() string {
	return station.pmp
}

// ProbeInit returns the StationURL ProbeInit value
func (station *StationURL) ProbeInit() string {
	return station.probeinit
}

// PRID returns the StationURL PRID value
func (station *StationURL) PRID() string {
	return station.prid
}

// FromString parses the StationURL data from a string
func (station *StationURL) FromString(str string) {
	split := strings.Split(str, ":/")

	station.scheme = split[0]
	fields := split[1]

	params := strings.Split(fields, ";")

	for i := 0; i < len(params); i++ {
		param := params[i]
		split = strings.Split(param, "=")

		name := split[0]
		value := split[1]

		switch name {
		case "address":
			station.address = value
		case "port":
			station.port = value
		case "stream":
			station.stream = value
		case "sid":
			station.sid = value
		case "CID":
			station.cid = value
		case "PID":
			station.pid = value
		case "type":
			station.transportType = value
		case "RVCID":
			station.rvcid = value
		case "natm":
			station.natm = value
		case "natf":
			station.natf = value
		case "upnp":
			station.upnp = value
		case "pmp":
			station.pmp = value
		case "probeinit":
			station.probeinit = value
		case "PRID":
			station.prid = value
		}
	}
}

// EncodeToString encodes the StationURL into a string
func (station *StationURL) EncodeToString() string {
	fields := []string{}

	if station.address != "" {
		fields = append(fields, "address="+station.address)
	}

	if station.port != "" {
		fields = append(fields, "port="+station.port)
	}

	if station.stream != "" {
		fields = append(fields, "stream="+station.stream)
	}

	if station.sid != "" {
		fields = append(fields, "sid="+station.sid)
	}

	if station.cid != "" {
		fields = append(fields, "CID="+station.cid)
	}

	if station.pid != "" {
		fields = append(fields, "PID="+station.pid)
	}

	if station.transportType != "" {
		fields = append(fields, "type="+station.transportType)
	}

	if station.rvcid != "" {
		fields = append(fields, "RVCID="+station.rvcid)
	}

	if station.natm != "" {
		fields = append(fields, "natm="+station.natm)
	}

	if station.natf != "" {
		fields = append(fields, "natf="+station.natf)
	}

	if station.upnp != "" {
		fields = append(fields, "upnp="+station.upnp)
	}

	if station.pmp != "" {
		fields = append(fields, "pmp="+station.pmp)
	}

	if station.probeinit != "" {
		fields = append(fields, "probeinit="+station.probeinit)
	}

	if station.prid != "" {
		fields = append(fields, "PRID="+station.prid)
	}

	return station.scheme + ":/" + strings.Join(fields, ";")
}

// NewStationURL returns a new StationURL
func NewStationURL(str string) *StationURL {
	station := &StationURL{}

	if str != "" {
		station.FromString(str)
	}

	return station
}

// Result is sent in methods which query large objects
type Result struct {
	code uint32
}

// IsSuccess returns true if the Result is a success
func (result *Result) IsSuccess() bool {
	return int(result.code)&errorMask == 0
}

// IsError returns true if the Result is a error
func (result *Result) IsError() bool {
	return int(result.code)&errorMask != 0
}

// ExtractFromStream extracts a Result structure from a stream
func (result *Result) ExtractFromStream(stream *StreamIn) error {
	result.code = stream.ReadUInt32LE()

	return nil
}

// Bytes encodes the Result and returns a byte array
func (result *Result) Bytes(stream *StreamOut) []byte {
	stream.WriteUInt32LE(result.code)

	return stream.Bytes()
}

// NewResult returns a new Result
func NewResult(code uint32) *Result {
	return &Result{code}
}

// NewResultSuccess returns a new Result set as a success
func NewResultSuccess(code uint32) *Result {
	return NewResult(uint32(int(code) & ^errorMask))
}

// NewResultError returns a new Result set as an error
func NewResultError(code uint32) *Result {
	return NewResult(uint32(int(code) | errorMask))
}

// ResultRange is sent in methods which query large objects
type ResultRange struct {
	Offset uint32
	Length uint32
	Structure
}

// ExtractFromStream extracts a ResultRange structure from a stream
func (resultRange *ResultRange) ExtractFromStream(stream *StreamIn) error {
	resultRange.Offset = stream.ReadUInt32LE()
	resultRange.Length = stream.ReadUInt32LE()

	return nil
}

// NewResultRange returns a new ResultRange
func NewResultRange() *ResultRange {
	return &ResultRange{}
}
