package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
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
  if nil == creation {
    err := errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return err
  }

  creation.Name = strings.TrimSpace(creation.Name)
  sanitizeTextWordIntersections(&creation.Name)
  creation.ID = toKebabCase(creation.Name)

  err := s.r.Add(ctx, creation)

  if nil != err {
    return err
  }

  s.setCache(ctx)

  return nil
}

func (s *topicsService) Get(ctx context.Context) (topics []*model.Topic, err error) {
  if s.hasCache() {
    return s.cache, nil
  }

  topics, err = s.r.Get(ctx)

  if nil != err {
    return nil, err
  }

  s.cache = topics

  return topics, err
}

func (s *topicsService) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  // TODO implement me
  panic("implement me")
}

func (s *topicsService) Remove(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (s *topicsService) setCache(ctx context.Context) {
  s.cache = nil
  s.cache, _ = s.Get(ctx)
}

func (s *topicsService) hasCache() bool {
  return 0 < len(s.cache)
}
