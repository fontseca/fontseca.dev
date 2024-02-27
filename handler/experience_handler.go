package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
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
