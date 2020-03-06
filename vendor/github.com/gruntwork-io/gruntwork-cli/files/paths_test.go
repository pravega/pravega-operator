package files

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"regexp"
)

func TestGetPathRelativeTo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		path	 string
		basePath string
		expected string
	}{
		{"", "", "."},
		{"/root", "/root", "."},
		{"/root", "/root/child", ".."},
		{"/root", "/root/child/sub-child/sub-sub-child", "../../.."},
		{"/root/other-child", "/root/child", "../other-child"},
		{"/root/other-child/sub-child", "/root/child/sub-child", "../../other-child/sub-child"},
		{"/root", "/other-root", "../root"},
		{"/root", "/other-root/sub-child/sub-sub-child", "../../../root"},
	}

	for _, testCase := range testCases {
		actual, err := GetPathRelativeTo(testCase.path, testCase.basePath)
		assert.Nil(t, err, "Unexpected error for path %s and basePath %s: %v", testCase.path, testCase.basePath, err)
		assert.Equal(t, testCase.expected, actual, "For path %s and basePath %s", testCase.path, testCase.basePath)
	}
}

func TestCanonicalPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		path	 string
		basePath string
		expected string
	}{
		{"", "/foo", "/foo"},
		{".", "/foo", "/foo"},
		{"bar", "/foo", "/foo/bar"},
		{"bar/baz/blah", "/foo", "/foo/bar/baz/blah"},
		{"bar/../blah", "/foo", "/foo/blah"},
		{"bar/../..", "/foo", "/"},
		{"bar/.././../baz", "/foo", "/baz"},
		{"bar", "/foo/../baz", "/baz/bar"},
		{"a/b/../c/d/..", "/foo/../baz/.", "/baz/a/c"},
		{"/other", "/foo", "/other"},
		{"/other/bar/blah", "/foo", "/other/bar/blah"},
		{"/other/../blah", "/foo", "/blah"},
	}

	for _, testCase := range testCases {
		actual, err := CanonicalPath(testCase.path, testCase.basePath)
		assert.Nil(t, err, "Unexpected error for path %s and basePath %s: %v", testCase.path, testCase.basePath, err)
		assert.Equal(t, testCase.expected, actual, "For path %s and basePath %s", testCase.path, testCase.basePath)
	}
}

func TestGrepFindsText(t *testing.T) {
	t.Parallel()

	regex, err := regexp.Compile("Hello, World!")
	if err != nil {
		t.Fatal(err)
	}

	found, err := Grep(regex, "../fixtures/files/**/*")
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.True(t, found)
}

func TestGrepFindsTextInSubfolder(t *testing.T) {
	t.Parallel()

	regex, err := regexp.Compile("f.+bar.+z")
	if err != nil {
		t.Fatal(err)
	}

	found, err := Grep(regex, "../fixtures/files/**/*")
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.True(t, found)
}

func TestGrepDoesNotFindInvalidText(t *testing.T) {
	t.Parallel()

	regex, err := regexp.Compile("this-should-not-be-found")
	if err != nil {
		t.Fatal(err)
	}

	found, err := Grep(regex, "../fixtures/files/**/*")
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.False(t, found)
}