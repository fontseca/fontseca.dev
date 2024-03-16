package handler

import (
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "fontseca/problem"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "net/http"
  "net/http/httptest"
  "net/url"
  "testing"
  "time"
)

func TestProjectsHandler_Get(t *testing.T) {
  const routine = "Get"
  const method = http.MethodGet
  const target = "/me.projects.list"

  t.Run("success", func(t *testing.T) {
    var projects = make([]*model.Project, 0)
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), []bool(nil)).Return(projects, nil)
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
  const routine = "Get"
  const method = http.MethodGet
  const target = "/me.projects.hidden.list"

  t.Run("success", func(t *testing.T) {
    var projects = make([]*model.Project, 0)
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), []bool{true}).Return(projects, nil)
    var engine = gin.Default()
    engine.GET(target, NewProjectsHandler(s).GetArchived)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, projects)), recorder.Body.String())
  })
}

func TestProjectsHandler_GetByID(t *testing.T) {
  const routine = "GetByID"
  const method = http.MethodGet
  const target = "/me.projects.info"

  t.Run("success", func(t *testing.T) {
    var language = "Go"
    var estimatedTime = 1
    var project = &model.Project{
      ID:             uuid.New(),
      Name:           "Name",
      Homepage:       "https://Homepage.com",
      Language:       &language,
      Summary:        "Summary.",
      Content:        "Content.",
      EstimatedTime:  &estimatedTime,
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
    var id = project.ID.String()
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(project, nil)
    var engine = gin.Default()
    engine.GET(target, NewProjectsHandler(s).GetByID)
    var request = httptest.NewRequest(method, target, nil)
    var query = url.Values{}
    query.Add("id", id)
    request.URL.RawQuery = query.Encode()
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, project)), recorder.Body.String())
  })
}

func TestProjectsHandler_Add(t *testing.T) {
  const routine = "Add"
  const method = http.MethodPost
  const target = "/me.projects.add"
  var id = uuid.New().String()
  var creation = &transfer.ProjectCreation{
    Name:           "Name",
    Homepage:       "https://Homepage.com",
    Language:       "Go",
    Summary:        "Summary.",
    Content:        "Content.",
    EstimatedTime:  1,
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return(id, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return("", expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return("", unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Set(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.set"
  var update = &transfer.ProjectUpdate{
    Name:          "Name",
    Homepage:      "https://Homepage.com",
    Language:      "Go",
    Summary:       "Summary.",
    Content:       "Content.",
    EstimatedTime: 1,
  }
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", update.Name)
  request.PostForm.Add("homepage", update.Homepage)
  request.PostForm.Add("language", update.Language)
  request.PostForm.Add("summary", update.Summary)
  request.PostForm.Add("content", update.Content)
  request.PostForm.Add("estimated_time", "1")

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Archive(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.archive"
  var update = &transfer.ProjectUpdate{Archived: true}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Archive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Unarchive(t *testing.T) {
  const routine = "Unarchive"
  const method = http.MethodPost
  const target = "/me.projects.unarchive"
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(false, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unarchive)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Finish(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.finish"
  var update = &transfer.ProjectUpdate{Finished: true}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Finish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Unfinish(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.unfinish"
  var update = &transfer.ProjectUpdate{Finished: false}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Unfinish)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetPlaygroundURL(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.setPlaygroundURL"
  var update = &transfer.ProjectUpdate{PlaygroundURL: "https://PlaygroundURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)
  request.PostForm.Add("url", update.PlaygroundURL)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetPlaygroundURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetFirstImageURL(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.setFirstImageURL"
  var update = &transfer.ProjectUpdate{FirstImageURL: "https://FirstImageURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)
  request.PostForm.Add("url", update.FirstImageURL)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetFirstImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetSecondImageURL(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.setSecondImageURL"
  var update = &transfer.ProjectUpdate{SecondImageURL: "https://SecondImageURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)
  request.PostForm.Add("url", update.SecondImageURL)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetSecondImageURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetGitHubURL(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.setGitHubURL"
  var update = &transfer.ProjectUpdate{GitHubURL: "https://GitHubURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)
  request.PostForm.Add("url", update.GitHubURL)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetGitHubURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_SetCollectionURL(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.projects.setCollectionURL"
  var update = &transfer.ProjectUpdate{CollectionURL: "https://CollectionURL.com"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)
  request.PostForm.Add("url", update.CollectionURL)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("failed update without error: conflicts with current resource state", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).SetCollectionURL)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestProjectsHandler_Remove(t *testing.T) {
  const routine = "Remove"
  const method = http.MethodPost
  const target = "/me.projects.remove"
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)
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
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything).Return(expected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewProjectsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything).Return(unexpected)
    var engine = gin.Default()
    engine.POST(target, NewProjectsHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}
