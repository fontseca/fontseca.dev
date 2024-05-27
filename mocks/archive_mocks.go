package mocks

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/mock"
)

type ArchiveRepository struct {
  mock.Mock
}

func NewArchiveRepository() *ArchiveRepository {
  return new(ArchiveRepository)
}

func (o *ArchiveRepository) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
  args := o.Called(ctx, creation)
  return args.String(0), args.Error(1)
}

func (o *ArchiveRepository) Publish(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArchiveRepository) Get(ctx context.Context, needle string, hidden, draftsOnly bool) (articles []*model.Article, err error) {
  args := o.Called(ctx, needle, hidden, draftsOnly)
  arg0 := args.Get(0)

  if nil != arg0 {
    articles = arg0.([]*model.Article)
  }

  return articles, args.Error(1)
}

func (o *ArchiveRepository) GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error) {
  args := o.Called(ctx, id, isDraft)
  arg0 := args.Get(0)

  if nil != arg0 {
    article = arg0.(*model.Article)
  }

  return article, args.Error(1)
}

func (o *ArchiveRepository) Amend(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArchiveRepository) Remove(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArchiveRepository) AddTopic(ctx context.Context, articleID, topicID string, isDraft ...bool) error {
  return o.Called(ctx, articleID, topicID, isDraft).Error(0)
}

func (o *ArchiveRepository) RemoveTopic(ctx context.Context, articleID, topicID string, isDraft ...bool) error {
  return o.Called(ctx, articleID, topicID, isDraft).Error(0)
}

func (o *ArchiveRepository) SetHidden(ctx context.Context, id string, hidden bool) error {
  return o.Called(ctx, id, hidden).Error(0)
}

func (o *ArchiveRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
  return o.Called(ctx, id, pinned).Error(0)
}

func (o *ArchiveRepository) Share(ctx context.Context, id string) (link string, err error) {
  args := o.Called(ctx, id)
  return args.String(0), args.Error(1)
}

func (o *ArchiveRepository) Discard(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArchiveRepository) Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error {
  return o.Called(ctx, id, revision).Error(0)
}

func (o *ArchiveRepository) Release(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArchiveRepository) GetPatches(ctx context.Context) (patches []*model.ArticlePatch, err error) {
  args := o.Called(ctx)
  arg0 := args.Get(0)

  if nil != arg0 {
    patches = arg0.([]*model.ArticlePatch)
  }

  return patches, args.Error(1)
}

type DraftsService struct {
  mock.Mock
}

func NewDraftsService() *DraftsService {
  return &DraftsService{}
}

func (o *DraftsService) Draft(ctx context.Context, creation *transfer.ArticleCreation) (insertedUUID uuid.UUID, err error) {
  args := o.Called(ctx, creation)
  return args.Get(0).(uuid.UUID), args.Error(1)
}

func (o *DraftsService) Publish(ctx context.Context, draftUUID string) error {
  return o.Called(ctx, draftUUID).Error(0)
}

func (o *DraftsService) Get(ctx context.Context, needle string) (drafts []*model.Article, err error) {
  args := o.Called(ctx, needle)
  arg0 := args.Get(0)

  if nil != arg0 {
    drafts = arg0.([]*model.Article)
  }

  return drafts, args.Error(1)
}

func (o *DraftsService) GetByID(ctx context.Context, draftUUID string) (draft *model.Article, err error) {
  args := o.Called(ctx, draftUUID)
  arg0 := args.Get(0)

  if nil != arg0 {
    draft = arg0.(*model.Article)
  }

  return draft, args.Error(1)
}

func (o *DraftsService) AddTopic(ctx context.Context, draftUUID, topicUUID string) error {
  return o.Called(ctx, draftUUID, topicUUID).Error(0)
}

func (o *DraftsService) RemoveTopic(ctx context.Context, draftUUID, topicUUID string) error {
  return o.Called(ctx, draftUUID, topicUUID).Error(0)
}

func (o *DraftsService) Share(ctx context.Context, draftUUID string) (link string, err error) {
  args := o.Called(ctx, draftUUID)
  return args.String(0), args.Error(1)
}

func (o *DraftsService) Discard(ctx context.Context, draftUUID string) error {
  return o.Called(ctx, draftUUID).Error(0)
}

func (o *DraftsService) Revise(ctx context.Context, draftUUID string, revision *transfer.ArticleUpdate) error {
  return o.Called(ctx, draftUUID, revision).Error(0)
}

type ArticlesService struct {
  mock.Mock
}

func NewArticlesService() *ArticlesService {
  return new(ArticlesService)
}

func (o *ArticlesService) Get(ctx context.Context, needle string) (articles []*model.Article, err error) {
  args := o.Called(ctx, needle)
  arg0 := args.Get(0)

  if nil != arg0 {
    articles = arg0.([]*model.Article)
  }

  return articles, args.Error(1)
}

func (o *ArticlesService) GetHidden(ctx context.Context, needle string) (articles []*model.Article, err error) {
  args := o.Called(ctx, needle)
  arg0 := args.Get(0)

  if nil != arg0 {
    articles = arg0.([]*model.Article)
  }

  return articles, args.Error(1)
}

func (o *ArticlesService) GetByID(ctx context.Context, id string) (article *model.Article, err error) {
  args := o.Called(ctx, id)
  arg0 := args.Get(0)

  if nil != arg0 {
    article = arg0.(*model.Article)
  }

  return article, args.Error(1)
}

func (o *ArticlesService) Hide(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArticlesService) Show(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArticlesService) Amend(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArticlesService) Remove(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArticlesService) Pin(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArticlesService) Unpin(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}

func (o *ArticlesService) AddTopic(ctx context.Context, articleUUID, topicUUID string) error {
  return o.Called(ctx, articleUUID, topicUUID).Error(0)
}

func (o *ArticlesService) RemoveTopic(ctx context.Context, articleUUID, topicUUID string) error {
  return o.Called(ctx, articleUUID, topicUUID).Error(0)
}
