package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "log/slog"
)

// ArticlesService is a high level provider for articles.
type ArticlesService interface {
  // Get retrieves all the published articles.
  //
  // If filter.Search is a non-empty string, then Get behaves like a search
  // function over articles, so it attempts to find and amass every
  // article whose title contains any of the keywords (if more than one)
  // in filter.Search.
  Get(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error)

  // Publications retrieves a list of distinct months during which articles have been published.
  Publications(ctx context.Context) (publications []*transfer.Publication, err error)

  // GetHidden retrieves all the published articles thar are hidden.
  //
  // If filter.Search is a non-empty string, then Get behaves like a search
  // function over articles, so it attempts to find and amass every
  // article whose title contains any of the keywords (if more than one)
  // in filter.Search.
  GetHidden(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error)

  // GetOne retrieves one published article by the URL '/archive/:topic/:year/:month/:slug'.
  GetOne(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error)

  // GetByID retrieves one article by its UUID.
  GetByID(ctx context.Context, articleUUID string) (article *model.Article, err error)

  // Hide hides an article.
  Hide(ctx context.Context, id string) error

  // Show shows a hidden article.
  Show(ctx context.Context, id string) error

  // Amend starts the process to update an article. To amend the article,
  // a public copy of it is kept available to everyone while a patch
  // is created to store any revision made to the article.
  //
  // If the article is already being amended, any call to this method has
  // no effect.
  Amend(ctx context.Context, id string) error

  // Remove completely removes an article and any patch it currently has.
  Remove(ctx context.Context, id string) error

  // Pin pins an article.
  Pin(ctx context.Context, id string) error

  // Unpin unpins a pinned article.
  Unpin(ctx context.Context, id string) error

  // AddTag adds a tag to the article. If the tag already
  // exists, it returns an error informing about a conflicting state.
  AddTag(ctx context.Context, articleUUID, tagID string) error

  // RemoveTag removes a tag from article. If the article
  // has no tag identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTag(ctx context.Context, articleUUID, tagID string) error
}

type articlesService struct {
  r repository.ArchiveRepository
}

func NewArticlesService(r repository.ArchiveRepository) ArticlesService {
  return &articlesService{r}
}

func (s *articlesService) doGet(ctx context.Context, filter *transfer.ArticleFilter, hidden ...bool) (articles []*transfer.Article, err error) {
  if 0 < len(hidden) {
    return s.r.Get(ctx, filter, hidden[0], false)
  }

  return s.r.Get(ctx, filter, false, false)
}

func (s *articlesService) Get(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error) {
  return s.doGet(ctx, filter)
}

func (s *articlesService) Publications(ctx context.Context) (publications []*transfer.Publication, err error) {
  return s.r.Publications(ctx)
}

func (s *articlesService) GetHidden(ctx context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error) {
  return s.doGet(ctx, filter, true)
}

func (s *articlesService) GetOne(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error) {
  if nil == request {
    err = errors.New("nil value for parameter: request")
    slog.Error(err.Error())
    return nil, err
  }

  return s.r.GetOne(ctx, request)
}

func (s *articlesService) GetByID(ctx context.Context, articleUUID string) (article *model.Article, err error) {
  if err = validateUUID(&articleUUID); nil != err {
    return nil, err
  }

  return s.r.GetByID(ctx, articleUUID, false)
}

func (s *articlesService) Hide(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetHidden(ctx, id, true)
}

func (s *articlesService) Show(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetHidden(ctx, id, false)
}

func (s *articlesService) Amend(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Amend(ctx, id)
}

func (s *articlesService) Remove(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.Remove(ctx, id)
}

func (s *articlesService) Pin(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetPinned(ctx, id, true)
}

func (s *articlesService) Unpin(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }

  return s.r.SetPinned(ctx, id, false)
}

func (s *articlesService) AddTag(ctx context.Context, articleUUID, tagID string) error {
  if err := validateUUID(&articleUUID); nil != err {
    return err
  }

  return s.r.AddTag(ctx, articleUUID, tagID)
}

func (s *articlesService) RemoveTag(ctx context.Context, articleUUID, tagID string) error {
  if err := validateUUID(&articleUUID); nil != err {
    return err
  }

  return s.r.RemoveTag(ctx, articleUUID, tagID)
}
