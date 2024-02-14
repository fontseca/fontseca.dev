package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
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
