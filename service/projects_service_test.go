package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "strings"
  "testing"
)

type projectsRepositoryMockAPI struct {
  projectsRepositoryAPI
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

type technologyTagsServiceMockAPI struct {
  technologyTagsServiceAPI
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

var sentinelTechnologyTagsService = &technologyTagsServiceMockAPI{}

func (mock *projectsRepositoryMockAPI) List(context.Context, bool) (projects []*model.Project, err error) {
  return mock.returns[0].([]*model.Project), mock.errors
}

func TestProjectsService_List(t *testing.T) {
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var projects = make([]*model.Project, 0)

    var r = &projectsRepositoryMockAPI{returns: []any{projects}}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).List(ctx, true)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    res, err = NewProjectsService(r, sentinelTechnologyTagsService).List(ctx, false)
    assert.NotNil(t, res)
    assert.NoError(t, err)

    res, err = NewProjectsService(r, sentinelTechnologyTagsService).List(ctx)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{returns: []any{([]*model.Project)(nil)}, errors: unexpected}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).List(ctx)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *projectsRepositoryMockAPI) Get(context.Context, string) (*model.Project, error) {
  return mock.returns[0].(*model.Project), mock.errors
}

func TestProjectsService_Get(t *testing.T) {
  var id = uuid.New().String()
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var project = new(model.Project)
    var r = &projectsRepositoryMockAPI{returns: []any{project}}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Get(ctx, id)
    assert.Equal(t, project, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{returns: []any{(*model.Project)(nil)}, errors: unexpected}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Get(ctx, id)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *projectsRepositoryMockAPI) GetBySlug(context.Context, string) (*model.Project, error) {
  return mock.returns[0].(*model.Project), mock.errors
}

func TestProjectsService_GetBySlug(t *testing.T) {
  var slug = "project-slug-name"
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var project = new(model.Project)
    var r = &projectsRepositoryMockAPI{returns: []any{project}}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).GetBySlug(ctx, slug)
    assert.Equal(t, project, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{returns: []any{(*model.Project)(nil)}, errors: unexpected}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).GetBySlug(ctx, slug)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *projectsRepositoryMockAPI) Create(_ context.Context, t *transfer.ProjectCreation) (string, error) {
  mock.called = true

  if nil != mock.t {
    assert.Equal(mock.t, mock.arguments[1], t)
  }

  return mock.returns[0].(string), mock.errors
}

func TestProjectsService_Create(t *testing.T) {
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var creation = transfer.ProjectCreation{
      Name:            "THIS Is The Project Name",
      Slug:            "this-is-the-project-name",
      Homepage:        "https://Homepage.com",
      Company:         "Foo",
      CompanyHomepage: "https://www.gotlim.com",
      Starts:          "2024-10-18",
      Ends:            "2024-10-18",
      Language:        "Language",
      Summary:         "Summary",
      ReadTime:        2,
      Content:         strings.TrimRight(strings.Repeat("word ", 200), " "),
      FirstImageURL:   "https://FirstImageURL.com",
      SecondImageURL:  "https://SecondImageURL.com",
      GitHubURL:       "https://GitHubURL.com",
      CollectionURL:   "https://CollectionURL.com",
    }
    var dirty = transfer.ProjectCreation{
      Name:            " \n\t THIS      Is\n\tThe \t Project    Name \n\t ",
      Homepage:        " \n\t " + creation.Homepage + " \n\t ",
      Company:         " \n\t " + creation.Company + " \n\t ",
      CompanyHomepage: " \n\t " + creation.CompanyHomepage + " \n\t ",
      Starts:          " \n\t " + creation.Starts + " \n\t ",
      Ends:            " \n\t " + creation.Ends + " \n\t ",
      Language:        " \n\t " + creation.Language + " \n\t ",
      Summary:         " \n\t " + creation.Summary + " \n\t ",
      Content:         " \n\t " + creation.Content + " \n\t ",
      FirstImageURL:   " \n\t " + creation.FirstImageURL + " \n\t ",
      SecondImageURL:  " \n\t " + creation.SecondImageURL + " \n\t ",
      GitHubURL:       " \n\t " + creation.GitHubURL + " \n\t ",
      CollectionURL:   " \n\t " + creation.CollectionURL + " \n\t ",
    }
    var r = &projectsRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, &creation},
      returns:   []any{id},
      errors:    nil,
    }

    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &dirty)
    assert.NoError(t, err)
    assert.Equal(t, id, res)
  })

  t.Run("success: content does not get created", func(t *testing.T) {
    var creation = transfer.ProjectCreation{
      Summary:  "Summary text",
      ReadTime: 1,
    }
    var dirty = transfer.ProjectCreation{
      Summary: " \n\t " + creation.Summary + " \n\t ",
    }
    var r = &projectsRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, &creation},
      returns:   []any{id},
    }

    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &dirty)
    assert.NoError(t, err)
    assert.Equal(t, id, res)
  })

  t.Run("no nil parameter", func(t *testing.T) {
    var r = &projectsRepositoryMockAPI{}
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, nil)
    require.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.Empty(t, res)
  })

  t.Run("creation validations", func(t *testing.T) {
    t.Run("len(creation.Name)<=36", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Name = strings.Repeat("x", 36)

      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Name = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Homepage)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Homepage = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Homepage = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Company)<=64", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Company = strings.Repeat("x", 64)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Company = strings.Repeat("x", 1+64)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.CompanyHomepage)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.CompanyHomepage = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.CompanyHomepage = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("creation.Starts has format YYYY-MM-DD and valid range", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Starts = "2024-10-18"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Starts = "2024/10/18"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)

      creation.Starts = "2024/10/81"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("creation.Ends has format YYYY-MM-DD and valid range", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Ends = "2024-10-18"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Ends = "2024/10/18"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)

      creation.Ends = "2024/10/81"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Language)<=64", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Language = strings.Repeat("x", 64)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Language = strings.Repeat("x", 1+64)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Summary)<=1024", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Summary = strings.Repeat("x", 1024)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Summary = strings.Repeat("x", 1+1024)

      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("wordsIn(creation.Summary)<=60", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Summary = strings.Repeat("word ", 60)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Summary = strings.Repeat("word ", 1+60)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Content)<=3MB", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Content = strings.Repeat("x", 3145728)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Content = strings.Repeat("x", 1+3145728)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.FirstImageURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.FirstImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.FirstImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.SecondImageURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.SecondImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.SecondImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.GitHubURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.GitHubURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.GitHubURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.CollectionURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.CollectionURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.CollectionURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, &creation},
        returns:   []any{id},
      }

      res, err = NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })
  })

  t.Run("expected error", func(t *testing.T) {
    var expected = problem.NewInternal()
    var r = &projectsRepositoryMockAPI{
      returns: []any{""},
      errors:  expected,
    }

    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, new(transfer.ProjectCreation))
    assert.ErrorAs(t, err, &expected)
    assert.Empty(t, res)
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{
      returns: []any{""},
      errors:  unexpected,
    }
    res, err := NewProjectsService(r, sentinelTechnologyTagsService).Create(ctx, new(transfer.ProjectCreation))
    assert.ErrorIs(t, err, unexpected)
    assert.Empty(t, res)
  })
}

func (mock *projectsRepositoryMockAPI) Exists(context.Context, string) error {
  return mock.errors
}

func TestProjectsService_Exists(t *testing.T) {
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var r = &projectsRepositoryMockAPI{errors: nil}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Exists(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{errors: unexpected}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Exists(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *projectsRepositoryMockAPI) Update(_ context.Context, id string, t *transfer.ProjectUpdate) error {
  mock.called = true

  if nil != mock.t {
    assert.Equal(mock.t, mock.arguments[1], id)
    assert.Equal(mock.t, mock.arguments[2], t)
  }

  return mock.errors
}

func TestProjectsService_Update(t *testing.T) {
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var update = transfer.ProjectUpdate{
      Name:            "THIS Is The new Project Name",
      Slug:            "this-is-the-new-project-name",
      Homepage:        "https://Homepage.com",
      Company:         "Foo",
      CompanyHomepage: "https://www.gotlim.com",
      Starts:          "2024-10-18",
      Ends:            "2024-10-19",
      Language:        "Language",
      Summary:         "Summary",
      ReadTime:        2,
      Content:         strings.TrimRight(strings.Repeat("word ", 300), " "),
      FirstImageURL:   "https://FirstImageURL.com",
      SecondImageURL:  "https://SecondImageURL.com",
      GitHubURL:       "https://GitHubURL.com",
      CollectionURL:   "https://CollectionURL.com",
      PlaygroundURL:   "https://PlaygroundURL.com",
      Archived:        false,
      Finished:        false,
    }
    var dirty = transfer.ProjectUpdate{
      Name:            " \n\t " + "THIS      Is\n\tThe   new \t Project    Name" + " \n\t ",
      Homepage:        " \n\t " + update.Homepage + " \n\t ",
      Company:         " \n\t " + update.Company + " \n\t ",
      CompanyHomepage: " \n\t " + update.CompanyHomepage + " \n\t ",
      Starts:          " \n\t " + update.Starts + " \n\t ",
      Ends:            " \n\t " + update.Ends + " \n\t ",
      Language:        " \n\t " + update.Language + " \n\t ",
      Summary:         " \n\t " + update.Summary + " \n\t ",
      Content:         " \n\t " + update.Content + " \n\t ",
      FirstImageURL:   " \n\t " + update.FirstImageURL + " \n\t ",
      SecondImageURL:  " \n\t " + update.SecondImageURL + " \n\t ",
      GitHubURL:       " \n\t " + update.GitHubURL + " \n\t ",
      CollectionURL:   " \n\t " + update.CollectionURL + " \n\t ",
      PlaygroundURL:   " \n\t " + update.PlaygroundURL + " \n\t ",
      Archived:        update.Archived,
      Finished:        update.Finished,
    }

    var r = &projectsRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id, &update},
      returns:   []any{true},
      errors:    nil,
    }

    err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &dirty)
    assert.NoError(t, err)
  })

  t.Run("success: content does not get updated", func(t *testing.T) {
    var update = transfer.ProjectUpdate{
      Summary:  "Summary text.",
      ReadTime: 1,
    }
    var dirty = transfer.ProjectUpdate{
      Summary: " \n\t " + update.Summary + " \n\t ",
    }
    var r = &projectsRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id, &update},
      returns:   []any{true},
      errors:    nil,
    }
    err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &dirty)
    assert.NoError(t, err)
  })

  t.Run("no nil parameter", func(t *testing.T) {
    var r = &projectsRepositoryMockAPI{}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, nil)
    require.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: update")
  })

  t.Run("update validations", func(t *testing.T) {
    t.Run("len(update.Name)<=36", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Name = strings.Repeat("x", 36)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Name = strings.Repeat("x", 1+36)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.Homepage)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Homepage = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Homepage = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(creation.Company)<=64", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Company = strings.Repeat("x", 64)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Company = strings.Repeat("x", 1+64)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(creation.CompanyHomepage)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.CompanyHomepage = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.CompanyHomepage = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("creation.Starts has format YYYY-MM-DD and valid range", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Starts = "2024-10-18"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Starts = "2024/10/18"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)

      update.Starts = "2024/10/81"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("creation.Ends has format YYYY-MM-DD and valid range", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Ends = "2024-10-18"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Ends = "2024/10/18"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)

      update.Ends = "2024/10/81"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        errors:    nil,
      }

      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.Language)<=64", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Language = strings.Repeat("x", 64)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Language = strings.Repeat("x", 1+64)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.Summary)<=1024", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Summary = strings.Repeat("x", 1024)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Summary = strings.Repeat("x", 1+1024)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("wordsIn(creation.Summary)<=60", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Summary = strings.Repeat("word ", 60)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Summary = strings.Repeat("word ", 1+60)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.Content)<=3MB", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Content = strings.Repeat("x", 3145728)
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.Content = strings.Repeat("x", 1+3145728)
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.FirstImageURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.FirstImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.FirstImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.SecondImageURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.SecondImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.SecondImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(creation.GitHubURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.GitHubURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.GitHubURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.CollectionURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.CollectionURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.CollectionURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })

    t.Run("len(update.PlaygroundURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.PlaygroundURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{true},
        errors:    nil,
      }
      err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.NoError(t, err)

      update.PlaygroundURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = &projectsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, &update},
        returns:   []any{false},
        errors:    nil,
      }
      err = NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, &update)
      assert.Error(t, err)
    })
  })

  t.Run("expected error", func(t *testing.T) {
    var expected = problem.NewInternal()
    var r = &projectsRepositoryMockAPI{returns: []any{false}, errors: expected}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, new(transfer.ProjectUpdate))
    assert.ErrorAs(t, err, &expected)
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{returns: []any{false}, errors: unexpected}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Update(ctx, id, new(transfer.ProjectUpdate))
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *projectsRepositoryMockAPI) Remove(context.Context, string) error {
  return mock.errors
}

func TestProjectsService_Remove(t *testing.T) {
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var r = &projectsRepositoryMockAPI{}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Remove(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &projectsRepositoryMockAPI{errors: unexpected}
    err := NewProjectsService(r, sentinelTechnologyTagsService).Remove(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}
