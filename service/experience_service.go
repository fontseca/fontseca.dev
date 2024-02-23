package service

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
)

// ExperienceService provides methods for interacting with
// experience data at a higher level.
type ExperienceService interface {
  // Get returns a slice of experience models and an error if
  // the operation fails. If hidden is true it returns all the
  // hidden experience records.
  Get(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error)

  // GetByID retrieves a single experience record by its ID.
  GetByID(ctx context.Context, id string) (experience *model.Experience, err error)

  // Save creates a new experience record with the provided creation data.
  // It returns a boolean indicating whether the experience was successfully
  // saved and an error if something went wrong.
  Save(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error)

  // Update modifies an existing experience record with the provided update data.
  // It returns a boolean indicating whether the experience was successfully updated
  // and an error if something went wrong.
  Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error)

  // Remove deletes an experience record by its ID.
  // It returns an error if the operation fails; for example,
  // if the record does not exist.
  Remove(ctx context.Context, id string) error
}
