package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "fontseca/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
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

func TestExperienceService_Save(t *testing.T) {
  const routine = "Save"

  t.Run("success", func(t *testing.T) {
    var expected = transfer.ExperienceCreation{
      Starts:   2020,
      Ends:     2023,
      JobTitle: "JobTitle",
      Company:  "Company",
      Country:  "Country",
      Summary:  "Summary",
    }
    var dirty = transfer.ExperienceCreation{
      Starts:   expected.Starts,
      Ends:     expected.Ends,
      JobTitle: " \n\t " + expected.JobTitle + " \n\t ",
      Company:  " \n\t " + expected.Company + " \n\t ",
      Country:  " \n\t " + expected.Country + " \n\t ",
      Summary:  " \n\t " + expected.Summary + " \n\t ",
    }
    var ctx = context.Background()
    var r = mocks.NewExperienceRepository()
    r.On(routine, ctx, &expected).Return(true, nil)
    res, err := NewExperienceService(r).Save(ctx, &dirty)
    assert.NoError(t, err)
    assert.True(t, res)
  })

  t.Run("error on nil creation", func(t *testing.T) {
    var r = mocks.NewExperienceRepository()
    var ctx = context.Background()
    r.AssertNotCalled(t, routine)
    res, err := NewExperienceService(r).Save(ctx, nil)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.False(t, res)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewExperienceRepository()
    var ctx = context.Background()
    r.On(routine, ctx, mock.AnythingOfType("*transfer.ExperienceCreation")).Return(false, unexpected)
    res, err := NewExperienceService(r).Save(ctx, new(transfer.ExperienceCreation))
    assert.False(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}
