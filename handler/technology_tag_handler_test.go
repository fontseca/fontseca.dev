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
