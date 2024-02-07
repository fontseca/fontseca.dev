package main

import (
  "bytes"
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestAdorn_Write(t *testing.T) {
  var (
    buffer   = new(bytes.Buffer)
    w        = adorn{buffer}
    expected = `{
  "level": "ERROR",
  "time": "2006-01-02T15:04:05Z07:00",
  "msg": "error",
  "source": {
    "file": "foo.go",
    "function": "Foo",
    "line": 1
  }
}
`
  )
  var n, err = w.Write([]byte(`{"time":"2006-01-02T15:04:05Z07:00","level":"ERROR","source":{"function":"Foo","file":"foo.go","line":1},"msg":"error"}`))
  assert.NoError(t, err)
  assert.Equal(t, len(expected)-1, n)
  assert.Equal(t, expected, buffer.String())
}
