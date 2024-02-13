package mocks

import (
  "context"
  "fontseca/model"
  "fontseca/transfer"
  "github.com/stretchr/testify/mock"
)

type MeRepository struct {
  mock.Mock
}

func NewMeRepository() *MeRepository {
  return new(MeRepository)
}

func (o *MeRepository) Register(ctx context.Context) {
  o.Called(ctx)
}

func (o *MeRepository) Get(ctx context.Context) (me *model.Me, err error) {
  var args = o.Called(ctx)
  var arg0 = args.Get(0)
  if nil != arg0 {
    me = arg0.(*model.Me)
  }
  return me, args.Error(1)
}

func (o *MeRepository) Update(ctx context.Context, update *transfer.MeUpdate) (ok bool, err error) {
  var args = o.Called(ctx, update)
  return args.Bool(0), args.Error(1)
}
