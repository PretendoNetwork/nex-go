package types

import (
	"encoding/csv"
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
func (l *List[T]) Scan(value interface{}) error {
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
	case Buffer, QBuffer, QUUID:
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
		case Buffer, QBuffer, QUUID:
			// * Each element is stored as a hex string encoded such as:
			// * `"\\x30","\\x31","\\x32","\\x33","\\x34"`.
			// * Go's CSV reader can handle these strings to get the individual
			// * bytes out. Assumes bytes are in the correct format

			reader := csv.NewReader(strings.NewReader(element))
			reader.Comma = ','

			byteStrings, err := reader.Read()
			if err != nil {
				return err
			}

			bytes := make([]byte, 0, len(byteStrings))

			// * Convert from \\xXX to bytes. Each "byte" is actually
			// * the ASCII value, not the real byte value. Unsure if
			// * that matters, but this works during testing
			for _, byteString := range byteStrings {
				byteString = strings.TrimPrefix(byteString, "\\\\x")
				hexVal, err := strconv.ParseUint(byteString, 16, 8)
				if err != nil {
					return err
				}

				char, err := strconv.Atoi(string(hexVal))
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
			case QUUID:
				result = append(result, any(QUUID(bytes)).(T))
			}
		case DateTime:
			dt := DateTime(0)
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
