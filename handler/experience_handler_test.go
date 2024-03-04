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

func TestExperienceHandler_Get(t *testing.T) {
  const routine = "Get"
  const method = http.MethodGet
  const target = "/experience.list"

  t.Run("success", func(t *testing.T) {
    var e = make([]*model.Experience, 1)
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), []bool(nil)).Return(e, nil)
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.GET(target, NewExperienceHandler(s).Get)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, e)), recorder.Body.String())
  })
}

func TestExperienceHandler_GetHidden(t *testing.T) {
  const routine = "Get"
  const method = http.MethodGet
  const target = "/experience.hidden.list"

  t.Run("success", func(t *testing.T) {
    var e = make([]*model.Experience, 1)
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), []bool{true}).Return(e, nil)
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.GET(target, NewExperienceHandler(s).GetHidden)
    var request = httptest.NewRequest(method, target, nil)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, e)), recorder.Body.String())
  })
}

func TestExperienceHandler_GetByID(t *testing.T) {
  const routine = "GetByID"
  const method = http.MethodGet
  const target = "/experience.info"

  t.Run("success", func(t *testing.T) {
    var i = 2023
    var e = &model.Experience{
      ID:        uuid.New(),
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
    var id = e.ID.String()
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(e, nil)
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.GET(target, NewExperienceHandler(s).GetByID)
    var request = httptest.NewRequest(method, target, nil)
    var query = url.Values{}
    query.Add("id", id)
    request.URL.RawQuery = query.Encode()
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, e)), recorder.Body.String())
  })
}

func TestExperienceHandler_Add(t *testing.T) {
  const routine = "Save"
  const method = http.MethodPost
  const target = "/experience.add"

  var e = &transfer.ExperienceCreation{
    Starts:   2028,
    Ends:     2030,
    JobTitle: "JobTitle",
    Company:  "Company",
    Country:  "Country",
    Summary:  "Summary",
  }

  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()
  request.PostForm.Add("starts", "2028")
  request.PostForm.Add("ends", "2030")
  request.PostForm.Add("job_title", "JobTitle")
  request.PostForm.Add("company", "Company")
  request.PostForm.Add("country", "Country")
  request.PostForm.Add("summary", "Summary")

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), e).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusCreated, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    var expected = &problem.Problem{}
    expected.Status(http.StatusGone)
    expected.Detail("Expected problem detail.")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), e).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), e).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Add)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Set(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/experience.set"

  var update = &transfer.ExperienceUpdate{
    Starts:   2028,
    Ends:     2030,
    JobTitle: "JobTitle",
    Company:  "Company",
    Country:  "Country",
    Summary:  "Summary",
  }

  var request = httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  var id = uuid.New().String()

  request.PostForm.Add("starts", "2028")
  request.PostForm.Add("ends", "2030")
  request.PostForm.Add("job_title", "JobTitle")
  request.PostForm.Add("company", "Company")
  request.PostForm.Add("country", "Country")
  request.PostForm.Add("summary", "Summary")

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is missing or not provided in the form data.")
  })

  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
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
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Set)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Hide(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/experience.hide"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("id", id)
  var update = &transfer.ExperienceUpdate{Hidden: true}

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
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
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Hide)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Show(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/experience.show"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("id", id)
  var update = &transfer.ExperienceUpdate{Hidden: false}

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
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
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Show)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Quit(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/experience.quit"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("id", id)
  var update = &transfer.ExperienceUpdate{Active: false, Ends: time.Now().Year()}

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(true, nil)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
  })

  t.Run("redirects when there's nothing new", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, nil)
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
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, expected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, update).Return(false, unexpected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Quit)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}

func TestExperienceHandler_Remove(t *testing.T) {
  const routine = "Remove"
  const method = http.MethodPost
  const target = "/experience.remove"

  var id = uuid.New().String()
  var request = httptest.NewRequest(method, target, nil)

  t.Run("missing 'id' parameter", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.AssertNotCalled(t, routine)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusBadRequest, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "The 'id' parameter is required but was not found in the request form data.")
  })

  _ = request.ParseForm()
  request.PostForm.Add("id", id)

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)
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
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusGone, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Expected problem detail.")
  })

  t.Run("unexpected error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var s = mocks.NewExperienceService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)
    var engine = gin.Default()
    engine.POST(target, NewExperienceHandler(s).Remove)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusInternalServerError, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "An unexpected error occurred while processing your request")
  })
}
