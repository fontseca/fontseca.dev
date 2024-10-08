package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
)

type topicsRepositoryAPI interface {
  Add(ctx context.Context, creation *transfer.TopicCreation) error
  Get(ctx context.Context) (topics []*model.Topic, err error)
  Update(ctx context.Context, id string, update *transfer.TopicUpdate) error
  Remove(ctx context.Context, id string) error
}

// TopicsService is a high level provider for topics.
type TopicsService struct {
  cache []*model.Topic
  r     topicsRepositoryAPI
}

func NewTopicsService(r topicsRepositoryAPI) *TopicsService {
  return &TopicsService{
    cache: nil,
    r:     r,
  }
}

// Add adds a new topic.
func (s *TopicsService) Add(ctx context.Context, creation *transfer.TopicCreation) error {
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

// Get retrieves all the topics.
func (s *TopicsService) Get(ctx context.Context) (topics []*model.Topic, err error) {
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

// Update updates an existing topic.
func (s *TopicsService) Update(ctx context.Context, id string, update *transfer.TopicUpdate) error {
  if nil == update {
    err := errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return err
  }

  update.Name = strings.TrimSpace(update.Name)
  sanitizeTextWordIntersections(&update.Name)
  update.ID = toKebabCase(update.Name)

  err := s.r.Update(ctx, id, update)

  if nil != err {
    return err
  }

  s.setCache(ctx)

  return nil
}

// Remove removes a topic and detaches it from any
// article that currently uses it.
func (s *TopicsService) Remove(ctx context.Context, id string) error {
  err := s.r.Remove(ctx, id)

  if nil != err {
    return err
  }

  s.setCache(ctx)

  return nil
}

func (s *TopicsService) setCache(ctx context.Context) {
  s.cache = nil
  s.cache, _ = s.Get(ctx)
}

func (s *TopicsService) hasCache() bool {
  return 0 < len(s.cache)
}
