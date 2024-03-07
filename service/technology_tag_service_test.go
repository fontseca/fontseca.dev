package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestTechnologyTagService_Get(t *testing.T) {
  const routine = "Get"

  t.Run("success", func(t *testing.T) {
    var r = mocks.NewTechnologyTagRepository()
    var ctx = context.Background()
    r.On(routine, ctx).Return(make([]*model.TechnologyTag, 0), nil)
    res, err := NewTechnologyTagService(r).Get(ctx)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("got an error", func(t *testing.T) {
    var r = mocks.NewTechnologyTagRepository()
    var unexpected = errors.New("unexpected error")
    var ctx = context.Background()
    r.On(routine, ctx).Return(nil, unexpected)
    res, err := NewTechnologyTagService(r).Get(ctx)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}
