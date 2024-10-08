package handler

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type patchesServiceAPI interface {
  Get(ctx context.Context) (patches []*model.ArticlePatch, err error)
  Revise(ctx context.Context, patchID string, revision *transfer.ArticleRevision) error
  Share(ctx context.Context, patchID string) (link string, err error)
  Discard(ctx context.Context, patchID string) error
  Release(ctx context.Context, patchID string) error
}

type PatchesHandler struct {
  patches patchesServiceAPI
}

func NewPatchesHandler(patches patchesServiceAPI) *PatchesHandler {
  return &PatchesHandler{patches}
}

func (h *PatchesHandler) Get(c *gin.Context) {
  patches, err := h.patches.Get(c)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, patches)
}

func (h *PatchesHandler) Revise(c *gin.Context) {
  patch, ok := c.GetPostForm("patch_uuid")

  if !ok {
    problem.NewMissingParameter("patch_uuid").Emit(c.Writer)
    return
  }

  var revision transfer.ArticleRevision

  if err := bindPostForm(c, &revision); check(err, c.Writer) {
    return
  }

  if err := h.patches.Revise(c, patch, &revision); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *PatchesHandler) Share(c *gin.Context) {
  draft, ok := c.GetPostForm("patch_uuid")

  if !ok {
    problem.NewMissingParameter("patch_uuid").Emit(c.Writer)
    return
  }

  link, err := h.patches.Share(c, draft)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, gin.H{"shareable_link": link})
}

func (h *PatchesHandler) Discard(c *gin.Context) {
  draft, ok := c.GetPostForm("patch_uuid")

  if !ok {
    problem.NewMissingParameter("patch_uuid").Emit(c.Writer)
    return
  }

  if err := h.patches.Discard(c, draft); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *PatchesHandler) Release(c *gin.Context) {
  draft, ok := c.GetPostForm("patch_uuid")

  if !ok {
    problem.NewMissingParameter("patch_uuid").Emit(c.Writer)
    return
  }

  if err := h.patches.Release(c, draft); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
