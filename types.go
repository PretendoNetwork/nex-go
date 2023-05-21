package nex

import (
	"bytes"
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
	Copy() StructureInterface
	Equals(StructureInterface) bool
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

// Copy returns a new copied instance of Data
func (data *Data) Copy() StructureInterface {
	return NewData() // * Has no fields, nothing to copy
}

// Equals checks if the passed Structure contains the same data as the current instance
func (data *Data) Equals(structure StructureInterface) bool {
	return true // * Has no fields, always equal
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

// Copy returns a new copied instance of DataHolder
func (dataHolder *DataHolder) Copy() *DataHolder {
	copied := NewDataHolder()

	copied.typeName = dataHolder.typeName
	copied.length1 = dataHolder.length1
	copied.length2 = dataHolder.length2
	copied.objectData = dataHolder.objectData.Copy()

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (dataHolder *DataHolder) Equals(other *DataHolder) bool {
	if dataHolder.typeName != other.typeName {
		return false
	}

	if dataHolder.length1 != other.length1 {
		return false
	}

	if dataHolder.length2 != other.length2 {
		return false
	}

	if !dataHolder.objectData.Equals(other.objectData) {
		return false
	}

	return true
}

// NewDataHolder returns a new DataHolder
func NewDataHolder() *DataHolder {
	return &DataHolder{}
}

// RVConnectionData represents a nex RVConnectionData type
type RVConnectionData struct {
	Structure
	stationURL                 string
	specialProtocols           []byte
	stationURLSpecialProtocols string
	time                       uint64
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

// Copy returns a new copied instance of RVConnectionData
func (rvConnectionData *RVConnectionData) Copy() StructureInterface {
	copied := NewRVConnectionData()

	copied.parentType = rvConnectionData.parentType
	copied.stationURL = rvConnectionData.stationURL
	copied.specialProtocols = make([]byte, len(rvConnectionData.specialProtocols))

	copy(copied.specialProtocols, rvConnectionData.specialProtocols)

	copied.stationURLSpecialProtocols = rvConnectionData.stationURLSpecialProtocols
	copied.time = rvConnectionData.time

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (rvConnectionData *RVConnectionData) Equals(structure StructureInterface) bool {
	other := structure.(*RVConnectionData)

	if rvConnectionData.stationURL != other.stationURL {
		return false
	}

	if !bytes.Equal(rvConnectionData.specialProtocols, other.specialProtocols) {
		return false
	}

	if rvConnectionData.stationURLSpecialProtocols != other.stationURLSpecialProtocols {
		return false
	}

	if rvConnectionData.time != other.time {
		return false
	}

	return true
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

// Copy returns a new copied instance of DateTime
func (datetime *DateTime) Copy() *DateTime {
	return NewDateTime(datetime.value)
}

// Equals checks if the passed Structure contains the same data as the current instance
func (datetime *DateTime) Equals(other *DateTime) bool {
	return datetime.value == other.value
}

// NewDateTime returns a new DateTime instance
func NewDateTime(value uint64) *DateTime {
	return &DateTime{value: value}
}

// StationURL contains the data for a NEX station URL
type StationURL struct {
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
func (stationURL *StationURL) SetScheme(scheme string) {
	stationURL.scheme = scheme
}

// SetAddress sets the StationURL address
func (stationURL *StationURL) SetAddress(address string) {
	stationURL.address = address
}

// SetPort sets the StationURL port
func (stationURL *StationURL) SetPort(port string) {
	stationURL.port = port
}

// SetStream sets the StationURL stream
func (stationURLstationURL *StationURL) SetStream(stream string) {
	stationURLstationURL.stream = stream
}

// SetSID sets the StationURL SID
func (stationURLstationURL *StationURL) SetSID(sid string) {
	stationURLstationURL.sid = sid
}

// SetCID sets the StationURL CID
func (stationURLstationURL *StationURL) SetCID(cid string) {
	stationURLstationURL.cid = cid
}

// SetPID sets the StationURL PID
func (stationURLstationURL *StationURL) SetPID(pid string) {
	stationURLstationURL.pid = pid
}

// SetType sets the StationURL transportType
func (stationURLstationURL *StationURL) SetType(transportType string) {
	stationURLstationURL.transportType = transportType
}

// SetRVCID sets the StationURL RVCID
func (stationURLstationURL *StationURL) SetRVCID(rvcid string) {
	stationURLstationURL.rvcid = rvcid
}

// SetNatm sets the StationURL Natm
func (stationURLstationURL *StationURL) SetNatm(natm string) {
	stationURLstationURL.natm = natm
}

// SetNatf sets the StationURL Natf
func (stationURLstationURL *StationURL) SetNatf(natf string) {
	stationURLstationURL.natf = natf
}

// SetUpnp sets the StationURL Upnp
func (stationURLstationURL *StationURL) SetUpnp(upnp string) {
	stationURLstationURL.upnp = upnp
}

// SetPmp sets the StationURL Pmp
func (stationURLstationURL *StationURL) SetPmp(pmp string) {
	stationURLstationURL.pmp = pmp
}

// SetProbeInit sets the StationURL ProbeInit
func (stationURLstationURL *StationURL) SetProbeInit(probeinit string) {
	stationURLstationURL.probeinit = probeinit
}

// SetPRID sets the StationURL PRID
func (stationURLstationURL *StationURL) SetPRID(prid string) {
	stationURLstationURL.prid = prid
}

// Scheme returns the StationURL scheme type
func (stationURLstationURL *StationURL) Scheme() string {
	return stationURLstationURL.address
}

// Address returns the StationURL address
func (stationURLstationURL *StationURL) Address() string {
	return stationURLstationURL.address
}

// Port returns the StationURL port
func (stationURLstationURL *StationURL) Port() string {
	return stationURLstationURL.port
}

// Stream returns the StationURL stream value
func (stationURLstationURL *StationURL) Stream() string {
	return stationURLstationURL.stream
}

// SID returns the StationURL SID value
func (stationURLstationURL *StationURL) SID() string {
	return stationURLstationURL.sid
}

// CID returns the StationURL CID value
func (stationURLstationURL *StationURL) CID() string {
	return stationURLstationURL.cid
}

// PID returns the StationURL PID value
func (stationURLstationURL *StationURL) PID() string {
	return stationURLstationURL.pid
}

// Type returns the StationURL type
func (stationURLstationURL *StationURL) Type() string {
	return stationURLstationURL.transportType
}

// RVCID returns the StationURL RVCID
func (stationURLstationURL *StationURL) RVCID() string {
	return stationURLstationURL.rvcid
}

// Natm returns the StationURL Natm value
func (stationURLstationURL *StationURL) Natm() string {
	return stationURLstationURL.natm
}

// Natf returns the StationURL Natf value
func (stationURL *StationURL) Natf() string {
	return stationURL.natf
}

// Upnp returns the StationURL Upnp value
func (stationURL *StationURL) Upnp() string {
	return stationURL.upnp
}

// Pmp returns the StationURL Pmp value
func (stationURL *StationURL) Pmp() string {
	return stationURL.pmp
}

// ProbeInit returns the StationURL ProbeInit value
func (stationURL *StationURL) ProbeInit() string {
	return stationURL.probeinit
}

// PRID returns the StationURL PRID value
func (stationURL *StationURL) PRID() string {
	return stationURL.prid
}

// FromString parses the StationURL data from a string
func (stationURL *StationURL) FromString(str string) {
	split := strings.Split(str, ":/")

	stationURL.scheme = split[0]
	fields := split[1]

	params := strings.Split(fields, ";")

	for i := 0; i < len(params); i++ {
		param := params[i]
		split = strings.Split(param, "=")

		name := split[0]
		value := split[1]

		switch name {
		case "address":
			stationURL.address = value
		case "port":
			stationURL.port = value
		case "stream":
			stationURL.stream = value
		case "sid":
			stationURL.sid = value
		case "CID":
			stationURL.cid = value
		case "PID":
			stationURL.pid = value
		case "type":
			stationURL.transportType = value
		case "RVCID":
			stationURL.rvcid = value
		case "natm":
			stationURL.natm = value
		case "natf":
			stationURL.natf = value
		case "upnp":
			stationURL.upnp = value
		case "pmp":
			stationURL.pmp = value
		case "probeinit":
			stationURL.probeinit = value
		case "PRID":
			stationURL.prid = value
		}
	}
}

// EncodeToString encodes the StationURL into a string
func (stationURL *StationURL) EncodeToString() string {
	fields := []string{}

	if stationURL.address != "" {
		fields = append(fields, "address="+stationURL.address)
	}

	if stationURL.port != "" {
		fields = append(fields, "port="+stationURL.port)
	}

	if stationURL.stream != "" {
		fields = append(fields, "stream="+stationURL.stream)
	}

	if stationURL.sid != "" {
		fields = append(fields, "sid="+stationURL.sid)
	}

	if stationURL.cid != "" {
		fields = append(fields, "CID="+stationURL.cid)
	}

	if stationURL.pid != "" {
		fields = append(fields, "PID="+stationURL.pid)
	}

	if stationURL.transportType != "" {
		fields = append(fields, "type="+stationURL.transportType)
	}

	if stationURL.rvcid != "" {
		fields = append(fields, "RVCID="+stationURL.rvcid)
	}

	if stationURL.natm != "" {
		fields = append(fields, "natm="+stationURL.natm)
	}

	if stationURL.natf != "" {
		fields = append(fields, "natf="+stationURL.natf)
	}

	if stationURL.upnp != "" {
		fields = append(fields, "upnp="+stationURL.upnp)
	}

	if stationURL.pmp != "" {
		fields = append(fields, "pmp="+stationURL.pmp)
	}

	if stationURL.probeinit != "" {
		fields = append(fields, "probeinit="+stationURL.probeinit)
	}

	if stationURL.prid != "" {
		fields = append(fields, "PRID="+stationURL.prid)
	}

	return stationURL.scheme + ":/" + strings.Join(fields, ";")
}

// Copy returns a new copied instance of StationURL
func (stationURL *StationURL) Copy() *StationURL {
	return NewStationURL(stationURL.EncodeToString())
}

// Equals checks if the passed Structure contains the same data as the current instance
func (stationURL *StationURL) Equals(other *StationURL) bool {
	return stationURL.EncodeToString() == other.EncodeToString()
}

// NewStationURL returns a new StationURL
func NewStationURL(str string) *StationURL {
	stationURL := &StationURL{}

	if str != "" {
		stationURL.FromString(str)
	}

	return stationURL
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

// Copy returns a new copied instance of Result
func (result *Result) Copy() *Result {
	return NewResult(result.code)
}

// Equals checks if the passed Structure contains the same data as the current instance
func (result *Result) Equals(other *Result) bool {
	return result.code == other.code
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
	Structure
	Offset uint32
	Length uint32
}

// ExtractFromStream extracts a ResultRange structure from a stream
func (resultRange *ResultRange) ExtractFromStream(stream *StreamIn) error {
	resultRange.Offset = stream.ReadUInt32LE()
	resultRange.Length = stream.ReadUInt32LE()

	return nil
}

// Copy returns a new copied instance of RVConnectionData
func (resultRange *ResultRange) Copy() StructureInterface {
	copied := NewResultRange()

	copied.Offset = resultRange.Offset
	copied.Length = resultRange.Length

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (resultRange *ResultRange) Equals(structure StructureInterface) bool {
	other := structure.(*ResultRange)

	if resultRange.Offset != other.Offset {
		return false
	}

	if resultRange.Length != other.Length {
		return false
	}

	return true
}

// NewResultRange returns a new ResultRange
func NewResultRange() *ResultRange {
	return &ResultRange{}
}
