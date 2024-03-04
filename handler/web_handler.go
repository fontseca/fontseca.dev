package handler

import (
  "fontseca/components/pages"
  "fontseca/service"
  "github.com/gin-gonic/gin"
)

type WebHandler struct {
  meService         service.MeService
  experienceService service.ExperienceService
}

func NewWebHandler(
  meService service.MeService,
  experienceService service.ExperienceService,
) *WebHandler {
  return &WebHandler{
    meService:         meService,
    experienceService: experienceService,
  }
}

func (h *WebHandler) RenderMe(c *gin.Context) {
  me, err := h.meService.Get(c)
  if nil != err {
    return
  }
  pages.Me(me).Render(c, c.Writer)
}

func (h *WebHandler) RenderExperience(c *gin.Context) {
  var exp, err = h.experienceService.Get(c)
  if nil != err {
    return
  }
  pages.Experience(exp).Render(c, c.Writer)
}
