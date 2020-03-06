package collections

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeCopyOfListMakesACopy(t *testing.T) {
	original := []string{"foo", "bar", "baz"}
	copyOfList := MakeCopyOfList(original)
	assert.Equal(t, original, copyOfList)
}

func TestListContainsElement(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		list     []string
		element  string
		expected bool
	}{
		{[]string{}, "", false},
		{[]string{}, "foo", false},
		{[]string{"foo"}, "foo", true},
		{[]string{"bar", "foo", "baz"}, "foo", true},
		{[]string{"bar", "foo", "baz"}, "nope", false},
		{[]string{"bar", "foo", "baz"}, "", false},
	}

	for _, testCase := range testCases {
		actual := ListContainsElement(testCase.list, testCase.element)
		assert.Equal(t, testCase.expected, actual, "For list %v and element %s", testCase.list, testCase.element)
	}
}

func TestRemoveElementFromList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		list     []string
		element  string
		expected []string
	}{
		{[]string{}, "", []string{}},
		{[]string{}, "foo", []string{}},
		{[]string{"foo"}, "foo", []string{}},
		{[]string{"bar"}, "foo", []string{"bar"}},
		{[]string{"bar", "foo", "baz"}, "foo", []string{"bar", "baz"}},
		{[]string{"bar", "foo", "baz"}, "nope", []string{"bar", "foo", "baz"}},
		{[]string{"bar", "foo", "baz"}, "", []string{"bar", "foo", "baz"}},
	}

	for _, testCase := range testCases {
		actual := RemoveElementFromList(testCase.list, testCase.element)
		assert.Equal(t, testCase.expected, actual, "For list %v and element %s", testCase.list, testCase.element)
	}
}

func TestBatchListIntoGroupsOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		stringList []string
		n          int
		result     [][]string
	}{
		{
			[]string{"macaroni", "gentoo", "magellanic", "adelie", "little", "king", "emperor"},
			2,
			[][]string{
				[]string{"macaroni", "gentoo"},
				[]string{"magellanic", "adelie"},
				[]string{"little", "king"},
				[]string{"emperor"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic", "adelie", "king", "emperor"},
			2,
			[][]string{
				[]string{"macaroni", "gentoo"},
				[]string{"magellanic", "adelie"},
				[]string{"king", "emperor"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			5,
			[][]string{
				[]string{"macaroni", "gentoo", "magellanic"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			5,
			[][]string{
				[]string{"macaroni", "gentoo", "magellanic"},
			},
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			-1,
			nil,
		},
		{
			[]string{"macaroni", "gentoo", "magellanic"},
			0,
			nil,
		},
		{
			[]string{},
			7,
			[][]string{},
		},
	}

	for idx, testCase := range testCases {
		t.Run(fmt.Sprintf("%s_%d", t.Name(), idx), func(t *testing.T) {
			t.Parallel()
			original := MakeCopyOfList(testCase.stringList)
			assert.Equal(t, BatchListIntoGroupsOf(testCase.stringList, testCase.n), testCase.result)
			// Make sure the function doesn't modify the original list
			assert.Equal(t, testCase.stringList, original)
		})
	}
}
