package handler

import (
  "encoding/json"
  "errors"
  "fontseca/problem"
  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
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

    var p problem.Problem
    p.Status(http.StatusBadRequest)

    switch {
    default:
      slog.Error(err.Error())
      problem.NewInternal().Emit(c.Writer)
      return
    case errors.Is(err, io.EOF):
      p.Title("Empty request body.")
      p.Detail("The request body must not be empty. Please provide a valid JSON object.")
    case strings.HasPrefix(err.Error(), "json: unknown field "):
      p.Title("Unexpected field in request body.")
      p.Detail("The request body contains an unexpected field. Please check the properties of the object.")
      field := strings.TrimPrefix(err.Error(), "json: unknown field \"")
      p.With("unexpected", strings.TrimSuffix(field, "\""))
    case errors.As(err, &validationErrors):
      p.Title("Invalid HTTP request body.")
      p.Status(http.StatusUnprocessableEntity)
      p.Detail("The provided JSON data does not meet the required validation criteria. Please review your input and try again.")
      var l = len(validationErrors)
      for _, e := range validationErrors {
        f := failure{
          Criterion: e.Tag(),
          Parameter: e.Param(),
          Field:     e.Field(),
        }
        if 1 == l {
          p.With("errors", []any{f})
        } else {
          p.With("errors", f)
        }
      }
    case errors.As(err, &syntaxError):
      p.Title("Malformed JSON in request body.")
      p.Detail("The request body contains invalid JSON syntax.")
      p.With("position", syntaxError.Offset)
      p.With("error", syntaxError.Error())
    case errors.As(err, &unmarshalTypeError):
      p.Title("Invalid value type in request body.")
      p.Detail("The request body contains a value that does not match the expected data type.")
      p.With("property", unmarshalTypeError.Field)
      p.With("has_type", unmarshalTypeError.Value)
      p.With("wants_type", unmarshalTypeError.Type.String())
    case errors.Is(err, io.ErrUnexpectedEOF):
      p.Title("Ill-formed JSON in request body.")
      p.Detail("The request body contains incomplete or truncated JSON data.")
    case err.Error() == "http: request body too large":
      p.Title("Request body too large.")
      p.Status(http.StatusRequestEntityTooLarge)
      p.Detail("The size of the request body must not exceed 1MB.")
    }
    p.Emit(c.Writer)
    return false
  }
  return true
}

func validateStruct(s any) error {
  var err = binding.Validator.ValidateStruct(s)
  if nil != err {
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
      var p problem.Problem
      p.Title("Failed to validate request data.")
      p.Status(http.StatusUnprocessableEntity)
      p.Detail("The provided data does not meet the required validation criteria. Please review your input and try again.")
      var l = len(validationErrors)
      for _, e := range validationErrors {
        f := failure{
          Criterion: e.Tag(),
          Parameter: e.Param(),
          Field:     e.Field(),
        }
        if 1 == l {
          p.With("errors", []any{f})
        } else {
          p.With("errors", f)
        }
      }
      err = &p
    } else {
      slog.Error(err.Error())
    }
  }
  return err
}
