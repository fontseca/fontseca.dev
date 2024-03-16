package handler

import (
  "fontseca/problem"
  "fontseca/service"
  "fontseca/transfer"
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

func (h *ProjectsHandler) GetArchived(c *gin.Context) {
  var projects, err = h.s.Get(c, true)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, projects)
}

func (h *ProjectsHandler) GetByID(c *gin.Context) {
  var id = c.Query("id")
  var project, err = h.s.GetByID(c, id)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, project)
}

func (h *ProjectsHandler) Add(c *gin.Context) {
  var creation = transfer.ProjectCreation{}
  if err := bindPostForm(c, &creation); check(err, c.Writer) {
    return
  }
  if err := validateStruct(&creation); check(err, c.Writer) {
    return
  }
  var insertedID, err = h.s.Add(c, &creation)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, gin.H{"inserted_id": insertedID})
}

func (h *ProjectsHandler) Set(c *gin.Context) {
  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }
  var update transfer.ProjectUpdate
  if err := bindPostForm(c, &update); check(err, c.Writer) {
    return
  }
  if err := validateStruct(&update); check(err, c.Writer) {
    return
  }
  var updated, err = h.s.Update(c, id, &update)
  if check(err, c.Writer) {
    return
  }
  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.projects.info?id="+id)
  }
}

func (h *ProjectsHandler) Archive(c *gin.Context) {
  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }
  _, err := h.s.Update(c, id, &transfer.ProjectUpdate{Archived: true})
  if check(err, c.Writer) {
    return
  }
  c.Status(http.StatusNoContent)
}

func (h *ProjectsHandler) Unarchive(c *gin.Context) {
  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }
  var unarchived, err = h.s.Unarchive(c, id)
  if check(err, c.Writer) {
    return
  }
  if unarchived {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.projects.info?id="+id)
  }
}
