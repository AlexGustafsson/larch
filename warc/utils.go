package warc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const tagName = "warc"

func unwrapStruct(v interface{}) (reflect.Type, reflect.Value, error) {
	structType := reflect.TypeOf(v)

	// Unwrap the value if it's a pointer
	isPointer := false
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
		isPointer = true
	}

	// Enforce a struct type
	if structType.Kind() != reflect.Struct {
		return nil, reflect.ValueOf(nil), fmt.Errorf("Expected kind struct got %v", structType.Kind())
	}

	// Alway's ensure we're dealing with a pointer to the value, to make the fields'
	// values addressable
	var structValue reflect.Value
	if isPointer {
		structValue = reflect.ValueOf(v).Elem()
	} else {
		structValue = reflect.New(reflect.Indirect(reflect.ValueOf(v)).Type()).Elem()
	}

	return structType, structValue, nil
}

// parseOptions parses a field tag, returning the specified name and whether or not the field's value should be omitted if empty.
func parseOptions(field reflect.StructField) (string, bool) {
	// Parse the field options
	options := strings.Split(field.Tag.Get(tagName), ",")
	omitEmpty := false
	wantedName := field.Name
	for _, option := range options {
		if option == "omitempty" {
			omitEmpty = true
		} else if option != "" {
			wantedName = option
		}
	}

	return wantedName, omitEmpty
}

// Marshal converts an interface to a `key: value` format.
func Marshal(v interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := io.Writer(buffer)

	structType, structValue, err := unwrapStruct(v)
	if err != nil {
		return nil, err
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip non-public values as they cannot be accessed
		if !fieldValue.CanInterface() {
			continue
		}

		wantedName, omitEmpty := parseOptions(field)

		value := fieldValue.Interface()

		include := !omitEmpty || (omitEmpty && !isEmpty(value))
		if include {
			fmt.Fprintf(writer, "%s: %s\r\n", wantedName, serializeValue(value))
		}
	}

	return buffer.Bytes(), nil
}

// UnmarshalStream sets values in a structure from a `key: value` format.
func UnmarshalStream(reader *bufio.Reader, v interface{}) error {
	// Actual name of fields by their "wanted name"
	fieldNames := make(map[string]string)

	structType, structValue, err := unwrapStruct(v)
	if err != nil {
		return err
	}

	// Populate the declared field names
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Skip non-public values as they cannot be accessed
		if !fieldValue.CanSet() || !fieldValue.CanInterface() {
			continue
		}

		wantedName, _ := parseOptions(field)
		fieldNames[wantedName] = field.Name
	}

	for {
		// TODO: This may cause issues with really long lines, implement "isPrefix" correctly
		buffer, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				// We reached the end of the file, stop parsing
				break
			} else {
				return fmt.Errorf("Unable to read line whilst unmarshalling: %v", err)
			}
		}

		// If the line is empty, stop reading
		line := string(buffer)
		if line == "" {
			break
		}

		values := strings.Split(line, ": ")
		if len(values) != 2 {
			return fmt.Errorf("Got bad line. Expected 'key: value' got '%v'", values)
		}

		name := values[0]
		value := values[1]

		actualName := fieldNames[name]
		if actualName == "" {
			continue
		}

		field := structValue.FieldByName(actualName)
		if !field.IsValid() {
			return fmt.Errorf("Got invalid field '%v'", actualName)
		}

		err = setValue(field, value)
		if err != nil {
			return fmt.Errorf("Unable to set value for field %v: %v", field, err)
		}
	}

	return nil
}

// Unmarshal sets values in a structure from a `key: value` format.
func Unmarshal(data []byte, v interface{}) error {
	reader := bytes.NewReader(data)
	bufferedReader := bufio.NewReader(reader)
	return UnmarshalStream(bufferedReader, v)
}

// serializeValue converts a supported value to its string representation.
func serializeValue(value interface{}) interface{} {
	switch castValue := value.(type) {
	case time.Time:
		return castValue.Format("2006-01-02T15:04:05-0700")
	case time.Duration:
		return serializeValue(castValue.Milliseconds())
	case int64:
		return strconv.FormatInt(castValue, 10)
	case uint64:
		return strconv.FormatUint(castValue, 10)
	default:
		return castValue
	}
}

// isEmpty returns whether or not a supported value is empty.
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

// setValue sets a value of a field given its serialized string representation.
func setValue(field reflect.Value, value string) error {
	switch field.Interface().(type) {
	case string:
		field.SetString(value)
	case int:
		parsedValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("Unable to parse value '%v' for field %v: %v", value, field, err)
		}

		field.SetInt(parsedValue)
	case uint64:
		parsedValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("Unable to parse value '%v' for field %v: %v", value, field, err)
		}

		field.SetUint(parsedValue)
	case time.Time:
		parsedValue, err := time.Parse("2006-01-02T15:04:05-0700", value)
		if err != nil {
			return nil
		}

		field.Set(reflect.ValueOf(parsedValue))
	case time.Duration:
		parsedValue, err := time.ParseDuration(value + "ms")
		if err != nil {
			return fmt.Errorf("Unable to parse value '%v' for field %v: %v", value, field, err)
		}

		field.Set(reflect.ValueOf(parsedValue))

	default:
		return fmt.Errorf("Unsupported type: %v", field.Type())
	}

	return nil
}
