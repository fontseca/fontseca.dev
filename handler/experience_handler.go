package handler

import (
  "fontseca/service"
)

type ExperienceHandler struct {
  s service.ExperienceService
}

func NewExperienceHandler(s service.ExperienceService) *ExperienceHandler {
  return &ExperienceHandler{s}
}
