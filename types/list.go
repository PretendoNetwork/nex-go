package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
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

// Scan implements sql.Scanner for database/sql
func (l *List[T]) Scan(value any) error {
	if value == nil {
		*l = List[T]{}
		return nil
	}

	var pgArray string
	switch v := value.(type) {
	case []byte:
		pgArray = string(v)
	case string:
		pgArray = v
	default:
		return fmt.Errorf("unsupported Scan type for List: %T", value)
	}

	if len(pgArray) < 2 || pgArray[0] != '{' || pgArray[len(pgArray)-1] != '}' {
		return fmt.Errorf("invalid PostgreSQL array format: %s", pgArray)
	}

	pgArray = pgArray[1 : len(pgArray)-1]

	if pgArray == "" {
		*l = List[T]{}
		return nil
	}

	var elements []string

	if strings.HasPrefix(pgArray, "\"") {
		reader := csv.NewReader(strings.NewReader(pgArray))
		reader.Comma = ','

		var err error
		elements, err = reader.Read()

		if err != nil {
			return fmt.Errorf("failed to parse array elements: %w", err)
		}
	} else {
		elements = strings.Split(pgArray, ",")
	}

	var zero T
	isByteaType := false

	switch any(zero).(type) {
	case Buffer, QBuffer:
		isByteaType = true
	}

	result := make(List[T], 0, len(elements))

	for _, element := range elements {
		var new T
		ptr := any(&new)

		if scanner, ok := ptr.(sql.Scanner); ok {
			var scanValue any = element

			if isByteaType {
				hexStr := element
				hexStr = strings.TrimPrefix(hexStr, "\\\\x")
				hexStr = strings.TrimPrefix(hexStr, "\\x")
				bytes, err := hex.DecodeString(hexStr)

				if err != nil {
					return fmt.Errorf("failed to decode hex for bytea element %q: %w", element, err)
				}

				scanValue = bytes
			}

			if err := scanner.Scan(scanValue); err != nil {
				return fmt.Errorf("failed to scan element: %w", err)
			}

			result = append(result, new)
		} else {
			return fmt.Errorf("list element type %T does not implement sql.Scanner", zero)
		}
	}

	*l = result

	return nil
}

// Value implements driver.Valuer for database/sql
func (l List[T]) Value() (driver.Value, error) {
	if len(l) == 0 {
		return "{}", nil
	}

	var b strings.Builder
	b.WriteString("{")

	for i, new := range l {
		if i > 0 {
			b.WriteString(",")
		}

		if valuer, ok := any(new).(driver.Valuer); ok {
			v, err := valuer.Value()
			if err != nil {
				return nil, fmt.Errorf("failed to get value for element %d: %w", i, err)
			}

			switch val := v.(type) {
			case string:
				escaped := strings.ReplaceAll(val, "\\", "\\\\")
				escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

				b.WriteString("\"")
				b.WriteString(escaped)
				b.WriteString("\"")
			case int64, float64, bool:
				b.WriteString(fmt.Sprintf("%v", val))
			case []byte:
				b.WriteString("\"\\\\x")
				b.WriteString(hex.EncodeToString(val))
				b.WriteString("\"")
			case time.Time:
				b.WriteString("\"")
				b.WriteString(val.Format("2006-01-02 15:04:05"))
				b.WriteString("\"")
			case nil:
				b.WriteString("NULL")
			default:
				return nil, fmt.Errorf("unsupported value type %T for element %d", val, i)
			}
		} else {
			return nil, fmt.Errorf("element type %T does not implement driver.Valuer", new)
		}
	}

	b.WriteString("}")

	return b.String(), nil
}

// NewList returns a new List of the provided type
func NewList[T RVType]() List[T] {
	return make(List[T], 0)
}
