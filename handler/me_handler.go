package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
  "strconv"
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

func (h *MeHandler) SetHireable(c *gin.Context) {
  hireable, err := strconv.ParseBool(strings.TrimSpace(c.DefaultPostForm("hireable", "false")))
  if nil != err {
    var numError *strconv.NumError
    if errors.As(err, &numError) {
      var b problem.Builder
      b.Title("Failure when parsing boolean value.")
      b.Status(http.StatusUnprocessableEntity)
      b.Detail("Failed to parse the provided value as a boolean. Please ensure the value is either 'true' or 'false'.")
      b.With("value", numError.Num)
      b.Problem().Emit(c.Writer)
    } else {
      problem.NewInternalProblem().Emit(c.Writer)
    }
    return
  }

  ok, err := h.s.Update(c, &transfer.MeUpdate{Hireable: hireable})
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

func (h *MeHandler) Update(c *gin.Context) {
  var update transfer.MeUpdate

  if ok := bindJSONRequestBody(c, &update); !ok {
    return
  }

  ok, err := h.s.Update(c, &update)
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
