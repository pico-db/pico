package umap

import (
	"fmt"
	"sort"
	"strings"
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
