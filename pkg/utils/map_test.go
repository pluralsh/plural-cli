package utils_test

import (
	"testing"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestPatchInterfaceMap(t *testing.T) {
	tests := []struct {
		name                  string
		defaultValues, values map[string]map[string]interface{}
		expectedResult        map[string]map[string]interface{}
		expectError           bool
	}{
		{
			name: `test if map are equal`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			expectedResult: map[string]map[string]interface{}{},
		},
		{
			name: `test if new element was added`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": "d"},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": "d"},
			},
		},
		{
			name: `test if new nested element was added`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"d": "e"}},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": map[string]interface{}{"d": "e"}},
			},
		},
		{
			name: `test if new element was added to nested element`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"d": "e"}},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"f": "g"}},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": map[string]interface{}{"f": "g"}},
			},
		},
		{
			name: `test if nested element was added changed`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"d": "e"}},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"d": "f"}},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": map[string]interface{}{"d": "f"}},
			},
		},
		{
			name: `test if element was changed bool true to false`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": true},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test-2", "b": 14, "c": false},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"a": "test-2", "b": float64(14), "c": false},
			},
		},
		{
			name: `test if element was changed bool false to true`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": false},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test-2", "b": 14, "c": true},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"a": "test-2", "b": float64(14), "c": true},
			},
		},
		{
			name: `test with equal list element`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": []interface{}{"a", "b"}},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": []interface{}{"a", "b"}},
			},
			expectedResult: map[string]map[string]interface{}{},
		},
		{
			name: `test with changed list element`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": []interface{}{"a", "b"}},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": []interface{}{"c", "d"}},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": []interface{}{"c", "d"}},
			},
		},
		{
			name: `test with removing element from list`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": []interface{}{"a", "b"}},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": []interface{}{"a"}},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": []interface{}{"a"}},
			},
		},
		{
			name: `test with empty values`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values:         map[string]map[string]interface{}{},
			expectedResult: map[string]map[string]interface{}{},
		},
		{
			name: `test with empty nested values`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"d": "test"}},
			},
			values: map[string]map[string]interface{}{
				"test": {"c": map[string]interface{}{}},
			},
			expectedResult: map[string]map[string]interface{}{},
		},
		{
			name: `test with empty top level map`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": map[string]interface{}{"d": "test"}},
			},
			values: map[string]map[string]interface{}{
				"test": {},
			},
			expectedResult: map[string]map[string]interface{}{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := utils.PatchInterfaceMap(test.defaultValues, test.values)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, res)
		})
	}
}
