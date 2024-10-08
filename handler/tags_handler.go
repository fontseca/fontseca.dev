package handler

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type tagsServiceAPI interface {
  Create(context.Context, *transfer.TagCreation) error
  List(context.Context) ([]*model.Tag, error)
  Update(context.Context, string, *transfer.TagUpdate) error
  Remove(context.Context, string) error
}

type TagsHandler struct {
  tags tagsServiceAPI
}

func NewTagsHandler(tags tagsServiceAPI) *TagsHandler {
  return &TagsHandler{tags: tags}
}

func (h *TagsHandler) Create(c *gin.Context) {
  var creation transfer.TagCreation

  if err := bindPostForm(c, &creation); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&creation); check(err, c.Writer) {
    return
  }

  if err := h.tags.Create(c, &creation); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusCreated)
}

func (h *TagsHandler) List(c *gin.Context) {
  tags, err := h.tags.List(c)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, tags)
}

func (h *TagsHandler) Set(c *gin.Context) {
  var update transfer.TagUpdate

  tag, ok := c.GetPostForm("tag_id")

  if !ok {
    problem.NewMissingParameter("tag_id").Emit(c.Writer)
    return
  }

  if err := bindPostForm(c, &update); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&update); check(err, c.Writer) {
    return
  }

  if err := h.tags.Update(c, tag, &update); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *TagsHandler) Remove(c *gin.Context) {
  tag, ok := c.GetPostForm("tag_id")

  if !ok {
    problem.NewMissingParameter("tag_id").Emit(c.Writer)
    return
  }

  if err := h.tags.Remove(c, tag); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
