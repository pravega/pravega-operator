package collections

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestMergeMapsNoMaps(t *testing.T) {
	t.Parallel()

	expected := map[string]interface{}{}
	assert.Equal(t, expected, MergeMaps())
}

func TestMergeMapsOneEmptyMap(t *testing.T) {
	t.Parallel()

	map1 := map[string]interface{}{}

	expected := map[string]interface{}{}

	assert.Equal(t, expected, MergeMaps(map1))
}

func TestMergeMapsMultipleEmptyMaps(t *testing.T) {
	t.Parallel()

	map1 := map[string]interface{}{}
	map2 := map[string]interface{}{}
	map3 := map[string]interface{}{}

	expected := map[string]interface{}{}

	assert.Equal(t, expected, MergeMaps(map1, map2, map3))
}

func TestMergeMapsOneNonEmptyMap(t *testing.T) {
	t.Parallel()

	map1 := map[string]interface{}{
		"key1": "value1",
	}

	expected := map[string]interface{}{
		"key1": "value1",
	}

	assert.Equal(t, expected, MergeMaps(map1))
}

func TestMergeMapsTwoNonEmptyMaps(t *testing.T) {
	t.Parallel()

	map1 := map[string]interface{}{
		"key1": "value1",
	}

	map2 := map[string]interface{}{
		"key2": "value2",
	}

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	assert.Equal(t, expected, MergeMaps(map1, map2))
}

func TestMergeMapsTwoNonEmptyMapsOverlappingKeys(t *testing.T) {
	t.Parallel()

	map1 := map[string]interface{}{
		"key1": "value1",
		"key3": "value3",
	}

	map2 := map[string]interface{}{
		"key1": "replacement",
		"key2": "value2",
	}

	expected := map[string]interface{}{
		"key1": "replacement",
		"key2": "value2",
		"key3": "value3",
	}

	assert.Equal(t, expected, MergeMaps(map1, map2))
}

func TestMergeMapsMultipleNonEmptyMapsOverlappingKeys(t *testing.T) {
	t.Parallel()

	map1 := map[string]interface{}{
		"key1": "value1",
		"key3": "value3",
	}

	map2 := map[string]interface{}{
		"key1": "replacement",
		"key2": "value2",
	}

	map3 := map[string]interface{}{
		"key1": "replacement-two",
		"key3": "replacement-two",
		"key4": "value4",
	}

	expected := map[string]interface{}{
		"key1": "replacement-two",
		"key2": "value2",
		"key3": "replacement-two",
		"key4": "value4",
	}

	assert.Equal(t, expected, MergeMaps(map1, map2, map3))
}

func TestKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input	 map[string]string
		expected []string
	}{
		{map[string]string{}, []string{}},
		{map[string]string{"a": "foo"}, []string{"a"}},
		{map[string]string{"a": "foo", "b": "bar", "c": "baz"}, []string{"a", "b", "c"}},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			actual := Keys(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}