package mocks

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/mock"
)

type TagsRepository struct {
  mock.Mock
}

func NewTagsRepository() *TagsRepository {
  return new(TagsRepository)
}

func (o *TagsRepository) Add(ctx context.Context, creation *transfer.TagCreation) error {
  return o.Called(ctx, creation).Error(0)
}

func (o *TagsRepository) Get(ctx context.Context) (tags []*model.Tag, err error) {
  args := o.Called(ctx)
  arg0 := args.Get(0)

  if arg0 != nil {
    tags = arg0.([]*model.Tag)
  }

  return tags, args.Error(1)
}

func (o *TagsRepository) Update(ctx context.Context, id string, update *transfer.TagUpdate) error {
  return o.Called(ctx, id, update).Error(0)
}

func (o *TagsRepository) Remove(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}
