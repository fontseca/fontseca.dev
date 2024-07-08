package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "strings"
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

func TestProjectService_GetBySlug(t *testing.T) {
  const routine = "GetBySlug"
  var slug = "project-slug-name"
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var project = new(model.Project)
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, slug).Return(project, nil)
    res, err := NewProjectsService(r).GetBySlug(ctx, slug)
    assert.Equal(t, project, res)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, slug).Return(nil, unexpected)
    res, err := NewProjectsService(r).GetBySlug(ctx, slug)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestProjectsService_Add(t *testing.T) {
  const routine = "Add"
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var creation = transfer.ProjectCreation{
      Name:           "THIS Is The Project Name",
      Slug:           "this-is-the-project-name",
      Homepage:       "https://Homepage.com",
      Language:       "Language",
      Summary:        "Summary",
      ReadTime:       2,
      Content:        strings.TrimRight(strings.Repeat("word ", 200), " "),
      FirstImageURL:  "https://FirstImageURL.com",
      SecondImageURL: "https://SecondImageURL.com",
      GitHubURL:      "https://GitHubURL.com",
      CollectionURL:  "https://CollectionURL.com",
    }
    var dirty = transfer.ProjectCreation{
      Name:           " \n\t THIS      Is\n\tThe \t Project    Name \n\t ",
      Homepage:       " \n\t " + creation.Homepage + " \n\t ",
      Language:       " \n\t " + creation.Language + " \n\t ",
      Summary:        " \n\t " + creation.Summary + " \n\t ",
      Content:        " \n\t " + creation.Content + " \n\t ",
      FirstImageURL:  " \n\t " + creation.FirstImageURL + " \n\t ",
      SecondImageURL: " \n\t " + creation.SecondImageURL + " \n\t ",
      GitHubURL:      " \n\t " + creation.GitHubURL + " \n\t ",
      CollectionURL:  " \n\t " + creation.CollectionURL + " \n\t ",
    }
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, &creation).Return(id, nil)
    res, err := NewProjectsService(r).Add(ctx, &dirty)
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
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, &creation).Return(id, nil)
    res, err := NewProjectsService(r).Add(ctx, &dirty)
    assert.NoError(t, err)
    assert.Equal(t, id, res)
  })

  t.Run("no nil parameter", func(t *testing.T) {
    var r = mocks.NewProjectsRepository()
    r.AssertNotCalled(t, routine)
    res, err := NewProjectsService(r).Add(ctx, nil)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.Empty(t, res)
  })

  t.Run("creation validations", func(t *testing.T) {
    t.Run("len(creation.Name)<=36", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Name = strings.Repeat("x", 36)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Name = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Homepage)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Homepage = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Homepage = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Language)<=64", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Language = strings.Repeat("x", 64)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Language = strings.Repeat("x", 1+64)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Summary)<=1024", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Summary = strings.Repeat("x", 1024)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Summary = strings.Repeat("x", 1+1024)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("wordsIn(creation.Summary)<=60", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Summary = strings.Repeat("word ", 60)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Summary = strings.Repeat("word ", 1+60)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.Content)<=3MB", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.Content = strings.Repeat("x", 3145728)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.Content = strings.Repeat("x", 1+3145728)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.FirstImageURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.FirstImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.FirstImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.SecondImageURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.SecondImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.SecondImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.GitHubURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.GitHubURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.GitHubURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })

    t.Run("len(creation.CollectionURL)<=2048", func(t *testing.T) {
      var creation = transfer.ProjectCreation{}

      creation.CollectionURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err := NewProjectsService(r).Add(ctx, &creation)
      assert.NoError(t, err)
      assert.Equal(t, id, res)

      creation.CollectionURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, &creation).Return(id, nil)
      res, err = NewProjectsService(r).Add(ctx, &creation)
      assert.Error(t, err)
      assert.Empty(t, res)
    })
  })

  t.Run("expected error", func(t *testing.T) {
    var expected = problem.NewInternal()
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, mock.Anything).Return("", expected)
    res, err := NewProjectsService(r).Add(ctx, new(transfer.ProjectCreation))
    assert.ErrorAs(t, err, &expected)
    assert.Empty(t, res)
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, mock.Anything).Return("", unexpected)
    res, err := NewProjectsService(r).Add(ctx, new(transfer.ProjectCreation))
    assert.ErrorIs(t, err, unexpected)
    assert.Empty(t, res)
  })
}

func TestProjectService_Exists(t *testing.T) {
  const routine = "Exists"
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id).Return(nil)
    err := NewProjectsService(r).Exists(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id).Return(unexpected)
    err := NewProjectsService(r).Exists(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestProjectsService_Update(t *testing.T) {
  const routine = "Update"
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var update = transfer.ProjectUpdate{
      Name:           "THIS Is The new Project Name",
      Slug:           "this-is-the-new-project-name",
      Homepage:       "https://Homepage.com",
      Language:       "Language",
      Summary:        "Summary",
      ReadTime:       2,
      Content:        strings.TrimRight(strings.Repeat("word ", 300), " "),
      FirstImageURL:  "https://FirstImageURL.com",
      SecondImageURL: "https://SecondImageURL.com",
      GitHubURL:      "https://GitHubURL.com",
      CollectionURL:  "https://CollectionURL.com",
      PlaygroundURL:  "https://PlaygroundURL.com",
      Archived:       false,
      Finished:       false,
    }
    var dirty = transfer.ProjectUpdate{
      Name:           " \n\t " + "THIS      Is\n\tThe   new \t Project    Name" + " \n\t ",
      Homepage:       " \n\t " + update.Homepage + " \n\t ",
      Language:       " \n\t " + update.Language + " \n\t ",
      Summary:        " \n\t " + update.Summary + " \n\t ",
      Content:        " \n\t " + update.Content + " \n\t ",
      FirstImageURL:  " \n\t " + update.FirstImageURL + " \n\t ",
      SecondImageURL: " \n\t " + update.SecondImageURL + " \n\t ",
      GitHubURL:      " \n\t " + update.GitHubURL + " \n\t ",
      CollectionURL:  " \n\t " + update.CollectionURL + " \n\t ",
      PlaygroundURL:  " \n\t " + update.PlaygroundURL + " \n\t ",
      Archived:       update.Archived,
      Finished:       update.Finished,
    }
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id, &update).Return(true, nil)
    res, err := NewProjectsService(r).Update(ctx, id, &dirty)
    assert.NoError(t, err)
    assert.True(t, res)
  })

  t.Run("success: content does not get updated", func(t *testing.T) {
    var update = transfer.ProjectUpdate{
      Summary:  "Summary text.",
      ReadTime: 1,
    }
    var dirty = transfer.ProjectUpdate{
      Summary: " \n\t " + update.Summary + " \n\t ",
    }
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id, &update).Return(true, nil)
    res, err := NewProjectsService(r).Update(ctx, id, &dirty)
    assert.NoError(t, err)
    assert.True(t, res)
  })

  t.Run("no nil parameter", func(t *testing.T) {
    var r = mocks.NewProjectsRepository()
    r.AssertNotCalled(t, routine)
    res, err := NewProjectsService(r).Update(ctx, id, nil)
    assert.ErrorContains(t, err, "nil value for parameter: update")
    assert.Empty(t, res)
  })

  t.Run("update validations", func(t *testing.T) {
    t.Run("len(update.Name)<=36", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Name = strings.Repeat("x", 36)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.Name = strings.Repeat("x", 1+36)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.Homepage)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Homepage = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.Homepage = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.Language)<=64", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Language = strings.Repeat("x", 64)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.Language = strings.Repeat("x", 1+64)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.Summary)<=1024", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Summary = strings.Repeat("x", 1024)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.Summary = strings.Repeat("x", 1+1024)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("wordsIn(creation.Summary)<=60", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Summary = strings.Repeat("word ", 60)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.Summary = strings.Repeat("word ", 1+60)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.Content)<=3MB", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.Content = strings.Repeat("x", 3145728)
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.Content = strings.Repeat("x", 1+3145728)
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.FirstImageURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.FirstImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.FirstImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.SecondImageURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.SecondImageURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.SecondImageURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(creation.GitHubURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.GitHubURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.GitHubURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.CollectionURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.CollectionURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.CollectionURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })

    t.Run("len(update.PlaygroundURL)<=2048", func(t *testing.T) {
      var update = transfer.ProjectUpdate{}

      update.PlaygroundURL = "https://" + strings.Repeat("x", 2036) + ".com"
      var r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(true, nil)
      res, err := NewProjectsService(r).Update(ctx, id, &update)
      assert.NoError(t, err)
      assert.True(t, res)

      update.PlaygroundURL = "https://" + strings.Repeat("x", 1+2036) + ".com"
      r = mocks.NewProjectsRepository()
      r.On(routine, ctx, id, &update).Return(false, nil)
      res, err = NewProjectsService(r).Update(ctx, id, &update)
      assert.Error(t, err)
      assert.False(t, res)
    })
  })

  t.Run("expected error", func(t *testing.T) {
    var expected = problem.NewInternal()
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything).Return(false, expected)
    res, err := NewProjectsService(r).Update(ctx, id, new(transfer.ProjectUpdate))
    assert.ErrorAs(t, err, &expected)
    assert.Empty(t, res)
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything).Return(false, unexpected)
    res, err := NewProjectsService(r).Update(ctx, id, new(transfer.ProjectUpdate))
    assert.ErrorIs(t, err, unexpected)
    assert.Empty(t, res)
  })
}

func TestProjectService_Remove(t *testing.T) {
  const routine = "Remove"
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id).Return(nil)
    err := NewProjectsService(r).Remove(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewProjectsRepository()
    r.On(routine, ctx, id).Return(unexpected)
    err := NewProjectsService(r).Remove(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}
