package handler

import (
  "fontseca.dev/service"
)

type DraftsHandler struct {
  drafts service.DraftsService
}

func NewDraftsHandler(drafts service.DraftsService) *DraftsHandler {
  return &DraftsHandler{drafts}
}
