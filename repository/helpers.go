package repository

import (
  "errors"
  "fmt"
  "github.com/lib/pq"
  "strings"
)

// getErrMsg formats and returns a detailed error message from a given error.
// If the error is of type *pq.Error, it includes the message, detail, and hint.
// Otherwise, it returns the error's default message.
func getErrMsg(err error) string {
  var pqerr *pq.Error
  if errors.As(err, &pqerr) {
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("code: %s | message: %s",
      pqerr.Code,
      pqerr.Message))

    if "" != pqerr.Detail {
      builder.WriteString(fmt.Sprintf(" | detail: %s", pqerr.Detail))
    }

    if "" != pqerr.Hint {
      builder.WriteString(fmt.Sprintf(" | hint: %s", pqerr.Hint))
    }

    if "" != pqerr.Position {
      builder.WriteString(fmt.Sprintf(" | position: %s", pqerr.Position))
    }

    if "" != pqerr.Where {
      builder.WriteString(fmt.Sprintf(" | where: %s", pqerr.Where))
    }

    return builder.String()
  }

  return err.Error()
}
