package mocks

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
  "github.com/stretchr/testify/mock"
)

type ProjectsRepository struct {
  mock.Mock
}

func NewProjectsRepository() *ProjectsRepository {
  return new(ProjectsRepository)
}

func (o *ProjectsRepository) Get(ctx context.Context, archived bool) (projects []*model.Project, err error) {
  var args = o.Called(ctx, archived)
  var arg0 = args.Get(0)
  if nil != arg0 {
    projects = arg0.([]*model.Project)
  }
  return projects, args.Error(1)
}

func (o *ProjectsRepository) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  var args = o.Called(ctx, id)
  var arg0 = args.Get(0)
  if nil != arg0 {
    project = arg0.(*model.Project)
  }
  return project, args.Error(1)
}

func (o *ProjectsRepository) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  var args = o.Called(ctx, slug)
  var arg0 = args.Get(0)
  if nil != arg0 {
    project = arg0.(*model.Project)
  }
  return project, args.Error(1)
}

func (o *ProjectsRepository) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  var args = o.Called(ctx, creation)
  return args.String(0), args.Error(1)
}

func (o *ProjectsRepository) Exists(ctx context.Context, id string) (err error) {
  var args = o.Called(ctx, id)
  return args.Error(0)
}

func (o *ProjectsRepository) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  var args = o.Called(ctx, id, update)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsRepository) Unarchive(ctx context.Context, id string) (unarchived bool, err error) {
  var args = o.Called(ctx, id)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsRepository) Remove(ctx context.Context, id string) (err error) {
  var args = o.Called(ctx, id)
  return args.Error(0)
}

func (o *ProjectsRepository) ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  var args = o.Called(ctx, projectID, technologyTagID)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsRepository) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  var args = o.Called(ctx, projectID, technologyTagID)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsRepository) RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error) {
  var args = o.Called(ctx, projectID, technologyTagID)
  return args.Bool(0), args.Error(1)
}

type ProjectsService struct {
  mock.Mock
}

func NewProjectsService() *ProjectsService {
  return new(ProjectsService)
}

func (o *ProjectsService) Get(ctx context.Context, archived ...bool) (projects []*model.Project, err error) {
  var args = o.Called(ctx, archived)
  var arg0 = args.Get(0)
  if nil != arg0 {
    projects = arg0.([]*model.Project)
  }
  return projects, args.Error(1)
}

func (o *ProjectsService) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  var args = o.Called(ctx, id)
  var arg0 = args.Get(0)
  if nil != arg0 {
    project = arg0.(*model.Project)
  }
  return project, args.Error(1)
}

func (o *ProjectsService) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  var args = o.Called(ctx, slug)
  var arg0 = args.Get(0)
  if nil != arg0 {
    project = arg0.(*model.Project)
  }
  return project, args.Error(1)
}

func (o *ProjectsService) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  var args = o.Called(ctx, creation)
  return args.String(0), args.Error(1)
}

func (o *ProjectsService) Exists(ctx context.Context, id string) (err error) {
  var args = o.Called(ctx, id)
  return args.Error(0)
}

func (o *ProjectsService) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  var args = o.Called(ctx, id, update)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsService) Unarchive(ctx context.Context, id string) (unarchived bool, err error) {
  var args = o.Called(ctx, id)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsService) Remove(ctx context.Context, id string) (err error) {
  var args = o.Called(ctx, id)
  return args.Error(0)
}

func (o *ProjectsService) ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  var args = o.Called(ctx, projectID, technologyTagID)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsService) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  var args = o.Called(ctx, projectID, technologyTagID)
  return args.Bool(0), args.Error(1)
}

func (o *ProjectsService) RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error) {
  var args = o.Called(ctx, projectID, technologyTagID)
  return args.Bool(0), args.Error(1)
}
