package problem

import (
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
