package handler

import (
  "fontseca/service"
  "github.com/gin-gonic/gin"
  "net/http"
)

type ProjectsHandler struct {
  s service.ProjectsService
}

func NewProjectsHandler(service service.ProjectsService) *ProjectsHandler {
  return &ProjectsHandler{
    s: service,
  }
}

func (h *ProjectsHandler) Get(c *gin.Context) {
  var projects, err = h.s.Get(c)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, projects)
}
