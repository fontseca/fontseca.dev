package handler

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type topicsServiceAPI interface {
  Add(context.Context, *transfer.TopicCreation) error
  Get(context.Context) ([]*model.Topic, error)
  Update(context.Context, string, *transfer.TopicUpdate) error
  Remove(context.Context, string) error
}

type TopicsHandler struct {
  topics topicsServiceAPI
}

func NewTopicsHandler(topics topicsServiceAPI) *TopicsHandler {
  return &TopicsHandler{topics: topics}
}

func (h *TopicsHandler) Add(c *gin.Context) {
  var creation transfer.TopicCreation

  if err := bindPostForm(c, &creation); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&creation); check(err, c.Writer) {
    return
  }

  if err := h.topics.Add(c, &creation); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusCreated)
}

func (h *TopicsHandler) Get(c *gin.Context) {
  topics, err := h.topics.Get(c)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, topics)
}

func (h *TopicsHandler) Update(c *gin.Context) {
  var update transfer.TopicUpdate

  topic, ok := c.GetPostForm("topic_id")

  if !ok {
    problem.NewMissingParameter("topic_id").Emit(c.Writer)
    return
  }

  if err := bindPostForm(c, &update); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&update); check(err, c.Writer) {
    return
  }

  if err := h.topics.Update(c, topic, &update); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *TopicsHandler) Remove(c *gin.Context) {
  topic, ok := c.GetPostForm("topic_id")

  if !ok {
    problem.NewMissingParameter("topic_id").Emit(c.Writer)
    return
  }

  if err := h.topics.Remove(c, topic); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
