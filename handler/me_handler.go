package handler

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
  "strconv"
  "strings"
)

type meServiceAPI interface {
  Get(context.Context) (*model.Me, error)
  Update(context.Context, *transfer.MeUpdate) error
  SetHireable(context.Context, bool) error
}

type MeHandler struct {
  s meServiceAPI
}

func NewMeHandler(s meServiceAPI) *MeHandler {
  return &MeHandler{s}
}

func (h *MeHandler) Get(c *gin.Context) {
  var me, err = h.s.Get(c)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, me)
}

func (h *MeHandler) SetPhoto(c *gin.Context) {
  var photoURL = strings.TrimSpace(c.PostForm("photo_url"))
  if "" == photoURL {
    problem.NewMissingParameter("photo_url").Emit(c.Writer)
    return
  }

  err := h.s.Update(c, &transfer.MeUpdate{PhotoURL: photoURL})
  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *MeHandler) SetResume(c *gin.Context) {
  var resumeURL = strings.TrimSpace(c.PostForm("resume_url"))
  if "" == resumeURL {
    problem.NewMissingParameter("resume_url").Emit(c.Writer)
    return
  }

  err := h.s.Update(c, &transfer.MeUpdate{ResumeURL: resumeURL})
  if check(err, c.Writer) {
  }

  c.Status(http.StatusNoContent)
}

func (h *MeHandler) SetHireable(c *gin.Context) {
  var hireableStr = strings.TrimSpace(c.PostForm("hireable"))
  if "" == hireableStr {
    problem.NewMissingParameter("hireable").Emit(c.Writer)
    return
  }

  hireable, err := strconv.ParseBool(hireableStr)
  if nil != err {
    var numError *strconv.NumError
    if errors.As(err, &numError) {
      var p problem.Problem
      p.Type(problem.TypeUnparseableValue)
      p.Title("Failure when parsing boolean value.")
      p.Status(http.StatusUnprocessableEntity)
      p.Detail("Failed to parse the provided value as a boolean. Please ensure the value is either 'true' or 'false'.")
      p.With("value", numError.Num)
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }

  err = h.s.SetHireable(c, hireable)
  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *MeHandler) Set(c *gin.Context) {
  var update transfer.MeUpdate
  if err := bindPostForm(c, &update); check(err, c.Writer) {
    return
  }

  err := h.s.Update(c, &update)
  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
