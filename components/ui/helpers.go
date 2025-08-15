package ui

import (
  "fmt"
  "strings"
)

func fmtNumber(x int64) string {
  s := fmt.Sprintf("%d", x)
  var result strings.Builder

  length := len(s)
  for i, c := range s {
    if i != 0 && (length-i)%3 == 0 {
      result.WriteByte(',')
    }

    result.WriteByte(byte(c))
  }

  return result.String()
}
