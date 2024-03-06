package service

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
)

// TechnologyTagService provides methods for interacting with technology
// tags data at a higher level and extra validation.
type TechnologyTagService interface {
  // Get retrieves a slice of technology tags.
  Get(ctx context.Context) (technologies []*model.TechnologyTag, err error)

  // Add creates a new technology tag record with the provided creation data.
  Add(ctx context.Context, creation *transfer.TechnologyTagCreation) (id string, err error)

  // Exists checks whether a technology tag exists in the database.
  // If it does, it returns nil; otherwise a not found error.
  Exists(ctx context.Context, id string) (err error)

  // Update modifies an existing technology tag record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error)

  // Remove deletes an existing technology tag. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)
}
