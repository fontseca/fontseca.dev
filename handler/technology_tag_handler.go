package handler

import (
  "fontseca.dev/problem"
  "fontseca.dev/service"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type TechnologyTagHandler struct {
  s service.TechnologyTagService
}

func NewTechnologyTagHandler(service service.TechnologyTagService) *TechnologyTagHandler {
  return &TechnologyTagHandler{s: service}
}

func (h *TechnologyTagHandler) Get(c *gin.Context) {
  var tags, err = h.s.Get(c)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, tags)
}

func (h *TechnologyTagHandler) Add(c *gin.Context) {
  var creation transfer.TechnologyTagCreation
  if err := bindPostForm(c, &creation); check(err, c.Writer) {
    return
  }
  if err := validateStruct(&creation); check(err, c.Writer) {
    return
  }
  insertedID, err := h.s.Add(c, &creation)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusOK, gin.H{"inserted_id": insertedID})
}

func (h *TechnologyTagHandler) Set(c *gin.Context) {
  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }
  var update transfer.TechnologyTagUpdate
  if err := bindPostForm(c, &update); check(err, c.Writer) {
    return
  }
  if err := validateStruct(&update); check(err, c.Writer) {
    return
  }
  updated, err := h.s.Update(c, id, &update)
  if check(err, c.Writer) {
    return
  }
  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Status(http.StatusConflict)
  }
}

func (h *TechnologyTagHandler) Remove(c *gin.Context) {
  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }
  err := h.s.Remove(c, id)
  if check(err, c.Writer) {
    return
  }
  c.Status(http.StatusNoContent)
}
