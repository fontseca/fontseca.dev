package handler

import (
  "fontseca/service"
)

type ProjectsHandler struct {
  s service.ProjectsService
}

func NewProjectsHandler(service service.ProjectsService) *ProjectsHandler {
  return &ProjectsHandler{
    s: service,
  }
}
