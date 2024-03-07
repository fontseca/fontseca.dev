package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "fontseca/problem"
  "fontseca/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
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

func TestTechnologyTagService_Add(t *testing.T) {
  const routine = "Add"
  var creation = &transfer.TechnologyTagCreation{Name: "Technology Tag Name"}
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var dirty = &transfer.TechnologyTagCreation{Name: "  \n\t\n  " + creation.Name + "  \n\t\n  "}
    var id = uuid.New().String()
    var r = mocks.NewTechnologyTagRepository()
    r.On(routine, ctx, creation).Return(id, nil)
    res, err := NewTechnologyTagService(r).Add(ctx, dirty)
    assert.Equal(t, id, res)
    assert.NoError(t, err)
  })

  t.Run("error on nil creation", func(t *testing.T) {
    var r = mocks.NewTechnologyTagRepository()
    r.AssertNotCalled(t, routine)
    res, err := NewTechnologyTagService(r).Add(ctx, nil)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.Empty(t, res)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewTechnologyTagRepository()
    r.On(routine, ctx, mock.Anything).Return("", unexpected)
    res, err := NewTechnologyTagService(r).Add(ctx, creation)
    assert.Empty(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestTechnologyTagService_Exists(t *testing.T) {
  const routine = "Exists"
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success: does exist", func(t *testing.T) {
    var r = mocks.NewTechnologyTagRepository()
    r.On(routine, ctx, id).Return(nil)
    err := NewTechnologyTagService(r).Exists(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("success: does not exist", func(t *testing.T) {
    var r = mocks.NewTechnologyTagRepository()
    var p = problem.NewNotFound(id, "technology_tag")
    r.On(routine, ctx, id).Return(p)
    err := NewTechnologyTagService(r).Exists(ctx, id)
    assert.ErrorAs(t, err, &p)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewTechnologyTagRepository()
    r.On(routine, ctx, mock.Anything).Return(unexpected)
    err := NewTechnologyTagService(r).Exists(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}
