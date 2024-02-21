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

type notFoundProblem struct {
  problem
}

func NewNotFoundProblem(id, recordType string) Problem {
  return &notFoundProblem{
    problem{
      typ:      "about:blank",
      status:   http.StatusNotFound,
      title:    "Record not found.",
      detail:   fmt.Sprintf("The %s record with ID '%s' could not be found in the database.", recordType, id),
      instance: "",
    },
  }
}
