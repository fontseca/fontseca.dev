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
    var e = &model.Experience{
      ID:        uuid.New(),
      Starts:    2020,
      Ends:      2030,
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
