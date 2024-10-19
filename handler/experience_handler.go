package handler

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
  "time"
)

type experienceServiceAPI interface {
  List(ctx context.Context, hidden ...bool) ([]*model.Experience, error)
  Get(ctx context.Context, id string) (*model.Experience, error)
  Create(ctx context.Context, creation *transfer.ExperienceCreation) (string, error)
  Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) error
  Hide(ctx context.Context, id string) error
  Show(ctx context.Context, id string) error
  Remove(ctx context.Context, id string) error
}

type ExperienceHandler struct {
  s experienceServiceAPI
}

func NewExperienceHandler(s experienceServiceAPI) *ExperienceHandler {
  return &ExperienceHandler{s}
}

func (h *ExperienceHandler) List(c *gin.Context) {
  var e, err = h.s.List(c)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, e)
}

func (h *ExperienceHandler) ListHidden(c *gin.Context) {
  var e, err = h.s.List(c, true)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, e)
}

func (h *ExperienceHandler) Get(c *gin.Context) {
  var id = c.Query("experience_uuid")
  var e, err = h.s.Get(c, id)
  if nil != err {
    var p *problem.Problem
    if errors.As(err, &p) {
      p.Emit(c.Writer)
    } else {
      problem.NewInternal().Emit(c.Writer)
    }
    return
  }
  c.JSON(http.StatusOK, e)
}

func (h *ExperienceHandler) Create(c *gin.Context) {
  var e transfer.ExperienceCreation

  if err := bindPostForm(c, &e); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&e); check(err, c.Writer) {
    return
  }

  created, err := h.s.Create(c, &e)
  if check(err, c.Writer) {
    return
  }
  c.JSON(http.StatusCreated, gin.H{"inserted_id": created})
}

func (h *ExperienceHandler) Set(c *gin.Context) {
  var update transfer.ExperienceUpdate

  var id, success = c.GetPostForm("experience_uuid")
  if !success {
    problem.NewMissingParameter("experience_uuid").Emit(c.Writer)
    return
  }

  if err := bindPostForm(c, &update); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&update); check(err, c.Writer) {
    return
  }

  err := h.s.Update(c, id, &update)
  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ExperienceHandler) Hide(c *gin.Context) {
  id, success := c.GetPostForm("experience_uuid")
  if !success {
    problem.NewMissingParameter("experience_uuid").Emit(c.Writer)
    return
  }

  err := h.s.Hide(c, id)
  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ExperienceHandler) Show(c *gin.Context) {
  id, success := c.GetPostForm("experience_uuid")
  if !success {
    problem.NewMissingParameter("experience_uuid").Emit(c.Writer)
    return
  }

  err := h.s.Show(c, id)
  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ExperienceHandler) Quit(c *gin.Context) {
  id, success := c.GetPostForm("experience_uuid")
  if !success {
    problem.NewMissingParameter("experience_uuid").Emit(c.Writer)
    return
  }

  var err = h.s.Update(c, id, &transfer.ExperienceUpdate{
    Active: false,
    Ends:   time.Now().Format(time.DateOnly),
  })

  if check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ExperienceHandler) Remove(c *gin.Context) {
  var id, success = c.GetPostForm("experience_uuid")
  if !success {
    problem.NewMissingParameter("experience_uuid").Emit(c.Writer)
    return
  }

  if check(h.s.Remove(c, id), c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
