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
  "strings"
  "testing"
  "time"
)

type experienceRepositoryMockAPI struct {
  repository.ExperienceRepository
  t         *testing.T
  arguments []any
  returns   []any
  errors    error
  called    bool
}

func (mock *experienceRepositoryMockAPI) List(context.Context, bool) ([]*model.Experience, error) {
  return mock.returns[0].([]*model.Experience), mock.errors
}

func TestExperienceService_List(t *testing.T) {
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

func (mock *experienceRepositoryMockAPI) Get(context.Context, string, bool) (*model.Experience, error) {
  return mock.returns[0].(*model.Experience), mock.errors
}

func TestExperienceService_Get(t *testing.T) {
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

func (mock *experienceRepositoryMockAPI) Create(_ context.Context, t *transfer.ExperienceCreation) (string, error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], t)
  }
  return mock.returns[0].(string), mock.errors
}

func TestExperienceService_Create(t *testing.T) {
  var id = "7e68dc64-3a3c-4ea2-b099-8938f0ce54d9"
  var r = &experienceRepositoryMockAPI{returns: []any{id}}

  t.Run("success", func(t *testing.T) {
    var expected = transfer.ExperienceCreation{
      Starts:          "2020-01-02",
      Ends:            "2023-01-02",
      JobTitle:        "JobTitle",
      Company:         "Company",
      CompanyHomepage: "http://foo.com",
      Country:         "Country",
      Summary:         "Summary",
    }
    var dirty = transfer.ExperienceCreation{
      Starts:          " \n\t " + expected.Starts + " \n\t ",
      Ends:            " \n\t " + expected.Ends + " \n\t ",
      JobTitle:        " \n\t " + expected.JobTitle + " \n\t ",
      Company:         " \n\t " + expected.Company + " \n\t ",
      CompanyHomepage: " \n\t " + expected.CompanyHomepage + " \n\t ",
      Country:         " \n\t " + expected.Country + " \n\t ",
      Summary:         " \n\t " + expected.Summary + " \n\t ",
    }
    var ctx = context.Background()
    r := &experienceRepositoryMockAPI{t: t, arguments: []any{ctx, &expected}, returns: []any{id}}
    res, err := NewExperienceService(r).Create(ctx, &dirty)
    assert.NoError(t, err)
    assert.Equal(t, id, res)
  })

  t.Run("error on nil creation", func(t *testing.T) {
    var ctx = context.Background()
    res, err := NewExperienceService(r).Create(ctx, nil)
    require.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.Empty(t, res)
  })

  t.Run("company homepage is an invalid url", func(t *testing.T) {
    var ctx = context.Background()
    res, err := NewExperienceService(r).Create(ctx, &transfer.ExperienceCreation{CompanyHomepage: "xxx"})
    assert.ErrorContains(t, err, "There was an error parsing the requested URL")
    assert.Empty(t, res)
  })

  t.Run("company_homepage is too long", func(t *testing.T) {
    var creation = transfer.ExperienceCreation{}

    creation.CompanyHomepage = "https://" + strings.Repeat("x", 2036) + ".com"

    res, err := NewExperienceService(r).Create(context.Background(), &creation)
    assert.NoError(t, err)
    assert.Equal(t, id, res)

    creation.CompanyHomepage = "https://" + strings.Repeat("x", 1+2036) + ".com"

    res, err = NewExperienceService(r).Create(context.Background(), &creation)
    assert.Error(t, err)
    assert.Empty(t, res)
  })

  t.Run("creation.Starts validations", func(t *testing.T) {
    t.Run("invalid format", func(t *testing.T) {
      _, err := NewExperienceService(r).Create(context.Background(), &transfer.ExperienceCreation{Starts: "2020/01/02"})
      assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
    })

    t.Run("out of range", func(t *testing.T) {
      _, err := NewExperienceService(r).Create(context.Background(), &transfer.ExperienceCreation{Starts: "2020-50-02"})
      assert.ErrorContains(t, err, "The value provided for 'starts' (2020-50-02) is out of range")
    })

    var creation = transfer.ExperienceCreation{Starts: "2020-01-02", Ends: "2023-01-02"}

    t.Run("2017<creation.Starts<=current_year", func(t *testing.T) {
      t.Run("fails:creation.Starts=2016", func(t *testing.T) {
        creation.Starts = "2016-01-02"
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.Empty(t, res)
      })

      t.Run("meets:creation.Starts=2020", func(t *testing.T) {
        creation.Starts = "2020-01-02"
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.Equal(t, id, res)
      })

      t.Run("meets:creation.Starts=now", func(t *testing.T) {
        creation.Starts = time.Now().Format(time.DateOnly)
        creation.Ends = creation.Starts
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.Equal(t, id, res)
      })

      t.Run("fails:creation.Starts=1+current_year", func(t *testing.T) {
        now := time.Now()
        creation.Starts = time.Date(1+now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.Empty(t, res)
      })
    })
  })

  t.Run("creation.Ends validations", func(t *testing.T) {
    t.Run("invalid format", func(t *testing.T) {
      _, err := NewExperienceService(r).Create(context.Background(), &transfer.ExperienceCreation{Ends: "2020/01/02"})
      assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
    })

    t.Run("out of range", func(t *testing.T) {
      _, err := NewExperienceService(r).Create(context.Background(), &transfer.ExperienceCreation{Ends: "2020-50-02"})
      assert.ErrorContains(t, err, "The value provided for 'ends' (2020-50-02) is out of range")
    })

    var creation = transfer.ExperienceCreation{Starts: "2020-01-02", Ends: "2023-01-02"}

    t.Run("creation.Starts<=creation.Ends<=current_year", func(t *testing.T) {
      t.Run("fails:creation.Ends=2016", func(t *testing.T) {
        creation.Ends = "2016-01-02"
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.Empty(t, res)
      })

      t.Run("meets:creation.Ends=creation.Starts", func(t *testing.T) {
        creation.Ends = creation.Starts
        var ctx = context.Background()

        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.Equal(t, id, res)
      })

      t.Run("meets:creation.Ends=1+creation.Starts", func(t *testing.T) {
        start, _ := time.Parse(time.DateOnly, creation.Starts)
        creation.Ends = time.Date(1+start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.Equal(t, id, res)
      })

      t.Run("meets:creation.Ends=current_year", func(t *testing.T) {
        creation.Ends = time.Now().Format(time.DateOnly)
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.NoError(t, err)
        assert.Equal(t, id, res)
      })

      t.Run("fails:creation.Ends=1+current_year", func(t *testing.T) {
        now := time.Now()
        creation.Ends = time.Date(1+now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
        var ctx = context.Background()
        res, err := NewExperienceService(r).Create(ctx, &creation)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        assert.Empty(t, res)
      })
    })
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &experienceRepositoryMockAPI{returns: []any{""}, errors: unexpected}
    var ctx = context.Background()
    res, err := NewExperienceService(r).Create(ctx, new(transfer.ExperienceCreation))
    assert.Empty(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

type experienceRepositoryMockAPIForUpdate struct {
  experienceRepositoryMockAPI
  getReturns []any
}

func (mock *experienceRepositoryMockAPIForUpdate) Update(_ context.Context, id string, t *transfer.ExperienceUpdate) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
    require.Equal(mock.t, mock.arguments[2], t)
  }
  return mock.errors
}

func (mock *experienceRepositoryMockAPIForUpdate) Get(context.Context, string, bool) (*model.Experience, error) {
  return mock.getReturns[0].(*model.Experience), mock.errors
}

func TestExperienceService_Update(t *testing.T) {
  var id = "   \t\n{7d7d4da0-093a-443b-b041-2da650381220}\n\t   "
  var ctx = context.Background()
  var r = &experienceRepositoryMockAPIForUpdate{getReturns: []any{&model.Experience{}}}

  t.Run("success", func(t *testing.T) {
    var expected = transfer.ExperienceUpdate{
      Starts:          "2020-01-02",
      Ends:            "2023-01-02",
      JobTitle:        "JobTitle",
      Company:         "Company",
      CompanyHomepage: "http://foo.com.",
      Country:         "Country",
      Summary:         "Summary",
    }
    var dirty = transfer.ExperienceUpdate{
      Starts:          expected.Starts,
      Ends:            expected.Ends,
      JobTitle:        " \n\t " + expected.JobTitle + " \n\t ",
      Company:         " \n\t " + expected.Company + " \n\t ",
      CompanyHomepage: " \n\t " + expected.CompanyHomepage + " \n\t ",
      Country:         " \n\t " + expected.Country + " \n\t ",
      Summary:         " \n\t " + expected.Summary + " \n\t ",
    }
    r := &experienceRepositoryMockAPIForUpdate{
      experienceRepositoryMockAPI: experienceRepositoryMockAPI{t: t, arguments: []any{ctx, "7d7d4da0-093a-443b-b041-2da650381220", &expected}},
      getReturns:                  []any{&model.Experience{}},
    }
    err := NewExperienceService(r).Update(ctx, id, &dirty)
    assert.NoError(t, err)
  })

  t.Run("company homepage is an invalid url", func(t *testing.T) {
    var ctx = context.Background()
    err := NewExperienceService(r).Update(ctx, id, &transfer.ExperienceUpdate{CompanyHomepage: "xxx"})
    assert.ErrorContains(t, err, "There was an error parsing the requested URL")
  })

  t.Run("company_homepage is too long", func(t *testing.T) {
    var update = transfer.ExperienceUpdate{}

    update.CompanyHomepage = "https://" + strings.Repeat("x", 2036) + ".com"

    err := NewExperienceService(r).Update(context.Background(), id, &update)
    assert.NoError(t, err)

    update.CompanyHomepage = "https://" + strings.Repeat("x", 1+2036) + ".com"

    err = NewExperienceService(r).Update(context.Background(), id, &update)
    assert.Error(t, err)
  })

  t.Run("update.Starts validations", func(t *testing.T) {
    t.Run("invalid format", func(t *testing.T) {
      err := NewExperienceService(r).Update(ctx, id, &transfer.ExperienceUpdate{Starts: "2020/01/02"})
      assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
    })

    t.Run("out of range", func(t *testing.T) {
      err := NewExperienceService(r).Update(ctx, id, &transfer.ExperienceUpdate{Starts: "2020-50-02"})
      assert.ErrorContains(t, err, "The value provided for 'starts' (2020-50-02) is out of range")
    })

    var update = transfer.ExperienceUpdate{Starts: "2020-01-02"}

    t.Run("2017<update.Starts<=current_year", func(t *testing.T) {
      t.Run("fails:update.Starts=2016", func(t *testing.T) {
        update.Starts = "2016-01-02"
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
      })

      t.Run("meets:update.Starts=2020", func(t *testing.T) {
        update.Starts = "2020-01-02"
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
      })

      t.Run("meets:update.Starts=current_year", func(t *testing.T) {
        update.Starts = time.Now().Format(time.DateOnly)
        update.Ends = update.Starts
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
      })

      t.Run("fails:update.Starts=1+current_year", func(t *testing.T) {
        now := time.Now()
        update.Starts = time.Date(1+now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
      })
    })
  })

  t.Run("update.Ends validations", func(t *testing.T) {
    t.Run("invalid format", func(t *testing.T) {
      err := NewExperienceService(r).Update(ctx, id, &transfer.ExperienceUpdate{Ends: "2020/01/02"})
      assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
    })

    t.Run("out of range", func(t *testing.T) {
      err := NewExperienceService(r).Update(ctx, id, &transfer.ExperienceUpdate{Ends: "2020-50-02"})
      assert.ErrorContains(t, err, "The value provided for 'ends' (2020-50-02) is out of range")
    })
    var update = transfer.ExperienceUpdate{Ends: "2023-01-02"}

    t.Run("update.Starts<=update.Ends<=current_year", func(t *testing.T) {
      t.Run("fails:update.Ends=2016", func(t *testing.T) {
        update.Ends = "2016-01-02"
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
      })

      t.Run("meets:update.Ends=update.Starts", func(t *testing.T) {
        update.Starts = "2020-01-02"
        update.Ends = update.Starts
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
      })

      t.Run("meets:update.Ends=1+update.Starts", func(t *testing.T) {
        starts, _ := time.Parse(time.DateOnly, update.Starts)
        update.Ends = time.Date(1+starts.Year(), starts.Month(), starts.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
      })

      t.Run("meets:update.Ends=current_year", func(t *testing.T) {
        update.Ends = time.Now().Format(time.DateOnly)
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.NoError(t, err)
      })

      t.Run("fails:update.Starts>update.Ends", func(t *testing.T) {
        prev := r.getReturns
        update.Ends = "2019-01-02"
        e := &model.Experience{}
        e.Starts, _ = time.Parse(time.DateOnly, "2020-01-02")
        r.getReturns = []any{e}
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
        r.getReturns = prev
      })

      t.Run("fails:update.Ends=1+current_year", func(t *testing.T) {
        now := time.Now()
        update.Ends = time.Date(1+now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
        err := NewExperienceService(r).Update(ctx, id, &update)
        assert.ErrorContains(t, err, "The provided data does not meet the required validation criteria")
      })
    })
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &experienceRepositoryMockAPIForUpdate{experienceRepositoryMockAPI: experienceRepositoryMockAPI{errors: unexpected}}
    err := NewExperienceService(r).Update(ctx, id, new(transfer.ExperienceUpdate))
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
