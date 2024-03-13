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

func TestProjectService_Get(t *testing.T) {
  const routine = "Get"
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var projects = make([]*model.Project, 0)

    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, true).Return(projects, nil)
    res, err := NewProjectsService(r).Get(ctx, true)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    r = mocks.NewProjectsRepository()
    r.On(routine, ctx, false).Return(projects, nil)
    res, err = NewProjectsService(r).Get(ctx, false)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    r = mocks.NewProjectsRepository()
    r.On(routine, ctx, false).Return(projects, nil)
    res, err = NewProjectsService(r).Get(ctx)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, false).Return(nil, unexpected)
    res, err := NewProjectsService(r).Get(ctx)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestProjectService_GetByID(t *testing.T) {
  const routine = "GetByID"
  var id = uuid.New().String()
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var project = new(model.Project)
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id).Return(project, nil)
    res, err := NewProjectsService(r).GetByID(ctx, id)
    assert.Equal(t, project, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id).Return(nil, unexpected)
    res, err := NewProjectsService(r).GetByID(ctx, id)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}
