package handler

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
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

type experienceServiceMockAPI struct {
  experienceServiceAPI
  returns []any
  errors  error
  called  bool
}

func (mock *experienceServiceMockAPI) List(context.Context, ...bool) (experience []*model.Experience, err error) {
  return mock.returns[0].([]*model.Experience), mock.errors
}

func TestExperienceHandler_Get(t *testing.T) {
  const method = http.MethodGet
  const target = "/experience.list"

  t.Run("success", func(t *testing.T) {
    var e = make([]*model.Experience, 1)
    var s = &experienceServiceMockAPI{returns: []any{e}}
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.GET(target, NewExperienceHandler(s).List)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, e)), recorder.Body.String())
  })
}

func TestExperienceHandler_GetHidden(t *testing.T) {
  const method = http.MethodGet
  const target = "/experience.hidden.list"

  t.Run("success", func(t *testing.T) {
    var e = make([]*model.Experience, 1)
    var s = &experienceServiceMockAPI{returns: []any{e}}
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.GET(target, NewExperienceHandler(s).ListHidden)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, e)), recorder.Body.String())
  })
}

func (mock *experienceServiceMockAPI) Get(context.Context, string) (*model.Experience, error) {
  return mock.returns[0].(*model.Experience), mock.errors
}

func TestExperienceHandler_GetByID(t *testing.T) {
  const method = http.MethodGet
  const target = "/experience.info"

  t.Run("success", func(t *testing.T) {
    var i = 2023
    var e = &model.Experience{
      UUID:      uuid.New(),
      Starts:    2020,
      Ends:      &i,
      JobTitle:  "JobTitle",
      Company:   "Company",
      Country:   "Country",
      Summary:   "Summary",
      Active:    false,
      Hidden:    false,
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
    }
    var id = e.UUID.String()
    var s = &experienceServiceMockAPI{returns: []any{e}}
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.GET(target, NewExperienceHandler(s).Get)
    var request = httptest.NewRequest(method, target, nil)
    var query = url.Values{}
    query.Add("experience_uuid", id)
    request.URL.RawQuery = query.Encode()
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, e)), recorder.Body.String())
  })
}

func (mock *experienceServiceMockAPI) Create(context.Context, *transfer.ExperienceCreation) (bool, error) {
  return mock.returns[0].(bool), mock.errors
}

func TestExperienceHandler_Add(t *testing.T) {
  const method = http.MethodPost
  const target = "/experience.add"

  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("starts", "2028")
  request.PostForm.Add("ends", "2030")
  request.PostForm.Add("job_title", "JobTitle")
  request.PostForm.Add("company", "Company")
  request.PostForm.Add("country", "Country")
  request.PostForm.Add("summary", "Summary")

  t.Run("success", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{true}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Create)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusCreated, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Create)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Create)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *experienceServiceMockAPI) Update(context.Context, string, *transfer.ExperienceUpdate) (bool, error) {
  return mock.returns[0].(bool), mock.errors
}

func TestExperienceHandler_Set(t *testing.T) {
  const method = http.MethodPost
  const target = "/experience.set"

  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  var id = uuid.New().String()

  request.PostForm.Add("starts", "2028")
  request.PostForm.Add("ends", "2030")
  request.PostForm.Add("job_title", "JobTitle")
  request.PostForm.Add("company", "Company")
  request.PostForm.Add("country", "Country")
  request.PostForm.Add("summary", "Summary")

  t.Run("missing 'experience_uuid' parameter", func(t *testing.T) {
    var s = &experienceServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'experience_uuid' parameter is required but was not found in the request form data.")
  })

  request.PostForm.Add("experience_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{true}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{false}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Hide(t *testing.T) {
  const method = http.MethodPost
  const target = "/experience.hide"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'experience_uuid' parameter", func(t *testing.T) {
    var s = &experienceServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'experience_uuid' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("experience_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{true}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{false}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Show(t *testing.T) {
  const method = http.MethodPost
  const target = "/experience.show"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'experience_uuid' parameter", func(t *testing.T) {
    var s = &experienceServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'experience_uuid' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("experience_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{true}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{false}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Quit(t *testing.T) {
  const method = http.MethodPost
  const target = "/experience.quit"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'experience_uuid' parameter", func(t *testing.T) {
    var s = &experienceServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'experience_uuid' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("experience_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{true}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = &experienceServiceMockAPI{returns: []any{false}}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusSeeOther, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &experienceServiceMockAPI{returns: []any{false}, errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func (mock *experienceServiceMockAPI) Remove(_ context.Context, _ string) error {
  return mock.errors
}

func TestExperienceHandler_Remove(t *testing.T) {
  const method = http.MethodPost
  const target = "/experience.remove"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'experience_uuid' parameter", func(t *testing.T) {
    var s = &experienceServiceMockAPI{}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    require.False(t, s.called)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'experience_uuid' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("experience_uuid", id)

  t.Run("success", func(t *testing.T) {
    var s = &experienceServiceMockAPI{errors: nil}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = &experienceServiceMockAPI{errors: expected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = &experienceServiceMockAPI{errors: unexpected}
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}
