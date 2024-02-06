package repository

import (
  "context"
  "database/sql"
  "fontseca/model"
  "fontseca/transfer"
)

type MeRepository interface {
  Get(ctx context.Context) (me model.Me, err error)
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
