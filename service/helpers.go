package service

import (
  "fontseca/problem"
  "github.com/google/uuid"
  "net/http"
  "net/url"
  "strings"
)

func sanitizeURL(urls ...*string) error {
  if 0 == len(urls) {
    return nil
  }

  for _, u := range urls {
    if nil == u {
      continue
    }

    *u = strings.TrimSpace(*u)
    if "" == *u {
      continue
    }

    uri, err := url.ParseRequestURI(*u)
    if nil != err {
      var p problem.Problem
      p.Title("Unprocessable URL format.")
      p.Status(http.StatusUnprocessableEntity)
      p.Detail("There was an error parsing the requested URL. Please try with a different URL or verify the current one for correctness.")
      p.With("wrong_url", *u)
      return &p
    }
    *u = uri.String()
  }

  return nil
}

func validateUUID(id *string) error {
  if nil == id {
    return nil
  }
  *id = strings.TrimSpace(*id)
  parsed, err := uuid.Parse(*id)
  if nil != err {
    var p problem.Problem
    p.Title("Could not parse UUID.")
    p.Status(http.StatusUnprocessableEntity)
    p.Detail("An error occurred while attempting to parse the provided UUID string.")
    p.With("uuid", *id)
    p.With("error", err.Error())
    *id = ""
    return &p
  }
  *id = parsed.String()
  return nil
}
