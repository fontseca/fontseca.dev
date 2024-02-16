package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
  "net/url"
  "strings"
)

type MeHandler struct {
  s service.MeService
}

func NewMeHandler(s service.MeService) *MeHandler {
  return &MeHandler{s}
}

func (h *MeHandler) Get(c *gin.Context) {
  var me, err = h.s.Get(c)
  if nil != err {
    var p problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternalProblem().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, me)
}

func (h *MeHandler) validateURL(w http.ResponseWriter, raw string) (u string, ok bool) {
  var uri, err = url.ParseRequestURI(raw)

  if nil != err {
    b := problem.Builder{}
    b.Title("Unprocessable photo URL format.")
    b.Status(http.StatusUnprocessableEntity)
    b.Detail("There was an error parsing the requested URL. Please try with a different URL or verify the current one for correctness.")
    b.With("photo_url", raw)
    b.Problem().Emit(w)
    return "", false
  }

  return uri.String(), true
}

func (h *MeHandler) SetPhoto(c *gin.Context) {
  var photoURL = strings.TrimSpace(c.PostForm("photo_url"))
  var ok bool

  if "" != photoURL {
    if photoURL, ok = h.validateURL(c.Writer, photoURL); !ok {
      return
    }
  }

  ok, err := h.s.Update(c, &transfer.MeUpdate{PhotoURL: photoURL})
  if nil != err {
    var p problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternalProblem().Emit(c.Writer)
    }
    return
  }

  if ok {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.info")
  }
}
