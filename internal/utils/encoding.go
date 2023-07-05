package utils

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Converts all value types into their standard type
//
// i.e int64 to int, or structs to maps
func Normalize(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	val, yes := v.(encoding.BinaryMarshaler)
	if yes {
		return val, nil
	}
	vv, vt := GetValueAndType(v)
	if vt.Kind() == reflect.Ptr {
		return nil, nil
	}
	intf := vv.Interface()
	_, isTime := intf.(time.Time)
	if isTime {
		return intf, nil
	}
	normalized, err := normalizeValues(vt, vv)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}

// Converts all 8, 16, 32, 64-bit value types into 64-bit or 32-bit
func normalizeValues(typ reflect.Type, val reflect.Value) (interface{}, error) {
	switch {
	case isUInt(typ):
		return val.Uint(), nil
	case isSInt(typ):
		return val.Int(), nil
	case isFloat(typ):
		return val.Float(), nil
	case isString(typ):
		return val.String(), nil
	case isBool(typ):
		return val.Bool(), nil
	// or []byte
	case isList(typ):
		// []byte returns as it, does not normalize into []uint
		return normalizeSlice(val)
	case isMap(typ):
		return normalizeMap(val)
	case isStruct(typ):
		return normalizeStruct(val)
	}
	return nil, fmt.Errorf("invalid type: %s", typ.Name())
}

// Check if type is a struct
func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

// Check if type is a map
func isMap(t reflect.Type) bool {
	return t.Kind() == reflect.Map
}

// Check if is string
func isString(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

// Check if is a boolean
func isBool(t reflect.Type) bool {
	return t.Kind() == reflect.Bool
}

// Check if is either an array or a slice
func isList(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

// Check if the type is unsigned integer
func isUInt(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Uint:
		return true
	case reflect.Uint8:
		return true
	case reflect.Uint16:
		return true
	case reflect.Uint32:
		return true
	case reflect.Uint64:
		return true
	default:
		return false
	}
}

// Check if the type is signed integer
func isSInt(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int:
		return true
	case reflect.Int8:
		return true
	case reflect.Int16:
		return true
	case reflect.Int32:
		return true
	case reflect.Int64:
		return true
	default:
		return false
	}
}

// Check if the type is float
func isFloat(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// Normalize a slice/array.
// A byte array returns as it
func normalizeSlice(v reflect.Value) (interface{}, error) {
	// uint8 = byte
	// slice/array of bytes should be returned as it
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return v.Interface(), nil
	}
	s := make([]interface{}, 0)
	for i := 0; i < v.Len(); i += 1 {
		sv := v.Index(i).Interface()
		v, err := Normalize(sv)
		if err != nil {
			return nil, err
		}
		s = append(s, v)
	}
	return s, nil
}

// Normalize a map by normalizing its elements
func normalizeMap(v reflect.Value) (map[string]interface{}, error) {
	keytype := v.Type().Key().Kind()
	if keytype != reflect.String {
		return nil, fmt.Errorf("map key type must be string")
	}
	m := make(map[string]interface{})
	for _, k := range v.MapKeys() {
		val := v.MapIndex(k)
		n, err := Normalize(val.Interface())
		if err != nil {
			return nil, err
		}
		m[k.String()] = n
	}
	return m, nil
}

// Normalize a struct into key-value pairs of attributes.
//
// For normal fields it normalizes the value and add to the resulting
// map to the key equal to the field name or the value in the `pson` tag.
//
// Embedded structs have their fields directly the the parent struct, not as a seperate key
func normalizeStruct(v reflect.Value) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for i := 0; i < v.NumField(); i += 1 {
		typ := v.Type().Field(i)
		val := v.Field(i)
		if typ.PkgPath == "" {
			picoTag := typ.Tag.Get("pson")
			key, omitempty := ProcessTag(picoTag)
			if len(key) == 0 {
				key = typ.Name
			}
			isZero := isZeroValued(val)
			if isZero && omitempty {
				continue
			}
			normal, err := Normalize(val.Interface())
			if err != nil {
				return nil, err
			}
			if !typ.Anonymous {
				m[key] = normal
				continue
			}
			// is an embedded field
			normalmap, isMap := normal.(map[string]interface{})
			if !isMap {
				m[key] = normal
				continue
			}
			for nk, nv := range normalmap {
				m[nk] = nv
			}
		}
	}
	return m, nil
}

// Process the struct tag and return both the first value (key to use in document) and either to omitempty the
// field associated with the tag
func ProcessTag(tag string) (string, bool) {
	ts := strings.Split(tag, ",")
	// `pson:""` -> ts = {""}, len(1)
	key := ts[0]
	if len(ts) > 1 {
		if ts[1] == "omitempty" {
			return key, true
		}
	}
	return key, false
}

// Check if a value is zero-valued
func isZeroValued(v reflect.Value) bool {
	typ := v.Type()
	kind := typ.Kind()
	switch {
	case isList(typ) || isString(typ) || isMap(typ):
		return v.Len() == 0
	case isBool(typ):
		return false
	case isSInt(typ):
		return v.Int() == 0
	case isUInt(typ):
		return v.Uint() == 0
	case isFloat(typ):
		return v.Float() == 0
	case kind == reflect.Interface || kind == reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

// Get the underlying value of v even if it is a pointer.
// Returns nil and type of kind reflect.Ptr if it is nil pointer
func GetValueAndType(v interface{}) (reflect.Value, reflect.Type) {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Ptr && !rv.IsNil() {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	return rv, rt
}
