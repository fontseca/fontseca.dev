package problem

import (
  "encoding/json"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "io"
  "net/http"
  "net/http/httptest"
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

func TestEmit(t *testing.T) {
  t.Run("Validation error problem", func(t *testing.T) {
    var expectedType = "https://example.net/validation-error"
    var expectedTitle = "Your request is not valid."
    var expectedStatus = http.StatusUnprocessableEntity
    var expectedError = []any{
      map[string]any{"detail": "must be a positive integer", "pointer": "#/age"},
      map[string]any{"detail": "must be 'green', 'red' or 'blue'", "pointer": "#/profile/color"},
    }

    var p = problem{typ: expectedType, title: expectedTitle, status: expectedStatus}
    p.Extension().Add("error", expectedError[0])
    p.Extension().Add("error", expectedError[1])

    var recorder = httptest.NewRecorder()
    p.Emit(recorder)

    var response = recorder.Result()
    defer response.Body.Close()

    var body, err = io.ReadAll(response.Body)
    require.NoError(t, err, "Failed reading JSON response body.")

    require.NotEmpty(t, body, "Response body must not be empty.")
    assert.Contains(t, response.Header["Content-Type"], "application/problem+json")

    var object map[string]any
    err = json.Unmarshal(body, &object)
    require.NoError(t, err, "Unmarshalling JSON response body failed.")

    assert.Equal(t, expectedStatus, response.StatusCode)
    assert.Equal(t, expectedType, object["type"])
    assert.Equal(t, expectedStatus, int(object["status"].(float64)))
    assert.Equal(t, expectedTitle, object["title"])
    assert.Equal(t, expectedError, object["error"])
    assert.NotContains(t, object, "detail", "Superfluous \"detail\" field in JSON body response.")
    assert.NotContains(t, object, "instance", "Superfluous \"instance\" field in JSON body  response.")
  })

  t.Run("Out-of-credit credit problem", func(t *testing.T) {
    var expectedType = "https://example.com/probs/out-of-credit"
    var expectedTitle = "You do not have enough credit."
    var expectedStatus = http.StatusForbidden
    var expectedDetail = "Your current balance is 30, but that costs 50."
    var expectedInstance = "/account/12345/msgs/abc"
    var expectedBalance = 30
    var expectedAccounts = []any{
      "/account/12345",
      "/account/67890",
    }

    var p = problem{typ: expectedType, title: expectedTitle, status: expectedStatus, detail: expectedDetail, instance: expectedInstance}
    p.Extension().Add("balance", expectedBalance)
    p.Extension().Add("accounts", expectedAccounts[0])
    p.Extension().Add("accounts", expectedAccounts[1])

    var recorder = httptest.NewRecorder()
    p.Emit(recorder)

    var response = recorder.Result()
    defer response.Body.Close()

    var body, err = io.ReadAll(response.Body)
    require.NoError(t, err, "Failed reading JSON response body.")

    require.NotEmpty(t, body, "Response body must not be empty.")
    assert.Contains(t, response.Header["Content-Type"], "application/problem+json")

    var object map[string]any
    err = json.Unmarshal(body, &object)
    require.NoError(t, err, "Unmarshalling JSON response body failed.")

    assert.Equal(t, expectedStatus, response.StatusCode)
    assert.Equal(t, expectedType, object["type"])
    assert.Equal(t, expectedStatus, int(object["status"].(float64)))
    assert.Equal(t, expectedTitle, object["title"])
    assert.Equal(t, expectedDetail, object["detail"])
    assert.Equal(t, expectedInstance, object["instance"])
    assert.Equal(t, expectedBalance, int(object["balance"].(float64)))
    assert.Equal(t, expectedAccounts, object["accounts"])
  })

  t.Run("Uncanny problem", func(t *testing.T) {
    var expectedType = "about:blank"
    var expectedStatus = http.StatusSeeOther
    var expectedTitle = http.StatusText(expectedStatus)

    var p = problem{status: expectedStatus}
    var recorder = httptest.NewRecorder()
    p.Emit(recorder)

    var response = recorder.Result()
    defer response.Body.Close()

    var body, err = io.ReadAll(response.Body)
    require.NoError(t, err, "Failed reading JSON response body.")

    require.NotEmpty(t, body, "Response body must not be empty.")
    assert.Contains(t, response.Header["Content-Type"], "application/problem+json")

    var object map[string]any
    err = json.Unmarshal(body, &object)
    require.NoError(t, err, "Unmarshalling JSON response body failed.")

    assert.Equal(t, expectedStatus, response.StatusCode)
    assert.Equal(t, expectedType, object["type"])
    assert.Equal(t, expectedStatus, int(object["status"].(float64)))
    assert.Equal(t, expectedTitle, object["title"])
    assert.NotContains(t, object, "detail", "Superfluous \"detail\" field in JSON body response.")
    assert.NotContains(t, object, "instance", "Superfluous \"instance\" field in JSON body response.")
    assert.NotContains(t, object, "balance", "Superfluous \"balance\" field in JSON body response.")
  })
}
