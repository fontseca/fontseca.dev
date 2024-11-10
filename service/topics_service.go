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
  Create(ctx context.Context, creation *transfer.TopicCreation) error
  List(ctx context.Context) (topics []*model.Topic, err error)
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

// Create adds a new topic.
func (s *TopicsService) Create(ctx context.Context, creation *transfer.TopicCreation) error {
  if nil == creation {
    err := errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return err
  }

  creation.Name = strings.TrimSpace(creation.Name)
  sanitizeTextWordIntersections(&creation.Name)
  creation.ID = toKebabCase(creation.Name)

  err := s.r.Create(ctx, creation)

  if nil != err {
    return err
  }

  s.setCache(ctx)

  return nil
}

// List retrieves all the topics.
func (s *TopicsService) List(ctx context.Context) (topics []*model.Topic, err error) {
  if s.hasCache() {
    return s.cache, nil
  }

  topics, err = s.r.List(ctx)

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

  id = strings.TrimSpace(id)
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
  s.cache, _ = s.List(ctx)
}

func (s *TopicsService) hasCache() bool {
  return 0 < len(s.cache)
}
