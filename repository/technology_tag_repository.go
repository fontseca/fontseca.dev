package repository

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
)

// TechnologyTagRepository provides methods for interacting with technology tags data in the database.
type TechnologyTagRepository interface {
  // Get retrieves a slice of technology tags.
  Get(ctx context.Context) (technologies []*model.TechnologyTag, err error)

  // Add creates a new technology tag record with the provided creation data.
  Add(ctx context.Context, creation *transfer.TechnologyTagCreation) (id string, err error)

  // Update modifies an existing technology tag record with the provided update data.
  Update(ctx context.Context, id string, update *transfer.TechnologyTagUpdate) (updated bool, err error)

  // Remove deletes an existing technology tag. If not found, returns a not found error.
  Remove(ctx context.Context, id string) (err error)
}
