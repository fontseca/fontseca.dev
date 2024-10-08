package handler

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
)

type articlesServiceAPI interface {
  List(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error)
  Publications(ctx context.Context) (publications []*transfer.Publication, err error)
  ListHidden(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error)
  Get(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error)
  GetByID(ctx context.Context, articleUUID string) (article *model.Article, err error)
  Hide(ctx context.Context, articleID string) error
  Show(ctx context.Context, articleID string) error
  Amend(ctx context.Context, articleID string) error
  SetSlug(ctx context.Context, articleID, slug string) error
  Remove(ctx context.Context, articleID string) error
  Pin(ctx context.Context, articleID string) error
  Unpin(ctx context.Context, articleID string) error
  AddTag(ctx context.Context, articleUUID, tagID string) error
  RemoveTag(ctx context.Context, articleUUID, tagID string) error
}

type ArticlesHandler struct {
  articles articlesServiceAPI
}

func NewArticlesHandler(articles articlesServiceAPI) *ArticlesHandler {
  return &ArticlesHandler{articles}
}

func (h *ArticlesHandler) List(c *gin.Context) {
  filter := getArticleFilter(c)
  articles, err := h.articles.List(c, filter)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, articles)
}

func (h *ArticlesHandler) ListHidden(c *gin.Context) {
  filter := getArticleFilter(c)
  articles, err := h.articles.ListHidden(c, filter)

  if check(err, c.Writer) {
    return
  }

  c.JSON(http.StatusOK, articles)
}

func (h *ArticlesHandler) Get(c *gin.Context) {
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
