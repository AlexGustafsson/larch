package warc

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const tagName = "warc"

// Marshal converts an interface to a `key: value` format.
func Marshal(v interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := io.Writer(buffer)

	structType := reflect.TypeOf(v)

	// Unwrap the value if it's a pointer
	isPointer := false
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
		isPointer = true
	}

	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Expected kind struct got %v", structType.Kind())
	}

	structValue := reflect.ValueOf(v)
	if isPointer {
		structValue = reflect.Indirect(structValue)
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		// Skip non-public values as they cannot be accessed
		if !fieldValue.CanInterface() {
			continue
		}

		// Parse the field options
		options := strings.Split(field.Tag.Get(tagName), ",")
		omitEmpty := false
		tag := field.Name
		for _, option := range options {
			if option == "omitempty" {
				omitEmpty = true
			} else if option != "" {
				tag = option
			}
		}

		value := fieldValue.Interface()

		include := !omitEmpty || (omitEmpty && !isEmpty(value))
		if include {
			fmt.Fprintf(writer, "%s: %s\r\n", tag, convertValue(value))
		}
	}

	return buffer.Bytes(), nil
}

// Unmarshal sets values in a structure from a `key: value` format.
func Unmarshal(data []byte, v interface{}) error {
	return nil
}

func convertValue(value interface{}) interface{} {
	switch castValue := value.(type) {
	case time.Time:
		return castValue.Format("2006-01-02T15:04:05-0700")
	case time.Duration:
		return convertValue(castValue.Milliseconds())
	case int64:
		return strconv.FormatInt(castValue, 10)
	case uint64:
		return strconv.FormatUint(castValue, 10)
	default:
		return castValue
	}
}

func isEmpty(value interface{}) bool {
	switch castValue := value.(type) {
	case string:
		return castValue == ""
	case time.Time:
		return castValue.IsZero()
	default:
		return castValue == nil
	}
}
