package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/transfer"
)

// ProjectsRepository provides methods for interacting with project data in the database.
type ProjectsRepository interface {
  // Get retrieves a slice of project types.
  Get(ctx context.Context, archived bool) (projects []*model.Project, err error)

  // GetByID retrieves a project type by its ID.
  GetByID(ctx context.Context, id string) (project *model.Project, err error)

  // Add creates a project record with the provided creation data.projectID
  Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error)

  // Update modifies an existing project record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error)

  // Remove deletes an existing project type. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)

  // AddTechnologyTag adds an existing technology tag that will belong to the project represented by projectID .
  AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error)

  // RemoveTechnologyTag removes a technology tag that belongs to the project represented by projectID.
  RemoveTechnologyTag(ctx context.Context, projectID, technologyID string) (err error)
}

type projectsRepository struct {
  db *sql.DB
}

func NewProjectsRepository(db *sql.DB) ProjectsRepository {
  return &projectsRepository{db}
}

func (r *projectsRepository) Get(ctx context.Context, archived bool) (projects []*model.Project, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) Remove(ctx context.Context, id string) (err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *projectsRepository) RemoveTechnologyTag(ctx context.Context, projectID, technologyID string) (err error) {
  // TODO implement me
  panic("implement me")
}
