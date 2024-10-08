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
  Update(context.Context, *transfer.MeUpdate) (bool, error)
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
  var photoURL = c.PostForm("photo_url")
  ok, err := h.s.Update(c, &transfer.MeUpdate{PhotoURL: photoURL})
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
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
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
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
      var p problem.Problem
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

  ok, err := h.s.Update(c, &transfer.MeUpdate{Hireable: hireable})
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }

  if ok {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.info")
  }
}

func (h *MeHandler) Set(c *gin.Context) {
  var update transfer.MeUpdate

  if ok := bindJSONRequestBody(c, &update); !ok {
    return
  }

  ok, err := h.s.Update(c, &update)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }

  if ok {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.info")
  }
}
