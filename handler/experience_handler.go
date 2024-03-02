package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type ExperienceHandler struct {
  s service.ExperienceService
}

func NewExperienceHandler(s service.ExperienceService) *ExperienceHandler {
  return &ExperienceHandler{s}
}

func (h *ExperienceHandler) Get(c *gin.Context) {
  var e, err = h.s.Get(c)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, e)
}

func (h *ExperienceHandler) GetHidden(c *gin.Context) {
  var e, err = h.s.Get(c, true)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, e)
}

func (h *ExperienceHandler) GetByID(c *gin.Context) {
  var id = c.Query("id")
  var e, err = h.s.GetByID(c, id)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, e)
}

func (h *ExperienceHandler) Add(c *gin.Context) {
  var e transfer.ExperienceCreation

  if err := bindPostForm(c, &e); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&e); check(err, c.Writer) {
    return
  }

  ok, err := h.s.Save(c, &e)
  if check(err, c.Writer) {
    return
  }

  if !ok {
    problem.NewInternal().Emit(c.Writer)
  }

  c.Status(http.StatusCreated)
}
