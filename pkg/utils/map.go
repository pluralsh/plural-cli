package utils

import (
	"fmt"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/exp/maps"
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

type DiffCondition func(key string, value, diffValue any) bool

var (
	equalDiffCondition DiffCondition = func(_ string, value, diffValue any) bool {
		return cmp.Equal(value, diffValue)
	}
)

// DiffMap removes keys from the base map based on provided DiffCondition match against the same keys in provided
// diff map. It always uses an equal comparison for the values, but specific keys can use extended comparison
// if needed by passing custom DiffCondition function.
//
// Example:
//  A: {a: 1, b: 1, c: 2}
//  B: {a: 1, d: 2, c: 3}
//  Result: {b: 1, c: 2}
//
// Note: It does not remove null value keys by default.
func DiffMap(base, diff map[string]interface{}, conditions ...DiffCondition) map[string]interface{} {
	result := make(map[string]interface{})
	maps.Copy(result, base)

	if diff == nil {
		diff = make(map[string]interface{})
	}

	for k, v := range base {
		switch v.(type) {
		case map[string]interface{}:
			dValue, _ := diff[k].(map[string]interface{})
			if dMap := DiffMap(v.(map[string]interface{}), dValue, conditions...); len(dMap) > 0 {
				result[k] = dMap
				break
			}

			delete(result, k)
		default:
			diffV, _ := diff[k]
			for _, condition := range append(conditions, equalDiffCondition) {
				if condition(k, v, diffV) {
					delete(result, k)
					break
				}
			}
		}
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
