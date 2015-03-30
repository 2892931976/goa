package design

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// A Kind represents the specific kind of type that a DataType represents.
type Kind uint

const (
	NullType Kind = iota
	BooleanType
	IntegerType
	NumberType
	StringType
	ArrayType
	ObjectType
)

// DataType interface represents both JSON schema types and media types.
// All data types have a kind (Integer, Number etc. for JSON schema types
// and Object for media types) and a "Load" method.
// The "Load" method checks that the value of the given argument is compatible
// with the type and returns the coerced value if that's case, an error otherwise.
// Data types are used to define the type of media type members and of action
// parameters.
type DataType interface {
	Kind() Kind                            // integer, number, string, ...
	Name() string                          // Human readable name
	Load(interface{}) (interface{}, error) // Validate and load
}

// Type for null, boolean, integer, number and string
type Primitive Kind

var (
	// Type for the JSON null value
	Null = Primitive(NullType)
	// Type for a JSON boolean
	Boolean = Primitive(BooleanType)
	// Type for a JSON number without a fraction or exponent part
	Integer = Primitive(IntegerType)
	// Type for any JSON number, including integers
	Number = Primitive(NumberType)
	// Type for a JSON string
	String = Primitive(StringType)
)

// Type kind
func (b Primitive) Kind() Kind {
	return Kind(b)
}

// Human readable name of basic type
func (b Primitive) Name() string {
	switch Kind(b) {
	case NullType:
		return "null"
	case BooleanType:
		return "boolean"
	case IntegerType:
		return "integer"
	case NumberType:
		return "number"
	case StringType:
		return "string"
	default:
		panic(fmt.Sprintf("goa bug: unknown basic type %#v", b))
	}
}

// Attempt to load value into basic type
// How a value is coerced depends on its type and the basic type kind:
// Only strings may be loaded in values of type String.
// Any integer value or string representing an integer may be loaded in values of type Integer.
// Any integer or float value or string representing integers or floats may be loaded in values of
// type Number.
// true, false, 1, 0, "false", "FALSE", "0", "f", "F", "true", "TRUE", "1", "t", "T" may be loaded
// in values of type Boolean.
// Returns nil and an error if coercion fails.
func (b Primitive) Load(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	var extra string
	switch Kind(b) {
	case BooleanType:
		switch v := value.(type) {
		case bool:
			return value, nil
		case string:
			if res, err := strconv.ParseBool(v); err == nil {
				return res, nil
			} else {
				extra = err.Error()
			}
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			if value == 0 {
				return false, nil
			} else if value == 1 {
				return true, nil
			} else {
				extra = "integer value must be 0 or 1"
			}
		}
	case IntegerType:
		switch v := value.(type) {
		case int:
			return v, nil
		case uint:
			return int(v), nil
		case int8:
			return int(v), nil
		case uint8:
			return int(v), nil
		case int16:
			return int(v), nil
		case uint16:
			return int(v), nil
		case int32:
			return int(v), nil
		case uint32:
			return int(v), nil
		case int64:
			return int(v), nil
		case uint64:
			return int(v), nil
		case string:
			if res, err := strconv.ParseInt(v, 10, 0); err == nil {
				return int(res), nil
			} else {
				extra = err.Error()
			}
		}
	case NumberType:
		switch v := value.(type) {
		case int:
			return float64(v), nil
		case uint:
			return float64(v), nil
		case int8:
			return float64(v), nil
		case uint8:
			return float64(v), nil
		case int16:
			return float64(v), nil
		case uint16:
			return float64(v), nil
		case int32:
			return float64(v), nil
		case uint32:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case uint64:
			return float64(v), nil
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		case string:
			if res, err := strconv.ParseFloat(v, 64); err == nil {
				return res, nil
			} else {
				extra = err.Error()
			}
		}
	case StringType:
		switch v := value.(type) {
		case time.Time:
			return v.Format(time.RFC3339), nil
		case string:
			return value, nil
		}
	}
	return nil, &IncompatibleValue{value: value, to: b.Name(), extra: extra}
}

// An array of values of type ElemType
type Array struct {
	ElemType DataType
}

// Type kind
func (a *Array) Kind() Kind {
	return ArrayType
}

// Load coerces the given value into a []interface{} where the array values have all been coerced recursively.
// `value` must either be a slice, an array or a string containing a JSON representation of an array.
// Load also applies any validation rule defined in the array element members.
// Returns nil and an error if coercion or validation fails.
func (a *Array) Load(value interface{}) (interface{}, error) {
	var arr []interface{}
	k := reflect.TypeOf(value).Kind()
	if k == reflect.String {
		if err := json.Unmarshal([]byte(value.(string)), &arr); err != nil {
			return nil, &IncompatibleValue{value: value, to: "Array",
				extra: fmt.Sprintf("failed to decode JSON: %v", err)}
		}
	} else if k == reflect.Slice || k == reflect.Array {
		v := reflect.ValueOf(value)
		for i := 0; i < v.Len(); i++ {
			arr = append(arr, v.Index(i).Interface())
		}
	} else {
		return nil, &IncompatibleValue{value: value, to: "Array",
			extra: "value must be an array or a slice"}
	}
	var res []interface{}
	varr := reflect.ValueOf(arr)
	for i := 0; i < varr.Len(); i++ {
		ev, err := a.ElemType.Load(varr.Index(i).Interface())
		if err != nil {
			return nil, &IncompatibleValue{value: value, to: "Array",
				extra: fmt.Sprintf("cannot load value at index %v: %v", i, err)}
		}
		res = append(res, ev)
	}
	return interface{}(res), nil
}

// JSON schema type name
func (a *Array) Name() string {
	return "array"
}

// A JSON object
type Object map[string]*Member

// Type kind
func (o Object) Kind() Kind {
	return ObjectType
}

// Load coerces the given value into a map[string]interface{} where the map values have all been coerced recursively.
// `value` must either be a map with string keys or to a string containing a JSON representation of a map.
// Load also applies any validation rule defined in the object members.
// Returns `nil` and an error if coercion or validation fails.
func (o Object) Load(value interface{}) (interface{}, error) {
	// First load from JSON if needed
	var m map[string]interface{}
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, &IncompatibleValue{
				value: v,
				to:    "Object",
				extra: "string is not a JSON object",
			}
		}
	case map[string]interface{}:
		m = v
	default:
		return nil, &IncompatibleValue{value: value, to: "Object"}
	}
	// Now go through each type member and load and validate value from map
	coerced := make(map[string]interface{})
	var errors []error
	for n, member := range o {
		val, _ := m[n]
		if val == nil {
			if member.DefaultValue != nil {
				val = member.DefaultValue
			}
		} else {
			var err error
			val, err = member.Load(n, val)
			if err != nil {
				errors = append(errors, &IncompatibleValue{
					value,
					"Object",
					fmt.Sprintf("could not load member %s: %s", n, err),
				})
				continue
			}
			for _, validation := range member.Validations {
				if err := validation(n, val); err != nil {
					errors = append(errors, err)
					continue
				}
			}
		}
		coerced[n] = val
	}
	if len(errors) > 0 {
		// TBD create MultiError type
		return nil, errors[0]
	}
	return coerced, nil
}

// JSON schema type name
func (a Object) Name() string {
	return "object"
}

// An object member with optional description, default value and validations
type Member struct {
	Type         DataType     // Member type
	Description  string       // Optional description
	Validations  []Validation // Optional validation functions
	DefaultValue interface{}  // Optional member default value
}

// A validation takes a value and produces nil on success or an error otherwise
type Validation func(name string, val interface{}) error

// M is a helper function that creates a member.
// M is intended for declaring a literal object.
// Usage:
//    obj := Object{"Name": M{String, "Object name"}}
func M(typ DataType, desc string) *Member {
	return &Member{Type: typ, Description: desc}
}

// Create member from object.
// Useful to define a payload member from a media type for example.
// Usage:
//    createPayload := From(TaskMediaType.Object).Required("Details")
func From(o Object) *Member {
	return &Member{Type: o}
}

// Load calls load on the underlying type then runs any member validation.
func (m *Member) Load(name string, value interface{}) (interface{}, error) {
	res, err := m.Type.Load(value)
	if err != nil {
		return nil, err
	}
	for _, validation := range m.Validations {
		if err := validation(name, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// Error raised when "Load" cannot coerce a value to the data type
type IncompatibleValue struct {
	value interface{} // Value being loaded
	to    string      // Name of type being coerced to
	extra string      // Extra error information if any
}

// Error returns the error message
func (e *IncompatibleValue) Error() string {
	extra := ""
	if len(e.extra) > 0 {
		extra = ": " + e.extra
	}
	return fmt.Sprintf("Cannot load value %v into a %v%s", e.value, e.to, extra)
}
