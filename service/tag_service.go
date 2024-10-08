package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
)

type tagsRepositoryAPI interface {
  Create(context.Context, *transfer.TagCreation) error
  List(context.Context) ([]*model.Tag, error)
  Update(context.Context, string, *transfer.TagUpdate) error
  Remove(context.Context, string) error
}

// TagsService is a high level provider for tags.
type TagsService struct {
  cache []*model.Tag
  r     tagsRepositoryAPI
}

func NewTagsService(r tagsRepositoryAPI) *TagsService {
  return &TagsService{
    cache: nil,
    r:     r,
  }
}

// Create adds a new tag.
func (s *TagsService) Create(ctx context.Context, creation *transfer.TagCreation) error {
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

// List retrieves all the tags.
func (s *TagsService) List(ctx context.Context) (tags []*model.Tag, err error) {
  if s.hasCache() {
    return s.cache, nil
  }

  tags, err = s.r.List(ctx)

  if nil != err {
    return nil, err
  }

  s.cache = tags

  return tags, err
}

// Update updates an existing tag.
func (s *TagsService) Update(ctx context.Context, id string, update *transfer.TagUpdate) error {
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

// Remove removes a tag and detaches it from any
// article that currently uses it.
func (s *TagsService) Remove(ctx context.Context, id string) error {
  err := s.r.Remove(ctx, id)

  if nil != err {
    return err
  }

  s.setCache(ctx)

  return nil
}

func (s *TagsService) setCache(ctx context.Context) {
  s.cache = nil
  s.cache, _ = s.List(ctx)
}

func (s *TagsService) hasCache() bool {
  return 0 < len(s.cache)
}
