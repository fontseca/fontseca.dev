package handler

import (
  "fontseca/components/pages"
  "fontseca/service"
  "github.com/gin-gonic/gin"
)

type WebHandler struct {
  meService service.MeService
}

func NewWebHandler(meService service.MeService) *WebHandler {
  return &WebHandler{meService: meService}
}

func (h *WebHandler) RenderMe(c *gin.Context) {
  me, err := h.meService.Get(c)
  if nil != err {
    return
  }
  pages.Me(me).Render(c, c.Writer)
}
