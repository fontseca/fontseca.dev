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

// TagsService is a high level provider for tags.
type TagsService interface {
  // Add adds a new tag.
  Add(ctx context.Context, creation *transfer.TagCreation) error

  // Get retrieves all the tags.
  Get(ctx context.Context) (tags []*model.Tag, err error)

  // Update updates an existing tag.
  Update(ctx context.Context, id string, update *transfer.TagUpdate) error

  // Remove removes a tag and detaches it from any
  // article that currently uses it.
  Remove(ctx context.Context, id string) error
}

type tagsService struct {
  cache []*model.Tag
  r     repository.TagsRepository
}

func NewTagsService(r repository.TagsRepository) TagsService {
  return &tagsService{
    cache: nil,
    r:     r,
  }
}

func (s *tagsService) Add(ctx context.Context, creation *transfer.TagCreation) error {
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

func (s *tagsService) Get(ctx context.Context) (tags []*model.Tag, err error) {
  if s.hasCache() {
    return s.cache, nil
  }

  tags, err = s.r.Get(ctx)

  if nil != err {
    return nil, err
  }

  s.cache = tags

  return tags, err
}

func (s *tagsService) Update(ctx context.Context, id string, update *transfer.TagUpdate) error {
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

func (s *tagsService) Remove(ctx context.Context, id string) error {
  err := s.r.Remove(ctx, id)

  if nil != err {
    return err
  }

  s.setCache(ctx)

  return nil
}

func (s *tagsService) setCache(ctx context.Context) {
  s.cache = nil
  s.cache, _ = s.Get(ctx)
}

func (s *tagsService) hasCache() bool {
  return 0 < len(s.cache)
}
