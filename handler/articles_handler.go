package handler

import (
  "fontseca.dev/problem"
  "fontseca.dev/service"
  "github.com/gin-gonic/gin"
  "net/http"
)

type ArticlesHandler struct {
  articles service.ArticlesService
}

func NewArticlesHandler(articles service.ArticlesService) *ArticlesHandler {
  return &ArticlesHandler{articles}
}

func (h *ArticlesHandler) Get(c *gin.Context) {
  filter := getArticleFilter(c)
  articles, err := h.articles.Get(c, filter)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, articles)
}

func (h *ArticlesHandler) GetHidden(c *gin.Context) {
  filter := getArticleFilter(c)
  articles, err := h.articles.GetHidden(c, filter)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, articles)
}

func (h *ArticlesHandler) GetByID(c *gin.Context) {
  id := c.Query("article_uuid")
  article, err := h.articles.GetByID(c, id)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, article)
}

func (h *ArticlesHandler) Hide(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  if err := h.articles.Hide(c, article); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) Show(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  if err := h.articles.Show(c, article); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) Amend(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  if err := h.articles.Amend(c, article); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) SetSlug(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  slug, ok := c.GetPostForm("slug")

  if !ok {
    problem.NewMissingParameter("slug").Emit(c.Writer)
    return
  }

  if err := h.articles.SetSlug(c, article, slug); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) Remove(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  if err := h.articles.Remove(c, article); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) Pin(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  if err := h.articles.Pin(c, article); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) Unpin(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  if err := h.articles.Unpin(c, article); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) AddTag(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  tag, ok := c.GetPostForm("tag_id")

  if !ok {
    problem.NewMissingParameter("tag_id").Emit(c.Writer)
    return
  }

  if err := h.articles.AddTag(c, article, tag); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}

func (h *ArticlesHandler) RemoveTag(c *gin.Context) {
  article, ok := c.GetPostForm("article_uuid")

  if !ok {
    problem.NewMissingParameter("article_uuid").Emit(c.Writer)
    return
  }

  tag, ok := c.GetPostForm("tag_id")

  if !ok {
    problem.NewMissingParameter("tag_id").Emit(c.Writer)
    return
  }

  if err := h.articles.RemoveTag(c, article, tag); check(err, c.Writer) {
    return
  }

  c.Status(http.StatusNoContent)
}
