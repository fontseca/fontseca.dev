package handler

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type projectsServiceAPI interface {
  List(ctx context.Context, archived ...bool) ([]*model.Project, error)
  Get(ctx context.Context, projectID string) (*model.Project, error)
  GetBySlug(ctx context.Context, projectID string) (*model.Project, error)
  Create(ctx context.Context, creation *transfer.ProjectCreation) (string, error)
  Exists(ctx context.Context, projectID string) error
  Update(ctx context.Context, projectID string, update *transfer.ProjectUpdate) (bool, error)
  Unarchive(ctx context.Context, projectID string) (bool, error)
  Remove(ctx context.Context, projectID string) error
  HasTag(ctx context.Context, projectID, tagID string) (bool, error)
  AddTag(ctx context.Context, projectID, tagID string) (bool, error)
  RemoveTag(ctx context.Context, projectID, tagID string) (bool, error)
}

type ProjectsHandler struct {
  s projectsServiceAPI
}

func NewProjectsHandler(service projectsServiceAPI) *ProjectsHandler {
  return &ProjectsHandler{
    s: service,
  }
}

func (h *ProjectsHandler) List(c *gin.Context) {
  var projects, err = h.s.List(c)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, projects)
}

func (h *ProjectsHandler) ListArchived(c *gin.Context) {
  var projects, err = h.s.List(c, true)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, projects)
}

func (h *ProjectsHandler) Get(c *gin.Context) {
  var id = c.Query("project_uuid")
  var project, err = h.s.Get(c, id)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, project)
}

func (h *ProjectsHandler) Create(c *gin.Context) {
  var creation = transfer.ProjectCreation{}
  if err := bindPostForm(c, &creation); check(err, c.Writer) {
    return
  }
  if err := validateStruct(&creation); check(err, c.Writer) {
    return
  }
  var insertedID, err = h.s.Create(c, &creation)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, gin.H{"inserted_id": insertedID})
}

func (h *ProjectsHandler) Set(c *gin.Context) {
  var id, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
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
    c.Redirect(http.StatusSeeOther, "/me.projects.get?id="+id)
  }
}

func (h *ProjectsHandler) Archive(c *gin.Context) {
  var id, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  _, err := h.s.Update(c, id, &transfer.ProjectUpdate{Archived: true})
  if check(err, c.Writer) {
    return
  }
  c.Status(http.StatusNoContent)
}

func (h *ProjectsHandler) Unarchive(c *gin.Context) {
  var id, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  var unarchived, err = h.s.Unarchive(c, id)
  if check(err, c.Writer) {
    return
  }
  if unarchived {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.projects.get?id="+id)
  }
}

func (h *ProjectsHandler) Finish(c *gin.Context) {
  var id, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  var updated, err = h.s.Update(c, id, &transfer.ProjectUpdate{Finished: true})
  if check(err, c.Writer) {
    return
  }
  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.projects.get?id="+id)
  }
}

func (h *ProjectsHandler) Unfinish(c *gin.Context) {
  var id, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  var updated, err = h.s.Update(c, id, &transfer.ProjectUpdate{Finished: false})
  if check(err, c.Writer) {
    return
  }
  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/me.projects.get?id="+id)
  }
}

func (h *ProjectsHandler) getIDAndURLParameters(c *gin.Context) (id string, url string, ok bool) {
  id, success := c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return "", "", false
  }
  url, success = c.GetPostForm("url")
  if !success {
    problem.NewMissingParameter("url").Emit(c.Writer)
    return "", "", false
  }
  return id, url, true
}

func (h *ProjectsHandler) setURL(c *gin.Context, id string, update *transfer.ProjectUpdate) {
  var updated, err = h.s.Update(c, id, update)
  if check(err, c.Writer) {
    return
  }
  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Status(http.StatusConflict)
  }
}

func (h *ProjectsHandler) SetPlaygroundURL(c *gin.Context) {
  id, url, ok := h.getIDAndURLParameters(c)
  if !ok {
    return
  }
  h.setURL(c, id, &transfer.ProjectUpdate{PlaygroundURL: url})
}

func (h *ProjectsHandler) SetFirstImageURL(c *gin.Context) {
  id, url, ok := h.getIDAndURLParameters(c)
  if !ok {
    return
  }
  h.setURL(c, id, &transfer.ProjectUpdate{FirstImageURL: url})
}

func (h *ProjectsHandler) SetSecondImageURL(c *gin.Context) {
  id, url, ok := h.getIDAndURLParameters(c)
  if !ok {
    return
  }
  h.setURL(c, id, &transfer.ProjectUpdate{SecondImageURL: url})
}

func (h *ProjectsHandler) SetGitHubURL(c *gin.Context) {
  id, url, ok := h.getIDAndURLParameters(c)
  if !ok {
    return
  }
  h.setURL(c, id, &transfer.ProjectUpdate{GitHubURL: url})
}

func (h *ProjectsHandler) SetCollectionURL(c *gin.Context) {
  id, url, ok := h.getIDAndURLParameters(c)
  if !ok {
    return
  }
  h.setURL(c, id, &transfer.ProjectUpdate{CollectionURL: url})
}

func (h *ProjectsHandler) Remove(c *gin.Context) {
  var id, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  err := h.s.Remove(c, id)
  if check(err, c.Writer) {
    return
  }
  c.Status(http.StatusNoContent)
}

func (h *ProjectsHandler) AddTag(c *gin.Context) {
  var projectID, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  technologyTagID, success := c.GetPostForm("technology_id")
  if !success {
    problem.NewMissingParameter("technology_id").Emit(c.Writer)
    return
  }
  var added, err = h.s.AddTag(c, projectID, technologyTagID)
  if check(err, c.Writer) {
    return
  }
  if added {
    c.Status(http.StatusNoContent)
  } else {
    c.Status(http.StatusConflict)
  }
}

func (h *ProjectsHandler) RemoveTag(c *gin.Context) {
  var projectID, success = c.GetPostForm("project_uuid")
  if !success {
    problem.NewMissingParameter("project_uuid").Emit(c.Writer)
    return
  }
  technologyTagID, success := c.GetPostForm("technology_id")
  if !success {
    problem.NewMissingParameter("technology_id").Emit(c.Writer)
    return
  }
  var removed, err = h.s.RemoveTag(c, projectID, technologyTagID)
  if check(err, c.Writer) {
    return
  }
  if removed {
    c.Status(http.StatusNoContent)
  } else {
    c.Status(http.StatusConflict)
  }
}
