package handler

import (
  "errors"
  "fontseca/problem"
  "fontseca/service"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
  "time"
)

type ExperienceHandler struct {
  s service.ExperienceService
}

func NewExperienceHandler(s service.ExperienceService) *ExperienceHandler {
  return &ExperienceHandler{s}
}

func (h *ExperienceHandler) Get(c *gin.Context) {
  var e, err = h.s.Get(c)
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

func (h *ExperienceHandler) GetHidden(c *gin.Context) {
  var e, err = h.s.Get(c, true)
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

func (h *ExperienceHandler) GetByID(c *gin.Context) {
  var id = c.Query("id")
  var e, err = h.s.GetByID(c, id)
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

func (h *ExperienceHandler) Add(c *gin.Context) {
  var e transfer.ExperienceCreation

  if err := bindPostForm(c, &e); check(err, c.Writer) {
    return
  }

  if err := validateStruct(&e); check(err, c.Writer) {
    return
  }

  ok, err := h.s.Save(c, &e)
  if check(err, c.Writer) {
    return
  }

  if !ok {
    problem.NewInternal().Emit(c.Writer)
  }

  c.Status(http.StatusCreated)
}

func (h *ExperienceHandler) Set(c *gin.Context) {
  var update transfer.ExperienceUpdate

  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }

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
    c.Redirect(http.StatusSeeOther, "/experience.info?id="+id)
  }
}

func (h *ExperienceHandler) Hide(c *gin.Context) {
  id, success := c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }

  updated, err := h.s.Update(c, id, &transfer.ExperienceUpdate{Hidden: true})
  if check(err, c.Writer) {
    return
  }

  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/experience.info?id="+id)
  }
}

func (h *ExperienceHandler) Show(c *gin.Context) {
  id, success := c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }

  updated, err := h.s.Update(c, id, &transfer.ExperienceUpdate{Hidden: false})
  if check(err, c.Writer) {
    return
  }

  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/experience.info?id="+id)
  }
}

func (h *ExperienceHandler) Quit(c *gin.Context) {
  id, success := c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }

  var updated, err = h.s.Update(c, id, &transfer.ExperienceUpdate{
    Active: false,
    Ends:   time.Now().Year(),
  })

  if check(err, c.Writer) {
    return
  }

  if updated {
    c.Status(http.StatusNoContent)
  } else {
    c.Redirect(http.StatusSeeOther, "/experience.info?id="+id)
  }
}

func (h *ExperienceHandler) Remove(c *gin.Context) {
  var id, success = c.GetPostForm("id")
  if !success {
    problem.NewMissingParameter("id").Emit(c.Writer)
    return
  }

  if check(h.s.Remove(c, id), c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
