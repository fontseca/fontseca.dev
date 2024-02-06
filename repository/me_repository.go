package repository

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
)

type MeRepository interface {
  Get(ctx context.Context) (me model.Me, err error)
  Update(ctx context.Context, update transfer.MeUpdate) (ok bool, err error)
}
