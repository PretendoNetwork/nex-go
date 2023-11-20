package nex

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PID represents a unique number to identify a user
//
// The true size of this value depends on the client version.
// Legacy clients (WiiU/3DS) use a uint32, whereas new clients (Nintendo Switch) use a uint64.
// Value is always stored as the higher uint64, the consuming API should assert accordingly
type PID struct {
	pid uint64
}

// Value returns the numeric value of the PID as a uint64 regardless of client version
func (p *PID) Value() uint64 {
	return p.pid
}

// LegacyValue returns the numeric value of the PID as a uint32, for legacy clients
func (p *PID) LegacyValue() uint32 {
	return uint32(p.pid)
}

// Equals checks if the two structs are equal
func (p *PID) Equals(other *PID) bool {
	return p.pid == other.pid
}

// Copy returns a copy of the current PID
func (p *PID) Copy() *PID {
	return NewPID(p.pid)
}

// String returns a string representation of the struct
func (p *PID) String() string {
	return p.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (p *PID) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("PID{\n")

	switch v := any(p.pid).(type) {
	case uint32:
		b.WriteString(fmt.Sprintf("%spid: %d (legacy)\n", indentationValues, v))
	case uint64:
		b.WriteString(fmt.Sprintf("%spid: %d (modern)\n", indentationValues, v))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewPID returns a PID instance. The size of PID depends on the client version
func NewPID[T uint32 | uint64](pid T) *PID {
	switch v := any(pid).(type) {
	case uint32:
		return &PID{pid: uint64(v)}
	case uint64:
		return &PID{pid: v}
	}

	// * This will never happen because Go will
	// * not compile any code where "pid" is not
	// * a uint32/uint64, so it will ALWAYS get
	// * caught by the above switch-case. This
	// * return is only here because Go won't
	// * compile without a default return
	return nil
}

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
	copied := NewData()

	copied.SetStructureVersion(data.StructureVersion())

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (data *Data) Equals(structure StructureInterface) bool {
	return data.StructureVersion() == structure.StructureVersion()
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
		return errors.New(message)
	}

	newObjectInstance := dataType.Copy()

	dataHolder.objectData, err = StreamReadStructure(stream, newObjectInstance)
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
	stream.WriteString(rvConnectionData.stationURL)
	stream.WriteListUInt8(rvConnectionData.specialProtocols)
	stream.WriteString(rvConnectionData.stationURLSpecialProtocols)

	if stream.Server.LibraryVersion().GreaterOrEqual("3.5.0") {
		rvConnectionData.SetStructureVersion(1)
		stream.WriteDateTime(rvConnectionData.time)
	}

	return stream.Bytes()
}

// Copy returns a new copied instance of RVConnectionData
func (rvConnectionData *RVConnectionData) Copy() StructureInterface {
	copied := NewRVConnectionData()

	copied.SetStructureVersion(rvConnectionData.StructureVersion())
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

	if rvConnectionData.StructureVersion() == other.StructureVersion() {
		return false
	}

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

// Copy returns a new copied instance of DateTime
func (dt *DateTime) Copy() *DateTime {
	return NewDateTime(dt.value)
}

// Equals checks if the passed Structure contains the same data as the current instance
func (dt *DateTime) Equals(other *DateTime) bool {
	return dt.value == other.value
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

// StationURL contains the data for a NEX station URL.
// Uses pointers to check for nil, 0 is valid
type StationURL struct {
	local         bool // * Not part of the data structure. Used for easier lookups elsewhere
	public        bool // * Not part of the data structure. Used for easier lookups elsewhere
	scheme        string
	address       string
	port          *uint32
	pl            *uint32 // * Seen in Minecraft
	stream        *uint32
	sid           *uint32
	cid           *uint32
	pid           *PID
	transportType *uint32
	rvcid         *uint32
	natm          *uint32
	natf          *uint32
	upnp          *uint32
	pmp           *uint32
	probeinit     *uint32
	prid          *uint32
}

// SetLocal marks the StationURL as an local URL
func (stationURL *StationURL) SetLocal() {
	stationURL.local = true
	stationURL.public = false
}

// SetPublic marks the StationURL as an public URL
func (stationURL *StationURL) SetPublic() {
	stationURL.local = false
	stationURL.public = true
}

// IsLocal checks if the StationURL is a local URL
func (stationURL *StationURL) IsLocal() bool {
	return stationURL.local
}

// IsPublic checks if the StationURL is a public URL
func (stationURL *StationURL) IsPublic() bool {
	return stationURL.public
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
func (stationURL *StationURL) SetPort(port uint32) {
	stationURL.port = &port
}

// SetPL sets the StationURL Pl
func (stationURL *StationURL) SetPL(pl uint32) {
	stationURL.pl = &pl
}

// SetStream sets the StationURL stream
func (stationURL *StationURL) SetStream(stream uint32) {
	stationURL.stream = &stream
}

// SetSID sets the StationURL SID
func (stationURL *StationURL) SetSID(sid uint32) {
	stationURL.sid = &sid
}

// SetCID sets the StationURL CID
func (stationURL *StationURL) SetCID(cid uint32) {
	stationURL.cid = &cid
}

// SetPID sets the StationURL PID
func (stationURL *StationURL) SetPID(pid *PID) {
	stationURL.pid = pid
}

// SetType sets the StationURL transportType
func (stationURL *StationURL) SetType(transportType uint32) {
	stationURL.transportType = &transportType
}

// SetRVCID sets the StationURL RVCID
func (stationURL *StationURL) SetRVCID(rvcid uint32) {
	stationURL.rvcid = &rvcid
}

// SetNatm sets the StationURL Natm
func (stationURL *StationURL) SetNatm(natm uint32) {
	stationURL.natm = &natm
}

// SetNatf sets the StationURL Natf
func (stationURL *StationURL) SetNatf(natf uint32) {
	stationURL.natf = &natf
}

// SetUpnp sets the StationURL Upnp
func (stationURL *StationURL) SetUpnp(upnp uint32) {
	stationURL.upnp = &upnp
}

// SetPmp sets the StationURL Pmp
func (stationURL *StationURL) SetPmp(pmp uint32) {
	stationURL.pmp = &pmp
}

// SetProbeInit sets the StationURL ProbeInit
func (stationURL *StationURL) SetProbeInit(probeinit uint32) {
	stationURL.probeinit = &probeinit
}

// SetPRID sets the StationURL PRID
func (stationURL *StationURL) SetPRID(prid uint32) {
	stationURL.prid = &prid
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
func (stationURL *StationURL) Port() uint32 {
	if stationURL.port == nil {
		return 0
	} else {
		return *stationURL.port
	}
}

// PL returns the StationURL Pl
func (stationURL *StationURL) PL() uint32 {
	if stationURL.pl == nil {
		return 0
	} else {
		return *stationURL.pl
	}
}

// Stream returns the StationURL stream value
func (stationURL *StationURL) Stream() uint32 {
	if stationURL.stream == nil {
		return 0
	} else {
		return *stationURL.stream
	}
}

// SID returns the StationURL SID value
func (stationURL *StationURL) SID() uint32 {
	if stationURL.sid == nil {
		return 0
	} else {
		return *stationURL.sid
	}
}

// CID returns the StationURL CID value
func (stationURL *StationURL) CID() uint32 {
	if stationURL.cid == nil {
		return 0
	} else {
		return *stationURL.cid
	}
}

// PID returns the StationURL PID value
func (stationURL *StationURL) PID() *PID {
	return stationURL.pid
}

// Type returns the StationURL type
func (stationURL *StationURL) Type() uint32 {
	if stationURL.transportType == nil {
		return 0
	} else {
		return *stationURL.transportType
	}
}

// RVCID returns the StationURL RVCID
func (stationURL *StationURL) RVCID() uint32 {
	if stationURL.rvcid == nil {
		return 0
	} else {
		return *stationURL.rvcid
	}
}

// Natm returns the StationURL Natm value
func (stationURL *StationURL) Natm() uint32 {
	if stationURL.natm == nil {
		return 0
	} else {
		return *stationURL.natm
	}
}

// Natf returns the StationURL Natf value
func (stationURL *StationURL) Natf() uint32 {
	if stationURL.natf == nil {
		return 0
	} else {
		return *stationURL.natf
	}
}

// Upnp returns the StationURL Upnp value
func (stationURL *StationURL) Upnp() uint32 {
	if stationURL.upnp == nil {
		return 0
	} else {
		return *stationURL.upnp
	}
}

// Pmp returns the StationURL Pmp value
func (stationURL *StationURL) Pmp() uint32 {
	if stationURL.pmp == nil {
		return 0
	} else {
		return *stationURL.pmp
	}
}

// ProbeInit returns the StationURL ProbeInit value
func (stationURL *StationURL) ProbeInit() uint32 {
	if stationURL.probeinit == nil {
		return 0
	} else {
		return *stationURL.probeinit
	}
}

// PRID returns the StationURL PRID value
func (stationURL *StationURL) PRID() uint32 {
	if stationURL.prid == nil {
		return 0
	} else {
		return *stationURL.prid
	}
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
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetPort(uint32(ui64))
		case "Pl":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetPL(uint32(ui64))
		case "stream":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetStream(uint32(ui64))
		case "sid":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetSID(uint32(ui64))
		case "CID":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetCID(uint32(ui64))
		case "PID":
			ui64, _ := strconv.ParseUint(value, 10, 64)
			stationURL.SetPID(NewPID(ui64))
		case "type":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetType(uint32(ui64))
		case "RVCID":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetRVCID(uint32(ui64))
		case "natm":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetNatm(uint32(ui64))
		case "natf":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetNatf(uint32(ui64))
		case "upnp":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetUpnp(uint32(ui64))
		case "pmp":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetPmp(uint32(ui64))
		case "probeinit":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetProbeInit(uint32(ui64))
		case "PRID":
			ui64, _ := strconv.ParseUint(value, 10, 32)
			stationURL.SetPRID(uint32(ui64))
		}
	}
}

// EncodeToString encodes the StationURL into a string
func (stationURL *StationURL) EncodeToString() string {
	fields := []string{}

	if stationURL.address != "" {
		fields = append(fields, "address="+stationURL.address)
	}

	if stationURL.port != nil {
		fields = append(fields, "port="+strconv.FormatUint(uint64(stationURL.Port()), 10))
	}

	if stationURL.pl != nil {
		fields = append(fields, "Pl="+strconv.FormatUint(uint64(stationURL.PL()), 10))
	}

	if stationURL.stream != nil {
		fields = append(fields, "stream="+strconv.FormatUint(uint64(stationURL.Stream()), 10))
	}

	if stationURL.sid != nil {
		fields = append(fields, "sid="+strconv.FormatUint(uint64(stationURL.SID()), 10))
	}

	if stationURL.cid != nil {
		fields = append(fields, "CID="+strconv.FormatUint(uint64(stationURL.CID()), 10))
	}

	if stationURL.pid != nil {
		fields = append(fields, "PID="+strconv.FormatUint(uint64(stationURL.PID().pid), 10))
	}

	if stationURL.transportType != nil {
		fields = append(fields, "type="+strconv.FormatUint(uint64(stationURL.Type()), 10))
	}

	if stationURL.rvcid != nil {
		fields = append(fields, "RVCID="+strconv.FormatUint(uint64(stationURL.RVCID()), 10))
	}

	if stationURL.natm != nil {
		fields = append(fields, "natm="+strconv.FormatUint(uint64(stationURL.Natm()), 10))
	}

	if stationURL.natf != nil {
		fields = append(fields, "natf="+strconv.FormatUint(uint64(stationURL.Natf()), 10))
	}

	if stationURL.upnp != nil {
		fields = append(fields, "upnp="+strconv.FormatUint(uint64(stationURL.Upnp()), 10))
	}

	if stationURL.pmp != nil {
		fields = append(fields, "pmp="+strconv.FormatUint(uint64(stationURL.Pmp()), 10))
	}

	if stationURL.probeinit != nil {
		fields = append(fields, "probeinit="+strconv.FormatUint(uint64(stationURL.ProbeInit()), 10))
	}

	if stationURL.prid != nil {
		fields = append(fields, "PRID="+strconv.FormatUint(uint64(stationURL.PRID()), 10))
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
	Code uint32
}

// IsSuccess returns true if the Result is a success
func (result *Result) IsSuccess() bool {
	return int(result.Code)&errorMask == 0
}

// IsError returns true if the Result is a error
func (result *Result) IsError() bool {
	return int(result.Code)&errorMask != 0
}

// ExtractFromStream extracts a Result structure from a stream
func (result *Result) ExtractFromStream(stream *StreamIn) error {
	code, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read Result code. %s", err.Error())
	}

	result.Code = code

	return nil
}

// Bytes encodes the Result and returns a byte array
func (result *Result) Bytes(stream *StreamOut) []byte {
	stream.WriteUInt32LE(result.Code)

	return stream.Bytes()
}

// Copy returns a new copied instance of Result
func (result *Result) Copy() *Result {
	return NewResult(result.Code)
}

// Equals checks if the passed Structure contains the same data as the current instance
func (result *Result) Equals(other *Result) bool {
	return result.Code == other.Code
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
		b.WriteString(fmt.Sprintf("%scode: %d (success)\n", indentationValues, result.Code))
	} else {
		b.WriteString(fmt.Sprintf("%scode: %d (error)\n", indentationValues, result.Code))
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

	copied.SetStructureVersion(resultRange.StructureVersion())
	copied.Offset = resultRange.Offset
	copied.Length = resultRange.Length

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (resultRange *ResultRange) Equals(structure StructureInterface) bool {
	other := structure.(*ResultRange)

	if resultRange.StructureVersion() == other.StructureVersion() {
		return false
	}

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
