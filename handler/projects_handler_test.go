package handler

import (
  "fontseca/mocks"
  "fontseca/model"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "net/http"
  "net/http/httptest"
  "testing"
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
