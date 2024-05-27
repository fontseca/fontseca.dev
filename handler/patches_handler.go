package handler

import (
  "fontseca.dev/problem"
  "fontseca.dev/service"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type PatchesHandler struct {
  patches service.PatchesService
}

func NewPatchesHandler(patches service.PatchesService) *PatchesHandler {
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

  var revision transfer.ArticleUpdate

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
