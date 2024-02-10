package problem

import (
  "net/http"
  "strings"
  "unicode"
)

// snakeToPascalCase converts a string from snake_case format to PascalCase.
// It capitalizes the first letter of each word separated by underscores and
// converts  the rest of the word to lowercase. Note: This implementation may
// not handle non-ASCII characters.
func snakeToPascalCase(s string) string {
  if "" == s {
    return ""
  }
  var f = func(c rune) bool { return '_' == c || unicode.IsSpace(c) }
  var words = strings.FieldsFunc(s, f)
  var builder strings.Builder
  for _, word := range words {
    builder.WriteRune(unicode.ToTitle(rune(word[0])))
    builder.WriteString(strings.ToLower(word[1:]))
  }
  return builder.String()
}

// canonicalSnakeCase returns the canonical snake_case format of the input string s.
// It converts the string s to lowercase and replaces any whitespace with underscores,
// ensuring it adheres to the canonical standard snake_case format. It does not
// perform any additional transformations; it assumes the input string is already in
// a non-canonical form of snake_case, such as "Snake_Case" or "snake case".
func canonicalSnakeCase(s string) string {
  if "" == s {
    return ""
  }
  return strings.ToLower(strings.Join(strings.Fields(s), "_"))
}

// Extension is a map that stores all the additional members that are
// specific to a problem type.
type Extension map[string][]any

// Add adds a new extension to the additional members of the problem.
func (e Extension) Add(key string, value any) {
  e[canonicalSnakeCase(key)] = append(e[canonicalSnakeCase(key)], value)
}

// Set sets the value of the extension for the specified key,
// replacing any existing values.
func (e Extension) Set(key string, value any) {
  e[canonicalSnakeCase(key)] = []any{value}
}

// Del deletes the extension for the specified key.
func (e Extension) Del(key string) {
  delete(e, canonicalSnakeCase(key))
}

// isValidHTTPStatusCode checks if the provided integer code is a valid HTTP status code.
// HTTP status codes are considered valid if they fall within the range of 100 to 599.
func isValidHTTPStatusCode(code int) bool {
  return code >= 100 && code <= 999
}

// Problem represents an RFC 9457 Problem Details object. When serialized in a JSON
// object, this format is identified with the "application/problem+json" media type.
// For more details, refer to RFC 9457: https://www.rfc-editor.org/rfc/rfc9457#name-the-problem-details-json-ob.
// It also implements the error interface for straightforward mobilization.
type Problem interface {
  error

  // Extension returns the Extension object associated with the
  // problem. If the Extension object is nil, then it's created.
  Extension() Extension

  // Emit sends the problem details as an HTTP response through the
  // provided http.ResponseWriter w. The response body is a JSON object
  // and its Content-Type is "application/problem+json".
  Emit(w http.ResponseWriter)
}
