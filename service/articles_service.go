package service

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "strings"
)

// ArticlesService is a high level provider for articles.
type ArticlesService interface {
  // Get retrieves all the published articles.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over articles, so it attempts to find and amass every
  // article whose title contains any of the keywords (if more than one)
  // in needle.
  Get(ctx context.Context, needle string) (articles []*model.Article, err error)

  // GetHidden retrieves all the published articles thar are hidden.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over articles, so it attempts to find and amass every
  // article whose title contains any of the keywords (if more than one)
  // in needle.
  GetHidden(ctx context.Context, needle string) (articles []*model.Article, err error)

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

  // AddTopic adds a topic to the article. If the topic already
  // exists, it returns an error informing about a conflicting state.
  AddTopic(ctx context.Context, articleUUID, topicUUID string) error

  // RemoveTopic removes a topic from article. If the article
  // has no topic identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTopic(ctx context.Context, articleUUID, topicUUID string) error
}

type articlesService struct {
  r repository.ArchiveRepository
}

func NewArticlesService(r repository.ArchiveRepository) ArticlesService {
  return &articlesService{r}
}

func (s *articlesService) doGet(ctx context.Context, needle string, hidden ...bool) (articles []*model.Article, err error) {
  needle = strings.TrimSpace(needle)

  if "" != needle {
    if strings.Contains(needle, "_") {
      needle = strings.ReplaceAll(needle, "_", " ")
    }

    words := wordsOnly.FindAllString(needle, -1)
    needle = strings.Join(words, " ")
  }

  if 0 < len(hidden) {
    return s.r.Get(ctx, needle, hidden[0], false)
  }

  return s.r.Get(ctx, needle, false, false)
}

func (s *articlesService) Get(ctx context.Context, needle string) (articles []*model.Article, err error) {
  return s.doGet(ctx, needle)
}

func (s *articlesService) GetHidden(ctx context.Context, needle string) (articles []*model.Article, err error) {
  return s.doGet(ctx, needle, true)
}

func (s *articlesService) GetByID(ctx context.Context, articleUUID string) (article *model.Article, err error) {
  if err = validateUUID(&articleUUID); nil != err {
    return nil, err
  }

  return s.r.GetByID(ctx, articleUUID, true)
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
  // TODO implement me
  panic("implement me")
}

func (s *articlesService) Pin(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (s *articlesService) Unpin(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (s *articlesService) AddTopic(ctx context.Context, articleUUID, topicUUID string) error {
  // TODO implement me
  panic("implement me")
}

func (s *articlesService) RemoveTopic(ctx context.Context, articleUUID, topicUUID string) error {
  // TODO implement me
  panic("implement me")
}
