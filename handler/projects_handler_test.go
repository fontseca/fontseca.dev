package handler

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/service"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "net/http"
  "net/http/httptest"
  "net/url"
  "testing"
  "time"
)

type projectsServiceMockAPI struct {
  service.ProjectsService
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

func (mock *projectsServiceMockAPI) Get(_ context.Context, b ...bool) ([]*model.Project, error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], b)
  }

  return mock.returns[0].([]*model.Project), mock.errors
}

func TestProjectsHandler_Get(t *testing.T) {
  const method = http.MethodGet
  const target = "/me.projects.list"

  t.Run("success", func(t *testing.T) {
    var projects = make([]*model.Project, 0)
    var s = &projectsServiceMockAPI{returns: []any{projects}}
    var engine = gin.Default()
    engine.GET(target, NewProjectsHandler(s).Get)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, projects)), recorder.Body.String())
  })
}

func TestProjectsHandler_GetArchived(t *testing.T) {
  const method = http.MethodGet
  const target = "/me.projects.hidden.list"

  t.Run("success", func(t *testing.T) {
    var projects = make([]*model.Project, 0)
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), []bool{true}},
      returns:   []any{projects},
    }
    var engine = gin.Default()
    engine.GET(target, NewProjectsHandler(s).GetArchived)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, projects)), recorder.Body.String())
  })
}

func (mock *projectsServiceMockAPI) GetByID(context.Context, string) (*model.Project, error) {
  return mock.returns[0].(*model.Project), mock.errors
}

func TestProjectsHandler_GetByID(t *testing.T) {
  const method = http.MethodGet
  const target = "/me.projects.info"

  t.Run("success", func(t *testing.T) {
    var language = "Go"
    var project = &model.Project{
      UUID:           uuid.New(),
      Name:           "Name",
      Homepage:       "https://Homepage.com",
      Language:       &language,
      Summary:        "Summary.",
      Content:        "Content.",
      FirstImageURL:  "https://FirstImageURL.com",
      SecondImageURL: "https://SecondImageURL.com",
      GitHubURL:      "https://GitHubURL.com",
      CollectionURL:  "https://CollectionURL.com",
      PlaygroundURL:  "https://PlaygroundURL.com",
      Playable:       true,
      Archived:       false,
      Finished:       false,
      TechnologyTags: nil,
      CreatedAt:      time.Now(),
      UpdatedAt:      time.Now(),
    }
    var id = project.UUID.String()
    var s = &projectsServiceMockAPI{returns: []any{project}}
    var engine = gin.Default()
    engine.GET(target, NewProjectsHandler(s).GetByID)
    var request = httptest.NewRequest(method, target, nil)
    var query = url.Values{}
    query.Add("project_uuid", id)
    request.URL.RawQuery = query.Encode()
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, project)), recorder.Body.String())
  })
}

func (mock *projectsServiceMockAPI) Add(_ context.Context, t *transfer.ProjectCreation) (string, error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], t)
  }

  return mock.returns[0].(string), mock.errors
}

func TestProjectsHandler_Add(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.add"
  var id = uuid.New().String()
  var creation = &transfer.ProjectCreation{
    Name:           "Name",
    Homepage:       "https://Homepage.com",
    Language:       "Go",
    Summary:        "Summary.",
    Content:        "Content.",
    FirstImageURL:  "https://FirstImageURL.com",
    SecondImageURL: "https://SecondImageURL.com",
    GitHubURL:      "https://GitHubURL.com",
    CollectionURL:  "https://CollectionURL.com",
  }
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", creation.Name)
  request.PostForm.Add("homepage", creation.Homepage)
  request.PostForm.Add("language", creation.Language)
  request.PostForm.Add("summary", creation.Summary)
  request.PostForm.Add("content", creation.Content)
  request.PostForm.Add("estimated_time", "1")
  request.PostForm.Add("first_image_url", creation.FirstImageURL)
  request.PostForm.Add("second_image_url", creation.SecondImageURL)
  request.PostForm.Add("github_url", creation.GitHubURL)
  request.PostForm.Add("collection_url", creation.CollectionURL)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), creation},
      returns:   []any{id},
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, gin.H{"inserted_id": id})), recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), creation},
      returns:   []any{""},
      errors:    expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), creation},
      returns:   []any{""},
      errors:    unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *projectsServiceMockAPI) Update(_ context.Context, id string, t *transfer.ProjectUpdate) (bool, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
    require.Equal(mock.t, mock.arguments[2], t)
  }

  return mock.returns[0].(bool), mock.errors
}

func TestProjectsHandler_Set(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.set"
  var update = &transfer.ProjectUpdate{
    Name:     "Name",
    Homepage: "https://Homepage.com",
    Language: "Go",
    Summary:  "Summary.",
    Content:  "Content.",
  }
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", update.Name)
  request.PostForm.Add("homepage", update.Homepage)
  request.PostForm.Add("language", update.Language)
  request.PostForm.Add("summary", update.Summary)
  request.PostForm.Add("content", update.Content)
  request.PostForm.Add("estimated_time", "1")

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Archive(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.archive"
  var update = &transfer.ProjectUpdate{Archived: true}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *projectsServiceMockAPI) Unarchive(_ context.Context, id string) (bool, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
  }

  return mock.returns[0].(bool), mock.errors
}

func TestProjectsHandler_Unarchive(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.unarchive"
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id},
      returns:   []any{true},
      errors:    nil,
    }

    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Finish(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.finish"
  var update = &transfer.ProjectUpdate{Finished: true}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Unfinish(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.unfinish"
  var update = &transfer.ProjectUpdate{Finished: false}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetPlaygroundURL(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.setPlaygroundURL"
  var update = &transfer.ProjectUpdate{PlaygroundURL: "https://PlaygroundURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)
  request.PostForm.Add("url", update.PlaygroundURL)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetFirstImageURL(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.setFirstImageURL"
  var update = &transfer.ProjectUpdate{FirstImageURL: "https://FirstImageURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)
  request.PostForm.Add("url", update.FirstImageURL)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetSecondImageURL(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.setSecondImageURL"
  var update = &transfer.ProjectUpdate{SecondImageURL: "https://SecondImageURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)
  request.PostForm.Add("url", update.SecondImageURL)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetGitHubURL(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.setGitHubURL"
  var update = &transfer.ProjectUpdate{GitHubURL: "https://GitHubURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)
  request.PostForm.Add("url", update.GitHubURL)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetCollectionURL(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.setCollectionURL"
  var update = &transfer.ProjectUpdate{CollectionURL: "https://CollectionURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)
  request.PostForm.Add("url", update.CollectionURL)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *projectsServiceMockAPI) Remove(_ context.Context, id string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
  }

  return mock.errors
}

func TestProjectsHandler_Remove(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.remove"
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), id},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), id},
      errors:    expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), id},
      errors:    unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *projectsServiceMockAPI) AddTechnologyTag(_ context.Context, id1 string, id2 string) (bool, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id1)
    require.Equal(mock.t, mock.arguments[2], id2)
  }

  return mock.returns[0].(bool), mock.errors
}

func TestProjectsHandler_AddTechnologyTag(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.technologies.add"
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).AddTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("missing 'technology_id' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).AddTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'technology_id' parameter is required but was not found in the request form data.")
  })

  var techID = uuid.New().String()
  request.PostForm.Add("technology_id", techID)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), id, techID},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).AddTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).AddTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).AddTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *projectsServiceMockAPI) RemoveTechnologyTag(_ context.Context, id1 string, id2 string) (bool, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id1)
    require.Equal(mock.t, mock.arguments[2], id2)
  }

  return mock.returns[0].(bool), mock.errors
}

func TestProjectsHandler_RemoveTechnologyTag(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.projects.technologies.remove"
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'project_uuid' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).RemoveTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'project_uuid' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("project_uuid", id)

  t.Run("missing 'technology_id' parameter", func(t *testing.T) {
    var s = &projectsServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).RemoveTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'technology_id' parameter is required but was not found in the request form data.")
  })

  var techID = uuid.New().String()
  request.PostForm.Add("technology_id", techID)

  t.Run("success", func(t *testing.T) {
    var s = &projectsServiceMockAPI{
      arguments: []any{context.Background(), id, techID},
      returns:   []any{true},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).RemoveTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).RemoveTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &projectsServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).RemoveTechnologyTag)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}
