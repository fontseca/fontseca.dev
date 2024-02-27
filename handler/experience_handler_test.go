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
