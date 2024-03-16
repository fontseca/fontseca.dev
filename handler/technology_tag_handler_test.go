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
  "testing"
)

func TestTechnologyTagHandler_Get(t *testing.T) {
  const routine = "Get"
  const method = http.MethodGet
  const target = "/technologies.list"

  t.Run("success", func(t *testing.T) {
    var technologies = []*model.TechnologyTag{
      new(model.TechnologyTag),
      new(model.TechnologyTag),
      new(model.TechnologyTag),
    }
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context")).Return(technologies, nil)
    var engine = gin.Default()
    engine.GET(target, NewTechnologyTagHandler(s).Get)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, technologies)), recorder.Body.String())
  })
}

func TestTechnologyTagHandler_Add(t *testing.T) {
  const routine = "Add"
  const method = http.MethodPost
  const target = "/technologies.add"
  var id = uuid.New().String()
  var creation = &transfer.TechnologyTagCreation{Name: "Name"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", creation.Name)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return(id, nil)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, gin.H{"inserted_id": id})), recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return("", expected)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return("", unexpected)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestTechnologyTagHandler_Set(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/technologies.set"
  var update = &transfer.TechnologyTagUpdate{Name: "Name"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", update.Name)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewTechnologyTagService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusConflict, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewTechnologyTagService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.Anything, mock.Anything).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}
