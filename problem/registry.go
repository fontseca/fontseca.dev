package problem

import (
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
