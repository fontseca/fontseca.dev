package service

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
)

// TopicsService is a high level provider for articles.
type TopicsService interface {
  // Add adds a new topic.
  Add(ctx context.Context, creation *transfer.TopicCreation) (err error)

  // Get retrieves all the topics.
  Get(ctx context.Context) (topics []*model.Topic, err error)

  // Update updates an existing topic.
  Update(ctx context.Context, id string, update *transfer.TopicUpdate) error

  // Remove removes a topic and detaches it from any article that uses it.
  Remove(ctx context.Context, id string) error
}

type topicsService struct {
  cache []*model.Topic
  r     repository.TopicsRepository
}

func NewTopicsService(r repository.TopicsRepository) TopicsService {
  return &topicsService{
    cache: nil,
    r:     r,
  }
}

func (s *topicsService) Add(ctx context.Context, creation *transfer.TopicCreation) error {
  // TODO implement me
  panic("implement me")
}

func (s *topicsService) Get(ctx context.Context) (topics []*model.Topic, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *topicsService) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  // TODO implement me
  panic("implement me")
}

func (s *topicsService) Remove(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}
