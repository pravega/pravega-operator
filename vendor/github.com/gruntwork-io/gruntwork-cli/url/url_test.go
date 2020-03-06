package url

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestFormatUrl(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		baseUrl  string
		parts    []string
		query    url.Values
		fragment string
		expected string
	}{
		{"base-url-only", "http://www.example.com", parts(), query0(), "", "http://www.example.com"},
		{"base-url-with-one-part", "http://www.example.com", parts("foo"), query0(), "", "http://www.example.com/foo"},
		{"base-url-with-multiple-parts", "http://www.example.com", parts("foo", "bar", "baz"), query0(), "", "http://www.example.com/foo/bar/baz"},
		{"base-url-with-trailing-slash-with-multiple-parts", "http://www.example.com/", parts("foo", "bar", "baz"), query0(), "", "http://www.example.com/foo/bar/baz"},
		{"base-url-with-trailing-slash-with-multiple-parts-with-slashes", "http://www.example.com/", parts("/foo", "bar/", "//baz///"), query0(), "", "http://www.example.com/foo/bar/baz"},
		{"base-url-with-path-with-multiple-parts", "http://www.example.com/a/b/c", parts("foo", "bar", "baz"), query0(), "", "http://www.example.com/a/b/c/foo/bar/baz"},
		{"base-url-with-query-string-with-multiple-parts", "http://www.example.com/?a=b&c=d", parts("foo", "bar", "baz"), query0(), "", "http://www.example.com/foo/bar/baz?a=b&c=d"},
		{"base-url-with-multiple-parts-encoding", "http://www.example.com", parts("foo a b c", "?#$@!"), query0(), "", "http://www.example.com/foo%20a%20b%20c/%3F%23$@%21"},
		{"base-url-with-one-query-param", "http://www.example.com", parts(), query1("foo", "bar"), "", "http://www.example.com?foo=bar"},
		{"base-url-with-one-query-param-encoding", "http://www.example.com", parts(), query1("foo a b c", "?#$@!"), "", "http://www.example.com?foo+a+b+c=%3F%23%24%40%21"},
		{"base-url-with-fragment", "http://www.example.com", parts(), query0(), "foo", "http://www.example.com#foo"},
		{"base-url-with-path-query-fragment", "http://www.example.com", parts("foo", "bar"), query1("baz", "blah"), "fragment", "http://www.example.com/foo/bar?baz=blah#fragment"},
	}

	for _, testCase := range testCases {
		// Store a copy in scope so all the test cases don't end up running the last item in the loop
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			actual, err := FormatUrl(testCase.baseUrl, testCase.parts, testCase.query, testCase.fragment)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func parts(p ... string) []string {
	return p
}

func query0() url.Values {
	return map[string][]string{}
}

func query1(key string, value string) url.Values {
	return map[string][]string{
		key: {value},
	}
}

func query2(key1 string, value1 string, key2 string, value2 string) url.Values {
	return map[string][]string{
		key1: {value1},
		key2: {value2},
	}
}

func query3(key1 string, value1 string, key2 string, value2 string, key3 string, value3 string) url.Values {
	return map[string][]string{
		key1: {value1},
		key2: {value2},
		key3: {value3},
	}
}
