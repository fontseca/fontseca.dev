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
  "testing"
)

type technologyTagServiceMockAPI struct {
  service.TechnologyTagService
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

func (mock *technologyTagServiceMockAPI) Get(context.Context) ([]*model.TechnologyTag, error) {
  mock.called = true
  return mock.returns[0].([]*model.TechnologyTag), mock.errors
}

func TestTechnologyTagHandler_Get(t *testing.T) {
  const method = http.MethodGet
  const target = "/technologies.list"

  t.Run("success", func(t *testing.T) {
    var technologies = []*model.TechnologyTag{
      new(model.TechnologyTag),
      new(model.TechnologyTag),
      new(model.TechnologyTag),
    }
    var s = &technologyTagServiceMockAPI{returns: []any{technologies}}
    var engine = gin.Default()
    engine.GET(target, NewTechnologyTagHandler(s).Get)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, technologies)), recorder.Body.String())
  })
}

func (mock *technologyTagServiceMockAPI) Add(_ context.Context, t *transfer.TechnologyTagCreation) (string, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], t)
  }

  return mock.returns[0].(string), mock.errors
}

func TestTechnologyTagHandler_Add(t *testing.T) {
  const method = http.MethodPost
  const target = "/technologies.add"
  var id = uuid.New().String()
  var creation = &transfer.TechnologyTagCreation{Name: "Name"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", creation.Name)

  t.Run("success", func(t *testing.T) {
    var s = &technologyTagServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), creation},
      returns:   []any{id},
    }
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
    var s = &technologyTagServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), creation},
      returns:   []any{""},
      errors:    expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &technologyTagServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), creation},
      returns:   []any{""},
      errors:    unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *technologyTagServiceMockAPI) Update(_ context.Context, id string, t *transfer.TechnologyTagUpdate) (bool, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
    require.Equal(mock.t, mock.arguments[2], t)
  }

  return mock.returns[0].(bool), mock.errors
}

func TestTechnologyTagHandler_Set(t *testing.T) {
  const method = http.MethodPost
  const target = "/technologies.set"
  var update = &transfer.TechnologyTagUpdate{Name: "Name"}
  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("name", update.Name)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = &technologyTagServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = &technologyTagServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{true},
    }
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("failed update without error", func(t *testing.T) {
    var s = &technologyTagServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id, update},
      returns:   []any{false},
    }
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
    var s = &technologyTagServiceMockAPI{
      returns: []any{false},
      errors:  expected,
    }
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &technologyTagServiceMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *technologyTagServiceMockAPI) Remove(_ context.Context, id string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
  }

  return mock.errors
}

func TestTechnologyTagHandler_Remove(t *testing.T) {
  const method = http.MethodPost
  const target = "/technologies.remove"
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = &technologyTagServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  var id = uuid.New().String()
  _ = request.ParseForm()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = &technologyTagServiceMockAPI{
      t:         t,
      arguments: []any{context.Background(), id},
      errors:    nil,
    }
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &technologyTagServiceMockAPI{errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &technologyTagServiceMockAPI{errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewTechnologyTagHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}
