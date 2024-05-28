package mocks

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/mock"
)

type TopicsRepository struct {
  mock.Mock
}

func NewTopicsRepository() *TopicsRepository {
  return new(TopicsRepository)
}

func (o *TopicsRepository) Add(ctx context.Context, creation *transfer.TopicCreation) error {
  return o.Called(ctx, creation).Error(0)
}

func (o *TopicsRepository) Get(ctx context.Context) (topics []*model.Topic, err error) {
  args := o.Called(ctx)
  arg0 := args.Get(0)

  if arg0 != nil {
    topics = arg0.([]*model.Topic)
  }

  return topics, args.Error(1)
}

func (o *TopicsRepository) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  return o.Called(ctx, id, update).Error(0)
}

func (o *TopicsRepository) Remove(ctx context.Context, id string) error {
  return o.Called(ctx, id).Error(0)
}
