package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
)

type technologyTagRepositoryAPI interface {
  Get(context.Context) ([]*model.TechnologyTag, error)
  Add(context.Context, *transfer.TechnologyTagCreation) (string, error)
  Exists(context.Context, string) error
  Update(context.Context, string, *transfer.TechnologyTagUpdate) (bool, error)
  Remove(context.Context, string) error
}

// TechnologyTagService provides methods for interacting with technology
// tags data at a higher level and extra validation.
type TechnologyTagService struct {
  r technologyTagRepositoryAPI
}

func NewTechnologyTagService(repository technologyTagRepositoryAPI) *TechnologyTagService {
  return &TechnologyTagService{repository}
}

// Get retrieves a slice of technology tags.
func (s *TechnologyTagService) Get(ctx context.Context) (technologies []*model.TechnologyTag, err error) {
  return s.r.Get(ctx)
}

// Add creates a new technology tag record with the provided creation data.
func (s *TechnologyTagService) Add(ctx context.Context, creation *transfer.TechnologyTagCreation) (id string, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return "", err
  }
  creation.Name = strings.TrimSpace(creation.Name)
  if 64 < len(creation.Name) {
    return "", problem.NewValidation([3]string{"name", "max", "64"})
  }
  return s.r.Add(ctx, creation)
}

// Exists checks whether a technology tag exists in the database.
// If it does, it returns nil; otherwise a not found error.
func (s *TechnologyTagService) Exists(ctx context.Context, id string) (err error) {
  err = validateUUID(&id)
  if nil != err {
    return err
  }
  return s.r.Exists(ctx, id)
}

// Update modifies an existing technology tag record with the provided update data.
func (s *TechnologyTagService) Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error) {
  if nil == update {
    err = errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return false, err
  }
  err = validateUUID(&id)
  if nil != err {
    return false, err
  }
  update.Name = strings.TrimSpace(update.Name)
  if 64 < len(update.Name) {
    return false, problem.NewValidation([3]string{"name", "max", "64"})
  }
  return s.r.Update(ctx, id, update)
}

// Remove deletes an existing technology tag. If not found, returns a not found error.
func (s *TechnologyTagService) Remove(ctx context.Context, id string) (err error) {
  err = validateUUID(&id)
  if nil != err {
    return err
  }
  return s.r.Remove(ctx, id)
}
