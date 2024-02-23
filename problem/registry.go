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
