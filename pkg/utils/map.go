package utils

import (
	"fmt"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch"
	jsoniter "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/util/sets"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func CleanUpInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
	}
	return result
}

func RemoveNulls(m map[string]interface{}) {
	val := reflect.ValueOf(m)
	for _, e := range val.MapKeys() {
		v := val.MapIndex(e)
		if v.IsNil() {
			delete(m, e.String())
			continue
		}

		t, ok := v.Interface().(map[string]interface{})
		if ok {
			RemoveNulls(t)
		}
		// if the map is empty, remove it
		// TODO: add a unit test for this
		if ok && len(t) == 0 {
			delete(m, e.String())
		}
	}
}

func MergeMap(defaultValues, values map[string]interface{}) (map[string]interface{}, error) {
	defaultJson, err := json.Marshal(defaultValues)
	if err != nil {
		return nil, err
	}
	valuesJson, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	patchJson, err := jsonpatch.MergePatch(defaultJson, valuesJson)
	if err != nil {
		return nil, err
	}

	patch := map[string]interface{}{}
	if err := json.Unmarshal(patchJson, &patch); err != nil {
		return nil, err
	}

	return patch, nil
}

func PatchInterfaceMap(defaultValues, values map[string]map[string]interface{}) (map[string]map[string]interface{}, error) {
	defaultJson, err := json.Marshal(defaultValues)
	if err != nil {
		return nil, err
	}
	valuesJson, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	patchJson, err := jsonpatch.CreateMergePatch(defaultJson, valuesJson)
	if err != nil {
		return nil, err
	}

	patch := map[string]map[string]interface{}{}
	if err := json.Unmarshal(patchJson, &patch); err != nil {
		return nil, err
	}
	for key := range patch {
		// if the map is empty, remove it
		if len(patch[key]) == 0 {
			delete(patch, key)
		} else {
			// remove nulls from the map
			RemoveNulls(patch[key])
			// if the map is empty after removing nulls, remove it
			if len(patch[key]) == 0 {
				delete(patch, key)
			}
		}
	}
	// if the patch is empty, return an empty map
	if len(patch) == 0 {
		return map[string]map[string]interface{}{}, nil
	}
	return patch, nil
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
