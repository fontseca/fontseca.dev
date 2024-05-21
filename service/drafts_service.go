package service

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
)

// DraftsService is a high level provider for article drafts.
type DraftsService interface {
  // Draft starts the creation process of an article. It returns the
  // UUID of the draft that was created.
  //
  // To draft an article, only its title is required, other fields
  // are completely optional and can be added in an eventual revision.
  Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error)

  // Publish makes a draft publicly available.
  //
  // Invoking Publish on an already published article has no effect.
  Publish(ctx context.Context, id string) error

  // Get retrieves all the ongoing articles drafts.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over draft articles, so it attempts to find and
  // amass every article whose title contains any of the keywords
  // (if more than one) in needle.
  Get(ctx context.Context, needle string) (articles []*model.Article, err error)

  // GetByID retrieves one article draft by its UUID.
  GetByID(ctx context.Context, id string) (article *model.Article, err error)

  // AddTopic adds a topic to the article draft. If the topic already
  // exists, it returns an error informing about a conflicting state.
  AddTopic(ctx context.Context, articleID, topicID string) error

  // RemoveTopic removes a topic from the article draft. If the article
  // has  no topic identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTopic(ctx context.Context, articleID, topicID string) error

  // Share creates a shareable link for an article draft. Only users
  // with that link can see the progress and provide feedback.
  //
  // A shareable link does not make an article public. This link will
  // eventually expire after a certain amount of time.
  Share(ctx context.Context, id string) (link string, err error)

  // Discard completely drops an article draft.
  Discard(ctx context.Context, id string) error

  // Revise adds a correction or inclusion to an article draft in order
  // to correct or improve it.
  Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error
}

type draftsService struct {
  r repository.ArchiveRepository
}

func NewDraftsService(r repository.ArchiveRepository) DraftsService {
  return &draftsService{r}
}

func (s *draftsService) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) Publish(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) Get(ctx context.Context, needle string) (articles []*model.Article, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) GetByID(ctx context.Context, id string) (article *model.Article, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) AddTopic(ctx context.Context, articleID, topicID string) error {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) RemoveTopic(ctx context.Context, articleID, topicID string) error {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) Share(ctx context.Context, id string) (link string, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) Discard(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (s *draftsService) Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error {
  // TODO implement me
  panic("implement me")
}
