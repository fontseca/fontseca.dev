package handler

import (
  "fontseca/mocks"
  "fontseca/model"
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
