package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestExperienceService_Get(t *testing.T) {
  const routine = "Get"

  t.Run("success", func(t *testing.T) {
    var ctx = context.Background()
    var exp = make([]*model.Experience, 0)

    var r = mocks.NewExperienceRepository()
    r.On(routine, ctx, true).Return(exp, nil)
    res, err := NewExperienceService(r).Get(ctx, true)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    r = mocks.NewExperienceRepository()
    r.On(routine, ctx, false).Return(exp, nil)
    res, err = NewExperienceService(r).Get(ctx, false)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    r = mocks.NewExperienceRepository()
    r.On(routine, ctx, false).Return(exp, nil)
    res, err = NewExperienceService(r).Get(ctx)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var ctx = context.Background()

    var r = mocks.NewExperienceRepository()
    r.On(routine, ctx, false).Return(nil, unexpected)
    res, err := NewExperienceService(r).Get(ctx)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}
