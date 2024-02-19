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

func (h *MeHandler) SetResume(c *gin.Context) {
  var resumeURL = strings.TrimSpace(c.PostForm("resume_url"))
  var ok bool

  if "" != resumeURL {
    if resumeURL, ok = h.validateURL(c.Writer, resumeURL); !ok {
      return
    }
  }

  ok, err := h.s.Update(c, &transfer.MeUpdate{ResumeURL: resumeURL})
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