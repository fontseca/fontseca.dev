package handler

import (
  "fontseca/service"
  "github.com/gin-gonic/gin"
  "net/http"
)

type TechnologyTagHandler struct {
  s service.TechnologyTagService
}

func NewTechnologyTagHandler(service service.TechnologyTagService) *TechnologyTagHandler {
  return &TechnologyTagHandler{s: service}
}

func (h *TechnologyTagHandler) Get(c *gin.Context) {
  var tags, err = h.s.Get(c)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, tags)
}
