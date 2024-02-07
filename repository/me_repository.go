package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/transfer"
)

// MeRepository is an abstraction of the database that provides
// actions for model.Me management.
type MeRepository interface {
  // Get retrieves the information of my profile.
  Get(ctx context.Context) (me *model.Me, err error)

  // Update updates the information of my profile.
  Update(ctx context.Context, update transfer.MeUpdate) (ok bool, err error)
}

type meRepositoryImpl struct {
  db *sql.DB
}

func NewRepository(db *sql.DB) MeRepository {
  return &meRepositoryImpl{db}
}

func (r *meRepositoryImpl) Get(ctx context.Context) (me model.Me, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *meRepositoryImpl) Update(ctx context.Context, update transfer.MeUpdate) (ok bool, err error) {
  // TODO implement me
  panic("implement me")
}
