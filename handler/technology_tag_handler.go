package handler

import (
  "fontseca/service"
)

type TechnologyTagHandler struct {
  s service.TechnologyTagService
}

func NewTechnologyTagHandler(service service.TechnologyTagService) *TechnologyTagHandler {
  return &TechnologyTagHandler{s: service}
}
