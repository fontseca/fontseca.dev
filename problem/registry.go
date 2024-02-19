package problem

import (
  "net/http"
)

type internalProblem struct {
  problem
}

func NewInternalProblem() Problem {
  return &internalProblem{
    problem: problem{
      typ:        "about:blank",
      status:     http.StatusInternalServerError,
      title:      "Internal Server Error.",
      detail:     "An unexpected error occurred while processing your request. Please try again later. If the problem persists, contact the developer for assistance.",
      instance:   "",
      extensions: []map[string][]any{{"contact": {"mailto:fontseca.dev@outlook.com"}}},
    },
  }
}
