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

func TestSetGlobalURL(t *testing.T) {
  t.Run("no call, nil by default", func(t *testing.T) {
    assert.Nil(t, base.url)
    assert.False(t, base.fragment)
    assert.True(t, base.empty)
  })

  t.Run("empty", func(t *testing.T) {
    SetGlobalURL("")
    assert.Nil(t, base.url)
    assert.False(t, base.fragment)
    assert.True(t, base.empty)
  })

  t.Run("empty, but with fragment", func(t *testing.T) {
    SetGlobalURL("", true)
    assert.Nil(t, base.url)
    assert.True(t, base.fragment)
    assert.True(t, base.empty)
  })

  t.Run("uncanny url", func(t *testing.T) {
    SetGlobalURL("foobar/")
    assert.Equal(t, "foobar", base.url.String())
    assert.False(t, base.fragment)
    assert.False(t, base.empty)
  })

  t.Run("good url, without fragment", func(t *testing.T) {
    SetGlobalURL("    \nhttps://github.com/////fontseca//////    \n")
    assert.Equal(t, "https://github.com/fontseca", base.url.String())
    assert.False(t, base.fragment)
    assert.False(t, base.empty)
  })

  t.Run("good url, with fragment", func(t *testing.T) {
    SetGlobalURL("\thttps://example.net//////problems/////\n", true)
    assert.Equal(t, "https://example.net/problems", base.url.String())
    assert.True(t, base.fragment)
    assert.False(t, base.empty)
  })

  t.Run("wrong url", func(t *testing.T) { // Sends an error to stderr.
    SetGlobalURL("postgres://user:abc{DEf1=foo@example.com:5432/db")
    assert.Nil(t, base.url)
    assert.False(t, base.fragment)
    assert.True(t, base.empty)
  })
}

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

func TestEmit(t *testing.T) {
  t.Run("Validation error problem", func(t *testing.T) {
    var expectedType = "https://example.net/validation-error"
    var expectedTitle = "Your request is not valid."
    var expectedStatus = http.StatusUnprocessableEntity
    var expectedError = []any{
      map[string]any{"detail": "must be a positive integer", "pointer": "#/age"},
      map[string]any{"detail": "must be 'green', 'red' or 'blue'", "pointer": "#/profile/color"},
    }

    var recorder = httptest.NewRecorder()

    var p Problem
    p.Type(Type(expectedType))
    p.Title(expectedTitle)
    p.Status(expectedStatus)
    p.With("error", expectedError[0])
    p.With("error", expectedError[1])
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

    var recorder = httptest.NewRecorder()
    var p Problem
    p.Type(Type(expectedType))
    p.Title(expectedTitle)
    p.Status(expectedStatus)
    p.Detail(expectedDetail)
    p.Instance(expectedInstance)
    p.With("balance", expectedBalance)
    p.With("accounts", expectedAccounts[0])
    p.With("accounts", expectedAccounts[1])
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

    var recorder = httptest.NewRecorder()
    var p Problem
    p.Status(expectedStatus)
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

func TestProblem(t *testing.T) {
  t.Run("Type", func(t *testing.T) {
    var p Problem

    t.Run("without invoking 'SetGlobalURL'", func(t *testing.T) {
      p.Type("")
      assert.Equal(t, "about:blank", p.typ)

      p.Type("www.foo.com/problems#out-of-credit")
      assert.Equal(t, "www.foo.com/problems#out-of-credit", p.typ)
    })

    t.Run("invoking 'SetGlobalURL', with empty url and fragment", func(t *testing.T) {
      SetGlobalURL("\n \t \n", true)

      p.Type("")
      assert.Equal(t, "about:blank", p.typ)

      p.Type("out-of-credit")
      assert.Equal(t, "out-of-credit", p.typ)
    })

    t.Run("with base URL, but no fragments", func(t *testing.T) {
      SetGlobalURL("https://www.example.com/")

      p.Type("")
      assert.Equal(t, "https://www.example.com", p.typ)

      p.Type("/path////to/////problems/out-of-credit")
      assert.Equal(t, "https://www.example.com/path/to/problems/out-of-credit", p.typ)
    })

    t.Run("with base URL and fragments", func(t *testing.T) {
      SetGlobalURL("https://www.example.com/problems", true)

      p.Type("")
      assert.Equal(t, "https://www.example.com/problems", p.typ)

      p.Type("out-of-credit")
      assert.Equal(t, "https://www.example.com/problems#out-of-credit", p.typ)
    })
  })

  t.Run("Status", func(t *testing.T) {
    var p Problem

    t.Run("invalid number, defaults to 200", func(t *testing.T) {
      p.Status(-1)
      assert.Equal(t, http.StatusOK, p.status)

      p.Status(1000)
      assert.Equal(t, http.StatusOK, p.status)
    })

    t.Run("correct status number", func(t *testing.T) {
      p.Status(http.StatusForbidden)
      assert.Equal(t, http.StatusForbidden, p.status)
    })
  })

  t.Run("Title", func(t *testing.T) {
    var p Problem

    t.Run("empty title", func(t *testing.T) {
      p.Title("\n\t \t\n")
      assert.Equal(t, "", p.title)
    })

    t.Run("good title", func(t *testing.T) {
      p.Title("You do not have enough credit.")
      assert.Equal(t, "You do not have enough credit.", p.title)
    })
  })

  t.Run("Detail", func(t *testing.T) {
    var p Problem

    t.Run("empty detail", func(t *testing.T) {
      p.Detail("\n\t \t\n")
      assert.Equal(t, "", p.detail)
    })

    t.Run("good detail", func(t *testing.T) {
      p.Detail("Your current balance is 30, but that costs 50.")
      assert.Equal(t, "Your current balance is 30, but that costs 50.", p.detail)
    })
  })

  t.Run("Instance", func(t *testing.T) {
    var p Problem

    t.Run("empty instance", func(t *testing.T) {
      p.Instance("\n\t \t\n")
      assert.Equal(t, "", p.instance)
    })

    t.Run("good instance", func(t *testing.T) {
      p.Instance("/account/12345/msgs/abc")
      assert.Equal(t, "/account/12345/msgs/abc", p.instance)
    })
  })

  t.Run("With", func(t *testing.T) {
    var p Problem
    p.With("", 1)
    p.With(" \n\t\n ", 2)
    p.With("chan", make(chan int))
    p.With("func", func() {})
    p.With("nil", nil)
    p.With("balance", 30)
    p.With("accounts", "/account/12345")
    p.With("accounts", "/account/67890")

    require.Len(t, p.extensions, 2)
    assert.Contains(t, p.extensions, map[string][]any{"balance": {30}}, "Extensions slice does not include 'balance' entry or entry's value is not correct.")
    assert.Contains(t, p.extensions, map[string][]any{"accounts": {"/account/12345", "/account/67890"}}, "Extensions slice does not include 'accounts' entry or entry's value is not correct.")
  })
}
