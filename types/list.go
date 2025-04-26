package types

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// List is an implementation of rdv::qList.
// This data type holds an array of other types.
//
// Unlike Buffer and qBuffer, which use the same data type with differing size field lengths,
// there does not seem to be an official rdv::List type
type List[T RVType] []T

// WriteTo writes the List to the given writable
func (l List[T]) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(len(l)))

	for _, v := range l {
		v.WriteTo(writable)
	}
}

func (l List[T]) extractType(t any, readable Readable) error {
	// * This just makes List.ExtractFrom() a bit cleaner
	// * since it doesn't have to type check
	if ptr, ok := t.(RVTypePtr); ok {
		return ptr.ExtractFrom(readable)
	}

	// * Maybe support other types..?

	return fmt.Errorf("Unsupported List type %T", t)
}

// ExtractFrom extracts the List from the given readable
func (l *List[T]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	slice := make([]T, 0, length)

	for i := 0; i < int(length); i++ {
		var value T
		if err := l.extractType(&value, readable); err != nil {
			return err
		}

		slice = append(slice, value)
	}

	*l = slice

	return nil
}

// Copy returns a pointer to a copy of the List. Requires type assertion when used
func (l List[T]) Copy() RVType {
	copied := make(List[T], 0)

	for _, v := range l {
		copied = append(copied, v.Copy().(T))
	}

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (l List[T]) Equals(o RVType) bool {
	if _, ok := o.(List[T]); !ok {
		return false
	}

	other := o.(List[T])

	if len(l) != len(other) {
		return false
	}

	for i := 0; i < len(l); i++ {
		if !l[i].Equals(other[i]) {
			return false
		}
	}

	return true
}

// CopyRef copies the current value of the List
// and returns a pointer to the new copy
func (l List[T]) CopyRef() RVTypePtr {
	copied := l.Copy().(List[T])
	return &copied
}

// Deref takes a pointer to the List
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (l *List[T]) Deref() RVType {
	return *l
}

// Contains checks if the provided value exists in the List
func (l List[T]) Contains(checkValue T) bool {
	for _, v := range l {
		if v.Equals(checkValue) {
			return true
		}
	}

	return false
}

// String returns a string representation of the struct
func (l List[T]) String() string {
	return fmt.Sprintf("%v", ([]T)(l))
}

// Scan implements the sql.Scanner interface for List[T]
//
// Only designed for Postgres databases
func (l *List[T]) Scan(value any) error {
	if value == nil {
		return nil
	}

	// * Ensure the value is in the right format
	var pgArray string
	switch v := value.(type) {
	case []byte:
		pgArray = string(v)
	case string:
		pgArray = v
	default:
		return fmt.Errorf("unsupported Scan type for List: %T", value)
	}

	// * Postgres formats arrays in curly braces,
	// * such as `{"string1"}`
	if len(pgArray) < 2 || pgArray[0] != '{' || pgArray[len(pgArray)-1] != '}' {
		return fmt.Errorf("invalid PostgreSQL array format: %s", pgArray)
	}

	var zero T
	isByteaArray := false

	pgArray = strings.TrimSuffix(pgArray, "}")
	pgArray = strings.TrimPrefix(pgArray, "{")

	// * Postgres formats bytea colums as arrays
	// * of hex strings. Which makes this a list
	// * of arrays, formatted like
	// * `{{"\\x35","\\x36","\\x37","\\x38","\\x39"},{"\\x35","\\x36","\\x37","\\x38","\\x39"}}`
	switch any(zero).(type) {
	case Buffer, QBuffer:
		// * Trim any extra off the ends, if exists.
		// * Nested arrays, parsed later
		pgArray = strings.TrimSuffix(pgArray, "}")
		pgArray = strings.TrimPrefix(pgArray, "{")
		isByteaArray = true
	}

	// * Bail if array is empty, who cares
	if pgArray == "" {
		return nil
	}

	var elements []string
	var err error

	if isByteaArray {
		// * Arrays of bytea (byte slices) are handled as
		// * nested arrays. Handle these as special cases
		elements = strings.Split(pgArray, "},{")
	} else {
		// * Basic array. Go's CSV reader can handle these strings,
		// * including strings which contain the delimiter. Such as:
		// * `{"string1","string2","strings can have spaces, commas, etc."}`
		reader := csv.NewReader(strings.NewReader(pgArray))
		reader.Comma = ','

		elements, err = reader.Read()
	}

	if err != nil {
		return nil
	}

	result := make(List[T], 0, len(elements))

	// * This is technically less effecient than running the
	// * switch-case FIRST, and using a for loop inside each
	// * case block. But that's ugly. This is fine for now.
	// * This also assumes all numbers are within range, which
	// * they should be for our purposes but this is not guaranteed
	for _, element := range elements {
		// * Only support basic/simple types. Nothing else can
		// * be safely stored in Postgres. Complex types should
		// * opt for JSON mode
		switch any(zero).(type) {
		case String:
			result = append(result, any(String(element)).(T))
		case Bool:
			b := element == "t" || element == "true"
			result = append(result, any(Bool(b)).(T))
		case Double:
			d, err := strconv.ParseFloat(element, 64)
			if err != nil {
				return err
			}

			result = append(result, any(Double(d)).(T))
		case Float:
			f, err := strconv.ParseFloat(element, 32)
			if err != nil {
				return err
			}

			result = append(result, any(Float(f)).(T))
		case Int8:
			i, err := strconv.ParseInt(element, 10, 8)
			if err != nil {
				return err
			}

			result = append(result, any(Int8(i)).(T))
		case Int16:
			i, err := strconv.ParseInt(element, 10, 16)
			if err != nil {
				return err
			}

			result = append(result, any(Int16(i)).(T))
		case Int32:
			i, err := strconv.ParseInt(element, 10, 32)
			if err != nil {
				return err
			}

			result = append(result, any(Int32(i)).(T))
		case Int64:
			i, err := strconv.ParseInt(element, 10, 64)
			if err != nil {
				return err
			}

			result = append(result, any(Int64(i)).(T))
		case UInt8:
			i, err := strconv.ParseUint(element, 10, 8)
			if err != nil {
				return err
			}

			result = append(result, any(UInt8(i)).(T))
		case UInt16:
			i, err := strconv.ParseUint(element, 10, 16)
			if err != nil {
				return err
			}

			result = append(result, any(UInt16(i)).(T))
		case UInt32:
			i, err := strconv.ParseUint(element, 10, 32)
			if err != nil {
				return err
			}

			result = append(result, any(UInt32(i)).(T))
		case UInt64:
			i, err := strconv.ParseUint(element, 10, 64)
			if err != nil {
				return err
			}

			result = append(result, any(UInt64(i)).(T))
		case PID:
			i, err := strconv.ParseUint(element, 10, 64)
			if err != nil {
				return err
			}

			result = append(result, any(PID(i)).(T))
		case QUUID:
			quuid := NewQUUID([]byte{})
			quuid.FromString(element)

			result = append(result, any(quuid).(T))
		case Buffer, QBuffer:
			// * Buffer and QBuffer are type aliases of []byte.
			// * When []byte is used in Postgres, the data is
			// * stored in a somewhat odd what. Every byte is
			// * encoded into the numeric string it represents
			// * and then that string is converted into hex and
			// * stored as a Postgres list.
			// *
			// * For example, assume the following is used to
			// * create a Buffer: `NewBuffer([]byte{0, 1, 2, 3, 4})`
			// *
			// * When this is sent to `List.Scan`, the data is:
			// * `[123 123 34 92 92 120 51 48 34 44 34 92 92 120
			// *   51 49 34 44 34 92 92 120 51 50 34 44 34 92 92
			// *   120 51 51 34 44 34 92 92 120 51 52 34 125 125]`
			// *
			// * When converted to a `string``, this becomes:
			// * `{{"\\x30","\\x31","\\x32","\\x33","\\x34"}}`
			// *
			// * The actual `Buffer` value in Postgres is the
			// * byte list `{"\\x30","\\x31","\\x32","\\x33","\\x34"}`
			// *
			// * Each element in this list is a byte for the ASCII
			// * value of the real number (`30` -> `0`, `31` -> `1`, etc.)
			// *
			// * The same goes for higher values as well. The byte
			// * `0xFF` is stored as `"\\x323535"` (the string `255`)
			// *
			// * To get the real values, we must first decode the string
			// * back into ASCII, and then convert the ASCII to a number
			// *
			// * Since the `byteStrings` will always represent bytes in a
			// * standard format, we can safely just split on `,`
			byteStrings := strings.Split(element, ",")
			bytes := make([]byte, 0, len(byteStrings))

			for _, byteString := range byteStrings {
				byteString = strings.TrimPrefix(byteString, "\"")
				byteString = strings.TrimSuffix(byteString, "\"")
				byteString = strings.TrimPrefix(byteString, "\\\\x")
				asciiBytes, err := hex.DecodeString(byteString)
				if err != nil {
					return err
				}

				char, err := strconv.Atoi(string(asciiBytes))
				if err != nil {
					return err
				}

				bytes = append(bytes, byte(char))
			}

			switch any(zero).(type) {
			case Buffer:
				result = append(result, any(Buffer(bytes)).(T))
			case QBuffer:
				result = append(result, any(QBuffer(bytes)).(T))
			}
		case DateTime:
			dt := NewDateTime(0)
			if err := dt.scanSQLString(element); err != nil {
				return err
			}
			result = append(result, any(dt).(T))
		case QResult:
			i, err := strconv.ParseUint(element, 10, 32)
			if err != nil {
				return err
			}

			result = append(result, any(QResult(i)).(T))
		case StationURL:
			stationURL := NewStationURL("")
			stationURL.SetURL(element)
			stationURL.Parse()
			result = append(result, any(stationURL).(T))
		default:
			return fmt.Errorf("unsupported List element type: %T", zero)
		}
	}

	*l = result
	return nil
}

// NewList returns a new List of the provided type
func NewList[T RVType]() List[T] {
	return make(List[T], 0)
}
