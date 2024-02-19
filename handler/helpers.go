package handler

import (
  "encoding/json"
  "errors"
  "fontseca/problem"
  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
  "io"
  "log/slog"
  "net/http"
  "strings"
  "testing"
)

func marshal(t *testing.T, value any) []byte {
  var data, err = json.Marshal(value)
  if nil != err {
    t.Fatal(err)
  }
  return data
}

type failure struct {
  Field     string `json:"field"`
  Criterion string `json:"criterion"`
  Parameter string `json:"parameter,omitempty"`
}

func bindJSONRequestBody(c *gin.Context, obj any) (ok bool) {
  c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
  var err = c.ShouldBindJSON(obj)
  if nil != err {
    var syntaxError *json.SyntaxError
    var unmarshalTypeError *json.UnmarshalTypeError
    var validationErrors validator.ValidationErrors

    var b problem.Builder
    b.Status(http.StatusBadRequest)

    switch {
    default:
      slog.Error(err.Error())
      problem.NewInternalProblem().Emit(c.Writer)
      return
    case errors.Is(err, io.EOF):
      b.Title("Empty request body.")
      b.Detail("The request body must not be empty. Please provide a valid JSON object.")
    case strings.HasPrefix(err.Error(), "json: unknown field "):
      b.Title("Unexpected field in request body.")
      b.Detail("The request body contains an unexpected field. Please check the properties of the object.")
      field := strings.TrimPrefix(err.Error(), "json: unknown field \"")
      b.With("unexpected", strings.TrimSuffix(field, "\""))
    case errors.As(err, &validationErrors):
      b.Title("Invalid HTTP request body.")
      b.Status(http.StatusUnprocessableEntity)
      b.Detail("The provided JSON data does not meet the required validation criteria. Please review your input and try again.")
      var l = len(validationErrors)
      for _, e := range validationErrors {
        f := failure{
          Criterion: e.Tag(),
          Parameter: e.Param(),
          Field:     e.Field(),
        }
        if 1 == l {
          b.With("errors", []any{f})
        } else {
          b.With("errors", f)
        }
      }
    case errors.As(err, &syntaxError):
      b.Title("Malformed JSON in request body.")
      b.Detail("The request body contains invalid JSON syntax.")
      b.With("position", syntaxError.Offset)
      b.With("error", syntaxError.Error())
    case errors.As(err, &unmarshalTypeError):
      b.Title("Invalid value type in request body.")
      b.Detail("The request body contains a value that does not match the expected data type.")
      b.With("property", unmarshalTypeError.Field)
      b.With("has_type", unmarshalTypeError.Value)
      b.With("wants_type", unmarshalTypeError.Type.String())
    case errors.Is(err, io.ErrUnexpectedEOF):
      b.Title("Ill-formed JSON in request body.")
      b.Detail("The request body contains incomplete or truncated JSON data.")
    case err.Error() == "http: request body too large":
      b.Title("Request body too large.")
      b.Status(http.StatusRequestEntityTooLarge)
      b.Detail("The size of the request body must not exceed 1MB.")
    }
    b.Problem().Emit(c.Writer)
    return false
  }
  return true
}
