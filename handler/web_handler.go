package handler

import (
  "fontseca.dev/components/pages"
  "fontseca.dev/service"
  "github.com/gin-gonic/gin"
  "net/http"
)

type WebHandler struct {
  meService         service.MeService
  experienceService service.ExperienceService
  projectsService   service.ProjectsService
}

func NewWebHandler(
  meService service.MeService,
  experienceService service.ExperienceService,
  projectsService service.ProjectsService,
) *WebHandler {
  return &WebHandler{
    meService:         meService,
    experienceService: experienceService,
    projectsService:   projectsService,
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

func (h *WebHandler) RenderProjects(c *gin.Context) {
  var projects, err = h.projectsService.Get(c, false)
  if nil != err {
    return
  }
  pages.Projects(projects).Render(c, c.Writer)
}

func (h *WebHandler) RenderProjectDetails(c *gin.Context) {
  slug := c.Param("project_slug")
  var project, err = h.projectsService.GetBySlug(c, slug)
  if nil != err {
    c.Status(http.StatusNotFound)
    pages.ProjectDetails(nil).Render(c, c.Writer)
    return
  }
  pages.ProjectDetails(project).Render(c, c.Writer)
}
