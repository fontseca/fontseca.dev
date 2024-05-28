package handler

import (
  "fontseca.dev/problem"
  "fontseca.dev/service"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type DraftsHandler struct {
  drafts service.DraftsService
}

func NewDraftsHandler(drafts service.DraftsService) *DraftsHandler {
  return &DraftsHandler{drafts}
}

func (h *DraftsHandler) Start(c *gin.Context) {
  var articleCreation transfer.ArticleCreation

  if err := bindPostForm(c, &articleCreation); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&articleCreation); check(err, c.Writer) {
    return
  }

  insertedUUID, err := h.drafts.Draft(c, &articleCreation)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusCreated, gin.H{"draft_uuid": insertedUUID})
}

func (h *DraftsHandler) Publish(c *gin.Context) {
  draft, ok := c.GetPostForm("draft_uuid")

  if !ok {
    problem.NewMissingParameter("draft_uuid").Emit(c.Writer)
    return
  }

  if err := h.drafts.Publish(c, draft); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *DraftsHandler) Get(c *gin.Context) {
  search := c.Query("search")
  drafts, err := h.drafts.Get(c, search)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, drafts)
}

func (h *DraftsHandler) GetByID(c *gin.Context) {
  id := c.Query("draft_uuid")
  draft, err := h.drafts.GetByID(c, id)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, draft)
}

func (h *DraftsHandler) AddTag(c *gin.Context) {
  draft, ok := c.GetPostForm("draft_uuid")

  if !ok {
    problem.NewMissingParameter("draft_uuid").Emit(c.Writer)
    return
  }

  tag, ok := c.GetPostForm("tag_id")

  if !ok {
    problem.NewMissingParameter("tag_id").Emit(c.Writer)
    return
  }

  if err := h.drafts.AddTag(c, draft, tag); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *DraftsHandler) RemoveTag(c *gin.Context) {
  draft, ok := c.GetPostForm("draft_uuid")

  if !ok {
    problem.NewMissingParameter("draft_uuid").Emit(c.Writer)
    return
  }

  tag, ok := c.GetPostForm("tag_id")

  if !ok {
    problem.NewMissingParameter("tag_id").Emit(c.Writer)
    return
  }

  if err := h.drafts.RemoveTag(c, draft, tag); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *DraftsHandler) Share(c *gin.Context) {
  draft, ok := c.GetPostForm("draft_uuid")

  if !ok {
    problem.NewMissingParameter("draft_uuid").Emit(c.Writer)
    return
  }

  link, err := h.drafts.Share(c, draft)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, gin.H{"shareable_link": link})
}

func (h *DraftsHandler) Discard(c *gin.Context) {
  draft, ok := c.GetPostForm("draft_uuid")

  if !ok {
    problem.NewMissingParameter("draft_uuid").Emit(c.Writer)
    return
  }

  if err := h.drafts.Discard(c, draft); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *DraftsHandler) Revise(c *gin.Context) {
  draft, ok := c.GetPostForm("draft_uuid")

  if !ok {
    problem.NewMissingParameter("draft_uuid").Emit(c.Writer)
    return
  }

  var revision transfer.ArticleUpdate

  if err := bindPostForm(c, &revision); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&revision); check(err, c.Writer) {
    return
  }

  if err := h.drafts.Revise(c, draft, &revision); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
