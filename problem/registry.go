package problem

import (
  "fmt"
  "net/http"
)

func NewInternal() *Problem {
  var p Problem
  p.Type("about:blank")
  p.Status(http.StatusInternalServerError)
  p.Title("Internal Server Error.")
  p.Detail("An unexpected error occurred while processing your request. Please try again later. If the problem persists, contact the developer for assistance.")
  p.With("contact", "mailto:fontseca.dev@outlook.com")
  return &p
}

func NewNotFound(id, recordType string) *Problem {
  var p Problem
  p.Type("about:blank")
  p.Status(http.StatusNotFound)
  p.Title("Record not found.")
  p.Detail(fmt.Sprintf("The %s record with ID '%s' could not be found in the database.", recordType, id))
  p.With("record_id", id)
  p.With("record_type", recordType)
  return &p
}

func NewSlugNotFound(slug, recordType string) *Problem {
  var p Problem
  p.Type("about:blank")
  p.Status(http.StatusNotFound)
  p.Title("Record not found.")
  p.Detail(fmt.Sprintf("The %s record with the slug '%s' could not be found in the database.", recordType, slug))
  p.With("slug", slug)
  p.With("record_type", recordType)
  return &p
}

func NewUnparsableValue(targetType, fieldName, fieldValue string) *Problem {
  var p Problem
  p.Type("about:blank")
  p.Status(http.StatusUnprocessableEntity)
  p.Title(fmt.Sprintf("Failure when parsing %s value.", targetType))
  p.Detail(fmt.Sprintf("Failed to parse the provided value as: %s. Please make sure the value is valid according to its type.", targetType))
  p.With("field_name", fieldName)
  p.With("field_type", targetType)
  p.With("field_value", fieldValue)
  return &p
}

func NewValueOutOfRange(targetType, fieldName, fieldValue string) *Problem {
  var p Problem
  p.Type("about:blank")
  p.Status(http.StatusUnprocessableEntity)
  p.Title(fmt.Sprintf("Failure when parsing %s value.", targetType))
  p.Detail(fmt.Sprintf("Out of range for the provided value as: %s. Please make sure the value is valid according to its type.", targetType))
  p.Detail(fmt.Sprintf("The value provided for '%s' (%s) is out of range for the specified type (%s). Please ensure the value falls within the acceptable range.", fieldName, fieldValue, targetType))
  p.With("field_name", fieldName)
  p.With("field_type", targetType)
  p.With("field_value", fieldValue)
  return &p
}

func NewValidation(failures ...[3]string) *Problem {
  var p Problem
  p.Status(http.StatusBadRequest)
  p.Title("Failed to validate request data.")
  p.Detail("The provided data does not meet the required validation criteria. Please review your input and try again.")

  for _, f := range failures {
    if "" != f[2] {
      p.With("errors", map[string]string{
        "field":     f[0],
        "criterion": f[1],
        "parameter": f[2],
      })
    } else {
      p.With("errors", map[string]string{
        "field":     f[0],
        "criterion": f[1],
      })
    }
  }

  return &p
}

func NewMissingParameter(parameter string) *Problem {
  var p Problem
  p.Status(http.StatusBadRequest)
  p.Title("Missing required parameter.")
  p.Detail(fmt.Sprintf("The '%s' parameter is required but was not found in the request form data.", parameter))
  p.With("missing_parameter", parameter)
  return &p
}
