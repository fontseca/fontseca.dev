package handler

import (
  "fontseca.dev/service"
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
