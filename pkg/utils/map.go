package utils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
)

// func DeepSet(v map[string]interface{}, path []string, val interface{}) map[string]interface{} {
// 	key := path[0]
// 	if len(path) == 1 {
// 		v[key] = val
// 		return v
// 	}

// 	if next, ok := v[key]; ok {
// 		switch next.(type) {
// 		case map[string]interface{}:
// 			v[key] = DeepSet(next.(map[string]interface{}), path[1:], val)
// 			return v
// 		}
// 	}

// 	return v
// }

func CleanUpInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpInterfaceArray(in []interface{}) []interface{} {
	result := make([]interface{}, len(in))
	for i, v := range in {
		result[i] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanUpInterfaceArray(v)
	case map[interface{}]interface{}:
		return CleanUpInterfaceMap(v)
	case string:
		return v
	case bool:
		return v
	case int:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func Dedupe(l []string) []string {
	return sets.NewString(l...).List()
}

type SimpleType interface {
	string | int
}

func Map[T any, R SimpleType](slice []T, mapper func(elem T) R) []R {
	res := make([]R, 0)

	for _, elem := range slice {
		res = append(res, mapper(elem))
	}

	return res
}
