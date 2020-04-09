package url

import (
	"fmt"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"net/url"
	"strings"
)

// Create a URL with the given base, path parts, query string, and fragment. This method will properly URI encode
// everything and handle leading and trailing slashes.
func FormatUrl(baseUrl string, pathParts []string, query url.Values, fragment string) (string, error) {
	parsedUrl, err := url.Parse(stripSlashes(baseUrl))
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	normalizedPathParts := []string{}

	for _, pathPart := range pathParts {
		normalizedPathPart := stripSlashes(pathPart)
		normalizedPathParts = append(normalizedPathParts, normalizedPathPart)
	}

	if len(normalizedPathParts) > 0 {
		parsedUrl.Path = fmt.Sprintf("%s/%s", stripSlashes(parsedUrl.Path), strings.Join(normalizedPathParts, "/"))
	}

	parsedUrl.RawQuery = mergeQuery(parsedUrl.Query(), query).Encode()
	parsedUrl.Fragment = fragment

	return parsedUrl.String(), nil
}

// Merge the two query params together. The new query will override the original.
func mergeQuery(originalQuery url.Values, newQuery url.Values) url.Values {
	result := map[string][]string{}

	for key, values := range originalQuery {
		result[key] = values
	}

	for key, values := range newQuery {
		result[key] = values
	}

	return result
}

// Remove all leading or trailing slashes in the given string
func stripSlashes(str string) string {
	return strings.Trim(str, "/")
}
