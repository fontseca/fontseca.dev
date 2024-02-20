package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/transfer"
)

// ExperienceRepository provides methods for interacting with experience data in the database.
type ExperienceRepository interface {
  // Get retrieves a slice of experience. If hidden is true it returns all
  // the hidden experience records.
  Get(ctx context.Context, hidden ...bool) (experience []model.Experience, err error)

  // GetByID retrieves a single experience record by its ID.
  GetByID(ctx context.Context, id string) (experience *model.Experience, err error)

  // Save creates a new experience record with the provided creation data.
  Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error)

  // Update modifies an existing experience record with the provided update data.
  Update(ctx context.Context, update *transfer.ExperienceUpdate) (updated bool, err error)

  // Remove deletes an experience record by its ID.
  Remove(ctx context.Context, id string) error
}

type experienceRepository struct {
  db *sql.DB
}

// NewExperienceRepository creates a new ExperienceRepository instance associating db as its database.
func NewExperienceRepository(db *sql.DB) ExperienceRepository {
  return &experienceRepository{db}
}

func (r *experienceRepository) Get(ctx context.Context, hidden ...bool) (experience []model.Experience, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *experienceRepository) GetByID(ctx context.Context, id string) (experience *model.Experience, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *experienceRepository) Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *experienceRepository) Update(ctx context.Context, update *transfer.ExperienceUpdate) (updated bool, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *experienceRepository) Remove(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}
