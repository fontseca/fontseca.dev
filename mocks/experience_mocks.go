package mocks

import (
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/mock"
)

type ExperienceRepository struct {
  mock.Mock
}

func NewExperienceRepository() *ExperienceRepository {
  return new(ExperienceRepository)
}

func (o *ExperienceRepository) Get(ctx context.Context, hidden bool) (experience []*model.Experience, err error) {
  var args = o.Called(ctx, hidden)
  var arg0 = args.Get(0)
  if nil != arg0 {
    experience = arg0.([]*model.Experience)
  }
  return experience, args.Error(1)
}

func (o *ExperienceRepository) GetByID(ctx context.Context, id string) (experience *model.Experience, err error) {
  var args = o.Called(ctx, id)
  var arg0 = args.Get(0)
  if nil != arg0 {
    experience = arg0.(*model.Experience)
  }
  return experience, args.Error(1)
}

func (o *ExperienceRepository) Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  var args = o.Called(ctx, creation)
  return args.Bool(0), args.Error(1)
}

func (o *ExperienceRepository) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error) {
  var args = o.Called(ctx, id, update)
  return args.Bool(0), args.Error(1)
}

func (o *ExperienceRepository) Remove(ctx context.Context, id string) error {
  var args = o.Called(ctx, id)
  return args.Error(0)
}

type ExperienceService struct {
  mock.Mock
}

func NewExperienceService() *ExperienceService {
  return new(ExperienceService)
}

func (o *ExperienceService) Get(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error) {
  var args = o.Called(ctx, hidden)
  var arg0 = args.Get(0)
  if nil != arg0 {
    experience = arg0.([]*model.Experience)
  }
  return experience, args.Error(1)
}

func (o *ExperienceService) GetByID(ctx context.Context, id string) (experience *model.Experience, err error) {
  var args = o.Called(ctx, id)
  var arg0 = args.Get(0)
  if nil != arg0 {
    experience = arg0.(*model.Experience)
  }
  return experience, args.Error(1)
}

func (o *ExperienceService) Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  var args = o.Called(ctx, creation)
  return args.Bool(0), args.Error(1)
}

func (o *ExperienceService) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error) {
  var args = o.Called(ctx, id, update)
  return args.Bool(0), args.Error(1)
}

func (o *ExperienceService) Remove(ctx context.Context, id string) error {
  var args = o.Called(ctx, id)
  return args.Error(0)
}
