package nex

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
)

// StructureInterface implements all Structure methods
type StructureInterface interface {
	SetParentType(StructureInterface)
	ParentType() StructureInterface
	SetStructureVersion(uint8)
	StructureVersion() uint8
	ExtractFromStream(*StreamIn) error
	Bytes(*StreamOut) []byte
	Copy() StructureInterface
	Equals(StructureInterface) bool
	FormatToString(int) string
}

// Structure represents a nex Structure type
type Structure struct {
	parentType       StructureInterface
	structureVersion uint8
	StructureInterface
}

// SetParentType sets the Structures parent type
func (structure *Structure) SetParentType(parentType StructureInterface) {
	structure.parentType = parentType
}

// ParentType returns the Structures parent type. nil if the type does not inherit another Structure
func (structure *Structure) ParentType() StructureInterface {
	return structure.parentType
}

// SetStructureVersion sets the structures version. Only used in NEX 3.5+
func (structure *Structure) SetStructureVersion(version uint8) {
	structure.structureVersion = version
}

// StructureVersion returns the structures version. Only used in NEX 3.5+
func (structure *Structure) StructureVersion() uint8 {
	return structure.structureVersion
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

// String returns a string representation of the struct
func (data *Data) String() string {
	return data.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (data *Data) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Data{\n")
	b.WriteString(fmt.Sprintf("%sstructureVersion: %d\n", indentationValues, data.structureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewData returns a new Data Structure
func NewData() *Data {
	return &Data{}
}

var dataHolderKnownObjects = make(map[string]StructureInterface)

// RegisterDataHolderType registers a structure to be a valid type in the DataHolder structure
func RegisterDataHolderType(name string, structure StructureInterface) {
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
	var err error

	dataHolder.typeName, err = stream.ReadString()
	if err != nil {
		return fmt.Errorf("Failed to read DataHolder type name. %s", err.Error())
	}

	dataHolder.length1, err = stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read DataHolder length 1. %s", err.Error())
	}

	dataHolder.length2, err = stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read DataHolder length 2. %s", err.Error())
	}

	dataType := dataHolderKnownObjects[dataHolder.typeName]
	if dataType == nil {
		// TODO - Should we really log this here, or just pass the error to the caller?
		message := fmt.Sprintf("UNKNOWN DATAHOLDER TYPE: %s", dataHolder.typeName)
		logger.Critical(message)
		return errors.New(message)
	}

	newObjectInstance := dataType.Copy()

	dataHolder.objectData, err = stream.ReadStructure(newObjectInstance)
	if err != nil {
		return fmt.Errorf("Failed to read DataHolder object data. %s", err.Error())
	}

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

// String returns a string representation of the struct
func (dataHolder *DataHolder) String() string {
	return dataHolder.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (dataHolder *DataHolder) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("DataHolder{\n")
	b.WriteString(fmt.Sprintf("%stypeName: %s,\n", indentationValues, dataHolder.typeName))
	b.WriteString(fmt.Sprintf("%slength1: %d,\n", indentationValues, dataHolder.length1))
	b.WriteString(fmt.Sprintf("%slength2: %d,\n", indentationValues, dataHolder.length2))
	b.WriteString(fmt.Sprintf("%sobjectData: %s\n", indentationValues, dataHolder.objectData.FormatToString(indentationLevel+1)))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
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
	time                       *DateTime
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
func (rvConnectionData *RVConnectionData) SetTime(time *DateTime) {
	rvConnectionData.time = time
}

// Bytes encodes the RVConnectionData and returns a byte array
func (rvConnectionData *RVConnectionData) Bytes(stream *StreamOut) []byte {
	nexVersion := stream.Server.NEXVersion()

	stream.WriteString(rvConnectionData.stationURL)
	stream.WriteListUInt8(rvConnectionData.specialProtocols)
	stream.WriteString(rvConnectionData.stationURLSpecialProtocols)

	if nexVersion.Major >= 3 && nexVersion.Minor >= 5 {
		rvConnectionData.SetStructureVersion(1)
		stream.WriteDateTime(rvConnectionData.time)
	}

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

	if rvConnectionData.time != nil {
		copied.time = rvConnectionData.time.Copy()
	}

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

	if rvConnectionData.time != nil && other.time == nil {
		return false
	}

	if rvConnectionData.time == nil && other.time != nil {
		return false
	}

	if rvConnectionData.time != nil && other.time != nil {
		if !rvConnectionData.time.Equals(other.time) {
			return false
		}
	}

	return true
}

// String returns a string representation of the struct
func (rvConnectionData *RVConnectionData) String() string {
	return rvConnectionData.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (rvConnectionData *RVConnectionData) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("RVConnectionData{\n")
	b.WriteString(fmt.Sprintf("%sstructureVersion: %d,\n", indentationValues, rvConnectionData.structureVersion))
	b.WriteString(fmt.Sprintf("%sstationURL: %q,\n", indentationValues, rvConnectionData.stationURL))
	b.WriteString(fmt.Sprintf("%sspecialProtocols: %v,\n", indentationValues, rvConnectionData.specialProtocols))
	b.WriteString(fmt.Sprintf("%sstationURLSpecialProtocols: %q,\n", indentationValues, rvConnectionData.stationURLSpecialProtocols))

	if rvConnectionData.time != nil {
		b.WriteString(fmt.Sprintf("%stime: %s\n", indentationValues, rvConnectionData.time.FormatToString(indentationLevel+1)))
	} else {
		b.WriteString(fmt.Sprintf("%stime: nil\n", indentationValues))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
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

// String returns a string representation of the struct
func (datetime *DateTime) String() string {
	return datetime.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (datetime *DateTime) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("DateTime{\n")
	b.WriteString(fmt.Sprintf("%svalue: %d (%s)\n", indentationValues, datetime.value, datetime.Standard().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
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
func (stationURL *StationURL) SetStream(stream string) {
	stationURL.stream = stream
}

// SetSID sets the StationURL SID
func (stationURL *StationURL) SetSID(sid string) {
	stationURL.sid = sid
}

// SetCID sets the StationURL CID
func (stationURL *StationURL) SetCID(cid string) {
	stationURL.cid = cid
}

// SetPID sets the StationURL PID
func (stationURL *StationURL) SetPID(pid string) {
	stationURL.pid = pid
}

// SetType sets the StationURL transportType
func (stationURL *StationURL) SetType(transportType string) {
	stationURL.transportType = transportType
}

// SetRVCID sets the StationURL RVCID
func (stationURL *StationURL) SetRVCID(rvcid string) {
	stationURL.rvcid = rvcid
}

// SetNatm sets the StationURL Natm
func (stationURL *StationURL) SetNatm(natm string) {
	stationURL.natm = natm
}

// SetNatf sets the StationURL Natf
func (stationURL *StationURL) SetNatf(natf string) {
	stationURL.natf = natf
}

// SetUpnp sets the StationURL Upnp
func (stationURL *StationURL) SetUpnp(upnp string) {
	stationURL.upnp = upnp
}

// SetPmp sets the StationURL Pmp
func (stationURL *StationURL) SetPmp(pmp string) {
	stationURL.pmp = pmp
}

// SetProbeInit sets the StationURL ProbeInit
func (stationURL *StationURL) SetProbeInit(probeinit string) {
	stationURL.probeinit = probeinit
}

// SetPRID sets the StationURL PRID
func (stationURL *StationURL) SetPRID(prid string) {
	stationURL.prid = prid
}

// Scheme returns the StationURL scheme type
func (stationURL *StationURL) Scheme() string {
	return stationURL.address
}

// Address returns the StationURL address
func (stationURL *StationURL) Address() string {
	return stationURL.address
}

// Port returns the StationURL port
func (stationURL *StationURL) Port() string {
	return stationURL.port
}

// Stream returns the StationURL stream value
func (stationURL *StationURL) Stream() string {
	return stationURL.stream
}

// SID returns the StationURL SID value
func (stationURL *StationURL) SID() string {
	return stationURL.sid
}

// CID returns the StationURL CID value
func (stationURL *StationURL) CID() string {
	return stationURL.cid
}

// PID returns the StationURL PID value
func (stationURL *StationURL) PID() string {
	return stationURL.pid
}

// Type returns the StationURL type
func (stationURL *StationURL) Type() string {
	return stationURL.transportType
}

// RVCID returns the StationURL RVCID
func (stationURL *StationURL) RVCID() string {
	return stationURL.rvcid
}

// Natm returns the StationURL Natm value
func (stationURL *StationURL) Natm() string {
	return stationURL.natm
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

// String returns a string representation of the struct
func (stationURL *StationURL) String() string {
	return stationURL.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (stationURL *StationURL) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("StationURL{\n")
	b.WriteString(fmt.Sprintf("%surl: %q\n", indentationValues, stationURL.EncodeToString()))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
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
	code, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read Result code. %s", err.Error())
	}

	result.code = code

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

// String returns a string representation of the struct
func (result *Result) String() string {
	return result.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (result *Result) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Result{\n")

	if result.IsSuccess() {
		b.WriteString(fmt.Sprintf("%scode: %d (success)\n", indentationValues, result.code))
	} else {
		b.WriteString(fmt.Sprintf("%scode: %d (error)\n", indentationValues, result.code))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
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
	offset, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read ResultRange offset. %s", err.Error())
	}

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read ResultRange length. %s", err.Error())
	}

	resultRange.Offset = offset
	resultRange.Length = length

	return nil
}

// Copy returns a new copied instance of ResultRange
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

// String returns a string representation of the struct
func (resultRange *ResultRange) String() string {
	return resultRange.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (resultRange *ResultRange) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("ResultRange{\n")
	b.WriteString(fmt.Sprintf("%sstructureVersion: %d,\n", indentationValues, resultRange.structureVersion))
	b.WriteString(fmt.Sprintf("%sOffset: %d,\n", indentationValues, resultRange.Offset))
	b.WriteString(fmt.Sprintf("%sLength: %d\n", indentationValues, resultRange.Length))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewResultRange returns a new ResultRange
func NewResultRange() *ResultRange {
	return &ResultRange{}
}

// Variant can hold one of 7 types; nil, int64, float64, bool, string, DateTime, or uint64
type Variant struct {
	TypeID uint8
	// * In reality this type does not have this many fields
	// * It only stores the type ID and then the value
	// * However to get better typing, we opt to store each possible
	// * type as it's own field and just check typeID to know which it has
	Int64    int64
	Float64  float64
	Bool     bool
	Str      string
	DateTime *DateTime
	UInt64   uint64
}

// ExtractFromStream extracts a Variant structure from a stream
func (variant *Variant) ExtractFromStream(stream *StreamIn) error {
	var err error

	variant.TypeID, err = stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read Variant type ID. %s", err.Error())
	}

	// * A type ID of 0 means no value
	switch variant.TypeID {
	case 1: // * sint64
		variant.Int64, err = stream.ReadInt64LE()
	case 2: // * double
		variant.Float64, err = stream.ReadFloat64LE()
	case 3: // * bool
		variant.Bool, err = stream.ReadBool()
	case 4: // * string
		variant.Str, err = stream.ReadString()
	case 5: // * datetime
		variant.DateTime, err = stream.ReadDateTime()
	case 6: // * uint64
		variant.UInt64, err = stream.ReadUInt64LE()
	}

	// * These errors contain details about each of the values type
	// * No need to return special errors for each value type
	if err != nil {
		return fmt.Errorf("Failed to read Variant value. %s", err.Error())
	}

	return nil
}

// Bytes encodes the Variant and returns a byte array
func (variant *Variant) Bytes(stream *StreamOut) []byte {
	stream.WriteUInt8(variant.TypeID)

	// * A type ID of 0 means no value
	switch variant.TypeID {
	case 1: // * sint64
		stream.WriteInt64LE(variant.Int64)
	case 2: // * double
		stream.WriteFloat64LE(variant.Float64)
	case 3: // * bool
		stream.WriteBool(variant.Bool)
	case 4: // * string
		stream.WriteString(variant.Str)
	case 5: // * datetime
		stream.WriteDateTime(variant.DateTime)
	case 6: // * uint64
		stream.WriteUInt64LE(variant.UInt64)
	}

	return stream.Bytes()
}

// Copy returns a new copied instance of Variant
func (variant *Variant) Copy() *Variant {
	copied := NewVariant()

	copied.TypeID = variant.TypeID
	copied.Int64 = variant.Int64
	copied.Float64 = variant.Float64
	copied.Bool = variant.Bool
	copied.Str = variant.Str

	if variant.DateTime != nil {
		copied.DateTime = variant.DateTime.Copy()
	}

	copied.UInt64 = variant.UInt64

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (variant *Variant) Equals(other *Variant) bool {
	if variant.TypeID != other.TypeID {
		return false
	}

	// * A type ID of 0 means no value
	switch variant.TypeID {
	case 0: // * no value, always equal
		return true
	case 1: // * sint64
		return variant.Int64 == other.Int64
	case 2: // * double
		return variant.Float64 == other.Float64
	case 3: // * bool
		return variant.Bool == other.Bool
	case 4: // * string
		return variant.Str == other.Str
	case 5: // * datetime
		return variant.DateTime.Equals(other.DateTime)
	case 6: // * uint64
		return variant.UInt64 == other.UInt64
	default: // * Something went horribly wrong
		return false
	}
}

// String returns a string representation of the struct
func (variant *Variant) String() string {
	return variant.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (variant *Variant) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Variant{\n")
	b.WriteString(fmt.Sprintf("%sTypeID: %d\n", indentationValues, variant.TypeID))

	switch variant.TypeID {
	case 0: // * no value
		b.WriteString(fmt.Sprintf("%svalue: nil\n", indentationValues))
	case 1: // * sint64
		b.WriteString(fmt.Sprintf("%svalue: %d\n", indentationValues, variant.Int64))
	case 2: // * double
		b.WriteString(fmt.Sprintf("%svalue: %g\n", indentationValues, variant.Float64))
	case 3: // * bool
		b.WriteString(fmt.Sprintf("%svalue: %t\n", indentationValues, variant.Bool))
	case 4: // * string
		b.WriteString(fmt.Sprintf("%svalue: %q\n", indentationValues, variant.Str))
	case 5: // * datetime
		b.WriteString(fmt.Sprintf("%svalue: %s\n", indentationValues, variant.DateTime.FormatToString(indentationLevel+1)))
	case 6: // * uint64
		b.WriteString(fmt.Sprintf("%svalue: %d\n", indentationValues, variant.UInt64))
	default:
		b.WriteString(fmt.Sprintf("%svalue: Unknown\n", indentationValues))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewVariant returns a new Variant
func NewVariant() *Variant {
	return &Variant{}
}
