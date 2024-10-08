package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "testing"
  "time"
)

type experienceRepositoryMockAPI struct {
  repository.ExperienceRepository
  returns []any
  errors  error
  called  bool
}

func (mock *experienceRepositoryMockAPI) List(context.Context, bool) ([]*model.Experience, error) {
  return mock.returns[0].([]*model.Experience), mock.errors
}

func TestExperienceService_Get(t *testing.T) {
  t.Run("success", func(t *testing.T) {
    var ctx = context.Background()
    var exp = make([]*model.Experience, 0)
    var r = &experienceRepositoryMockAPI{returns: []any{exp}}

    res, err := NewExperienceService(r).List(ctx, true)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    res, err = NewExperienceService(r).List(ctx, false)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    res, err = NewExperienceService(r).List(ctx)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var ctx = context.Background()
    var r = &experienceRepositoryMockAPI{returns: []any{[]*model.Experience(nil)}, errors: unexpected}
    res, err := NewExperienceService(r).List(ctx)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *experienceRepositoryMockAPI) Get(context.Context, string) (*model.Experience, error) {
  return mock.returns[0].(*model.Experience), mock.errors
}

func TestExperienceService_GetByID(t *testing.T) {
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var r = &experienceRepositoryMockAPI{returns: []any{new(model.Experience)}}
    var ctx = context.Background()
    res, err := NewExperienceService(r).Get(ctx, id)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &experienceRepositoryMockAPI{returns: []any{(*model.Experience)(nil)}, errors: unexpected}
    var ctx = context.Background()
    res, err := NewExperienceService(r).Get(ctx, id)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *experienceRepositoryMockAPI) Create(context.Context, *transfer.ExperienceCreation) (bool, error) {
  return mock.returns[0].(bool), mock.errors
}

func TestExperienceService_Save(t *testing.T) {
  var r = &experienceRepositoryMockAPI{returns: []any{true}}

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
    res, err := NewExperienceService(r).Create(ctx, &dirty)
    assert.NoError(t, err)
    assert.True(t, res)
  })

  t.Run("error on nil creation", func(t *testing.T) {
    var ctx = context.Background()
    res, err := NewExperienceService(r).Create(ctx, nil)
    require.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.False(t, res)
  })

  t.Run("creation.Starts validations", func(t *testing.T) {
    var creation = transfer.ExperienceCreation{Starts: 2020, Ends: 2023}

    t.Run("2017<creation.Starts<=current_year", func(t *testing.T) {
      t.Run("fails:creation.Starts=2016", func(t *testing.T) {
        creation.Starts = 2016
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })

      t.Run("meets:creation.Starts=2020", func(t *testing.T) {
        creation.Starts = 2020
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("meets:creation.Starts=current_year", func(t *testing.T) {
        creation.Starts = time.Now().Year()
        creation.Ends = creation.Starts
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("fails:creation.Starts=1+current_year", func(t *testing.T) {
        creation.Starts = 1 + time.Now().Year()
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })
    })
  })

  t.Run("creation.Ends validations", func(t *testing.T) {
    var creation = transfer.ExperienceCreation{Starts: 2020, Ends: 2023}

    t.Run("creation.Starts<=creation.Ends<=current_year", func(t *testing.T) {
      t.Run("fails:creation.Ends=2016", func(t *testing.T) {
        creation.Ends = 2016
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })

      t.Run("meets:creation.Ends=creation.Starts", func(t *testing.T) {
        creation.Ends = creation.Starts
        var ctx = context.Background()

        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("meets:creation.Ends=1+creation.Starts", func(t *testing.T) {
        creation.Ends = 1 + creation.Starts
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("meets:creation.Ends=current_year", func(t *testing.T) {
        creation.Ends = time.Now().Year()
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("fails:creation.Ends=1+current_year", func(t *testing.T) {
        creation.Ends = 1 + time.Now().Year()
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })
    })
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &experienceRepositoryMockAPI{returns: []any{false}, errors: unexpected}
    var ctx = context.Background()
    res, err := NewExperienceService(r).Create(ctx, new(transfer.ExperienceCreation))
    assert.False(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *experienceRepositoryMockAPI) Update(context.Context, string, *transfer.ExperienceUpdate) (bool, error) {
  return mock.returns[0].(bool), mock.errors
}

func TestExperienceService_Update(t *testing.T) {
  var id = "   \t\n{7d7d4da0-093a-443b-b041-2da650381220}\n\t   "
  var ctx = context.Background()
  var r = &experienceRepositoryMockAPI{returns: []any{true}}

  t.Run("success", func(t *testing.T) {
    var expected = transfer.ExperienceUpdate{
      Starts:   2020,
      Ends:     2023,
      JobTitle: "JobTitle",
      Company:  "Company",
      Country:  "Country",
      Summary:  "Summary",
    }
    var dirty = transfer.ExperienceUpdate{
      Starts:   expected.Starts,
      Ends:     expected.Ends,
      JobTitle: " \n\t " + expected.JobTitle + " \n\t ",
      Company:  " \n\t " + expected.Company + " \n\t ",
      Country:  " \n\t " + expected.Country + " \n\t ",
      Summary:  " \n\t " + expected.Summary + " \n\t ",
    }
    res, err := NewExperienceService(r).Update(ctx, id, &dirty)
    assert.True(t, res)
    assert.NoError(t, err)
  })

  t.Run("update.Starts validations", func(t *testing.T) {
    var update = transfer.ExperienceUpdate{Starts: 2020}

    t.Run("2017<update.Starts<=current_year", func(t *testing.T) {
      t.Run("fails:update.Starts=2016", func(t *testing.T) {
        update.Starts = 2016
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })

      t.Run("meets:update.Starts=2020", func(t *testing.T) {
        update.Starts = 2020
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("meets:update.Starts=current_year", func(t *testing.T) {
        update.Starts = time.Now().Year()
        update.Ends = update.Starts
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("fails:update.Starts=1+current_year", func(t *testing.T) {
        update.Starts = 1 + time.Now().Year()
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })
    })
  })

  t.Run("update.Ends validations", func(t *testing.T) {
    var update = transfer.ExperienceUpdate{Ends: 2023}

    t.Run("update.Starts<=update.Ends<=current_year", func(t *testing.T) {
      t.Run("fails:update.Ends=2016", func(t *testing.T) {
        update.Ends = 2016
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })

      t.Run("meets:update.Ends=update.Starts", func(t *testing.T) {
        update.Starts = 2020
        update.Ends = update.Starts
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("meets:update.Ends=1+update.Starts", func(t *testing.T) {
        update.Ends = 1 + update.Starts
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("meets:update.Ends=current_year", func(t *testing.T) {
        update.Ends = time.Now().Year()
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
        assert.True(t, res)
      })

      t.Run("fails:update.Starts>update.Ends", func(t *testing.T) {
        update.Starts = 2020
        update.Ends = 2019
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })

      t.Run("fails:update.Ends=1+current_year", func(t *testing.T) {
        update.Ends = 1 + time.Now().Year()
        res, err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.False(t, res)
      })
    })
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &experienceRepositoryMockAPI{returns: []any{false}, errors: unexpected}
    res, err := NewExperienceService(r).Update(ctx, id, new(transfer.ExperienceUpdate))
    assert.False(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *experienceRepositoryMockAPI) Remove(context.Context, string) error {
  return mock.errors
}

func TestExperienceService_Remove(t *testing.T) {
  var id = "   \t\n{7d7d4da0-093a-443b-b041-2da650381220}\n\t   "
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var r = &experienceRepositoryMockAPI{errors: nil}
    err := NewExperienceService(r).Remove(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &experienceRepositoryMockAPI{errors: unexpected}
    err := NewExperienceService(r).Remove(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}
