package problem

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestSnakeToPascalCase(t *testing.T) {
  var cases = map[string]string{
    "snake_case":            "SnakeCase",
    "snake":                 "Snake",
    "PASCAL_SNAKE_CASE":     "PascalSnakeCase",
    "1_st_user_name":        "1StUserName",
    "josé_zúñiga":           "JoséZúñiga",
    "\t\n HELLO_WORLD \t\n": "HelloWorld",
    "":                      "",
  }
  for input, expected := range cases {
    var output = snakeToPascalCase(input)
    assert.Equal(t, expected, output)
  }
}

func TestCanonicalSnakeCase(t *testing.T) {
  var cases = map[string]string{
    "\tsnake\ncase\n ":   "snake_case",
    "first name":         "first_name",
    "Snake_case":         "snake_case",
    "HELLO_WORLD":        "hello_world",
    "Pascal Case String": "pascal_case_string",
  }
  for input, expected := range cases {
    var output = canonicalSnakeCase(input)
    assert.Equal(t, expected, output)
  }
}

func TestExtension(t *testing.T) {
  var extension = Extension{}

  t.Run("Add", func(t *testing.T) {
    extension.Add("Key_Name", "value1")
    extension.Add("KEY_NAME", "value2")
    assert.Equal(t, Extension{"key_name": {"value1", "value2"}}, extension)
  })

  t.Run("Set", func(t *testing.T) {
    extension.Set("Key_Name", "value1")
    assert.Equal(t, Extension{"key_name": {"value1"}}, extension)
  })

  t.Run("Del", func(t *testing.T) {
    extension.Del("Key_Name")
    assert.Equal(t, Extension{}, extension)
  })
}
