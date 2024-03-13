package service

import (
  "context"
  "fontseca/model"
  "fontseca/repository"
  "fontseca/transfer"
)

// ProjectsService provides methods for interacting with projects
// data at a higher level and allows extra validation.
type ProjectsService interface {
  // Get retrieves a slice of projects.
  Get(ctx context.Context, archived ...bool) (projects []*model.Project, err error)

  // GetByID retrieves a single project by its ID.
  GetByID(ctx context.Context, id string) (project *model.Project, err error)

  // Add creates a project record with the provided creation data.
  Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error)

  // Exists checks whether a given project exists in the database.
  // If it does, it returns nil; otherwise a not found error.
  Exists(ctx context.Context, id string) (err error)

  // Update modifies an existing project record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error)

  // Remove deletes an existing project. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)

  // ContainsTechnologyTag checks whether technologyTagID belongs to projectID.
  ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error)

  // AddTechnologyTag adds an existing technology tag that will belong to the project represented by projectID .
  AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error)

  // RemoveTechnologyTag removes a technology tag that belongs to the project represented by projectID.
  RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error)
}

type projectsService struct {
  r repository.ProjectsRepository
}

func NewProjectsService(repository repository.ProjectsRepository) ProjectsService {
  return &projectsService{repository}
}

func (s *projectsService) Get(ctx context.Context, archived ...bool) (projects []*model.Project, err error) {
  var a = false
  if 0 != len(archived) && archived[0] {
    a = true
  }
  return s.r.Get(ctx, a)
}

func (s *projectsService) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  if err = validateUUID(&id); err != nil {
    return nil, err
  }
  return s.r.GetByID(ctx, id)
}

func (s *projectsService) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *projectsService) Exists(ctx context.Context, id string) (err error) {
  // TODO implement me
  panic("implement me")
}

func (s *projectsService) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *projectsService) Remove(ctx context.Context, id string) (err error) {
  // TODO implement me
  panic("implement me")
}

func (s *projectsService) ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *projectsService) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (s *projectsService) RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error) {
  // TODO implement me
  panic("implement me")
}
