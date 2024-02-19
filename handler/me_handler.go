package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
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
  var photoURL = c.PostForm("photo_url")
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
  var resumeURL = c.PostForm("resume_url")
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
