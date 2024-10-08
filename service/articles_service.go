package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
)

type archiveRepositoryAPIForArticles interface {
  SetSlug(ctx context.Context, articleID, slug string) error
  Publications(ctx context.Context) (publications []*transfer.Publication, err error)
  List(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error)
  Get(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error)
  GetByID(ctx context.Context, articleID string, isDraft bool) (article *model.Article, err error)
  Amend(ctx context.Context, articleID string) error
  Remove(ctx context.Context, articleID string) error
  AddTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error
  RemoveTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error
  SetHidden(ctx context.Context, articleID string, hidden bool) error
  SetPinned(ctx context.Context, articleID string, pinned bool) error
}

// ArticlesService is a high level provider for articles.
type ArticlesService struct {
  r archiveRepositoryAPIForArticles
}

func NewArticlesService(r archiveRepositoryAPIForArticles) *ArticlesService {
  return &ArticlesService{r}
}

func (s *ArticlesService) list(ctx context.Context, filter *transfer.ArticleFilter, hidden ...bool) (articles []*transfer.Article, err error) {
  if 0 < len(hidden) {
    return s.r.List(ctx, filter, hidden[0], false)
  }

  return s.r.List(ctx, filter, false, false)
}

// List retrieves all the published articles.
//
// If filter.Search is a non-empty string, then Get behaves like a search
// function over articles, so it attempts to find and amass every
// article whose title contains any of the keywords (if more than one)
// in filter.Search.
func (s *ArticlesService) List(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error) {
  return s.list(ctx, filter)
}

// Publications retrieves a list of distinct months during which articles have been published.
func (s *ArticlesService) Publications(ctx context.Context) (publications []*transfer.Publication, err error) {
  return s.r.Publications(ctx)
}

// ListHidden retrieves all the published articles thar are hidden.
//
// If filter.Search is a non-empty string, then Get behaves like a search
// function over articles, so it attempts to find and amass every
// article whose title contains any of the keywords (if more than one)
// in filter.Search.
func (s *ArticlesService) ListHidden(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error) {
  return s.list(ctx, filter, true)
}

// Get retrieves one published article by the URL '/archive/:topic/:year/:month/:slug'.
func (s *ArticlesService) Get(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error) {
  if nil == request {
    err = errors.New("nil value for parameter: request")
    slog.Error(err.Error())
    return nil, err
  }

  return s.r.Get(ctx, request)
}

// GetByID retrieves one article by its UUID.
func (s *ArticlesService) GetByID(ctx context.Context, articleUUID string) (article *model.Article, err error) {
  if err = validateUUID(&articleUUID); nil != err {
    return nil, err
  }

  return s.r.GetByID(ctx, articleUUID, false)
}

// Hide hides an article.
func (s *ArticlesService) Hide(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetHidden(ctx, id, true)
}

// Show shows a hidden article.
func (s *ArticlesService) Show(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetHidden(ctx, id, false)
}

// SetSlug changes the slug of an article.
func (s *ArticlesService) SetSlug(ctx context.Context, id, slug string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  slug = strings.TrimSpace(slug)

  if "" == slug {
    return nil
  }

  return s.r.SetSlug(ctx, id, generateSlug(slug))
}

// Amend starts the process to update an article. To amend the article,
// a public copy of it is kept available to everyone while a patch
// is created to store any revision made to the article.
//
// If the article is already being amended, any call to this method has
// no effect.
func (s *ArticlesService) Amend(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Amend(ctx, id)
}

// Remove completely removes an article and any patch it currently has.
func (s *ArticlesService) Remove(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Remove(ctx, id)
}

// Pin pins an article.
func (s *ArticlesService) Pin(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetPinned(ctx, id, true)
}

// Unpin unpins a pinned article.
func (s *ArticlesService) Unpin(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetPinned(ctx, id, false)
}

// AddTag adds a tag to the article. If the tag already
// exists, it returns an error informing about a conflicting state.
func (s *ArticlesService) AddTag(ctx context.Context, articleUUID, tagID string) error {
  if err := validateUUID(&articleUUID); nil != err {
    return err
  }

  return s.r.AddTag(ctx, articleUUID, tagID)
}

// RemoveTag removes a tag from article. If the article
// has no tag identified by its UUID, it returns an error indication
// a not found state.
func (s *ArticlesService) RemoveTag(ctx context.Context, articleUUID, tagID string) error {
  if err := validateUUID(&articleUUID); nil != err {
    return err
  }

  return s.r.RemoveTag(ctx, articleUUID, tagID)
}
