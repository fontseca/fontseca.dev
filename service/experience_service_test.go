package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "github.com/google/uuid"
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

func TestExperienceService_GetByID(t *testing.T) {
  const routine = "GetByID"
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var r = mocks.NewExperienceRepository()
    var ctx = context.Background()
    r.On(routine, ctx, id).Return(new(model.Experience), nil)
    res, err := NewExperienceService(r).GetByID(ctx, id)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewExperienceRepository()
    var ctx = context.Background()
    r.On(routine, ctx, id).Return(nil, unexpected)
    res, err := NewExperienceService(r).GetByID(ctx, id)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}
