package umap

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pico-db/pico/internal/utils"
	"github.com/vmihailenco/msgpack/v5"
)

// Get all the keys inside a map
func Keys(m map[string]interface{}, sorted bool, includeSubKeys bool) []string {
	kres := make([]string, 0, len(m))
	for k, v := range m {
		added := false
		if includeSubKeys {
			sub, is := v.(map[string]interface{})
			if is {
				sfields := Keys(sub, false, includeSubKeys)
				for _, sk := range sfields {
					kres = append(kres, fmt.Sprintf("%s.%s", k, sk))
				}
				added = true
			}
		}
		if !added {
			kres = append(kres, k)
		}
	}
	if sorted {
		sort.Slice(kres, func(i, j int) bool {
			return kres[i] < kres[j]
		})
	}
	return kres
}

// Make copy of an existing map
func Copy(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		nestedMap, yes := v.(map[string]interface{})
		if yes {
			cp[k] = Copy(nestedMap)
			continue
		}
		cp[k] = v
	}
	return cp
}

// Finds a field inside the map, provided that the key k is formatted as following:
//   - For a first-level key: "key"
//   - For a key inside nested maps: "key.key1.key2"
func Lookup(provided map[string]interface{}, k string, toNearestParent bool) (map[string]interface{}, interface{}, string) {
	splitted := strings.Split(k, ".")
	var exists bool
	var currentValue interface{}
	currentMap := provided
	for ind, subfield := range splitted {
		currentValue, exists = currentMap[subfield]
		m, isMap := currentValue.(map[string]interface{})
		if toNearestParent {
			if (!exists || !isMap) && ind < len(splitted)-1 {
				m := make(map[string]interface{})
				currentMap[subfield] = m
				currentValue = m
				currentMap = m
				continue
			}
		}
		if !exists {
			return nil, nil, ""
		}
		// if not the last field
		if ind < len(splitted)-1 {
			currentMap = m
		}
	}
	// the map at which the last key is inside
	// the value
	// the last key
	return currentMap, currentValue, splitted[len(splitted)-1]
}

// Convert the provided map into bytes.
//
// MessagePack encoding is used due to its efficient use of space
func Encode(v map[string]interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Decode the byte array into a map
func Decode(data []byte, m *map[string]interface{}) error {
	return msgpack.Unmarshal(data, m)
}

// Unpack the map into a struct
func Unmarshal(m map[string]interface{}, to interface{}) error {
	renamed := renameKeysRec(m, to)
	js, err := json.Marshal(renamed)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, to)
}

// Rename the keys in a recursive manner to rename all children maps
func renameKeysRec(m map[string]interface{}, typ interface{}) map[string]interface{} {
	rv, rt := utils.GetValueAndType(typ)
	if rt.Kind() != reflect.Struct {
		return nil
	}
	// after renaming, the map's key is the name of the struct field
	// or the key from m
	renamed := renameKeys(m, rv)
	for i := 0; i < rv.NumField(); i += 1 {
		field := rv.Type().Field(i)
		value := renamed[field.Name]
		t := field.Type
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		valueMap, isMap := value.(map[string]interface{})
		if isMap && t.Kind() == reflect.Struct {
			subrenamed := renameKeysRec(valueMap, rv.Field(i).Interface())
			renamed[field.Name] = subrenamed
		}
	}
	return renamed
}

// Takes in a map and a struct.
// Unpack the struct to find fields annotated with `pson` and rename the fields accordingly.
//
// The key of the resulting map is the field name of the struct that includes `pson` or the key in the existing map
func renameKeys(m map[string]interface{}, typ reflect.Value) map[string]interface{} {
	if typ.Kind() != reflect.Struct {
		return nil
	}
	renames := renameMap(typ)
	updated := make(map[string]interface{})
	for key, value := range m {
		into := renames[key]
		if len(into) > 0 {
			updated[into] = value
		} else {
			updated[key] = value
		}
	}
	return updated
}

// Matches the fields with `pson` to its field name.
//
// For example, A string `pson:"a"` -> map["a"] = "A"
func renameMap(v reflect.Value) map[string]string {
	mp := make(map[string]string)
	for i := 0; i < v.NumField(); i += 1 {
		typ := v.Type().Field(i)
		tag, found := typ.Tag.Lookup("pson")
		if found {
			name, _ := utils.ProcessTag(tag)
			mp[name] = typ.Name
		}
	}
	return mp
}
