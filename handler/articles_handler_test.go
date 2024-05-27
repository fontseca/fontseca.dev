package handler

import (
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "net/http"
  "net/http/httptest"
  "testing"
)

func TestArticlesHandler_Get(t *testing.T) {
  const (
    routine = "Get"
    method  = http.MethodGet
    target  = "/archive.articles.list"
  )

  request := httptest.NewRequest(method, target, nil)
  articles := []*model.Article{{}, {}, {}}

  t.Run("success without search", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, articles))

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(articles, nil)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Equal(t, expectedBody, recorder.Body.String())
    assert.Empty(t, recorder.Result().Cookies())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    expectedStatusCode := http.StatusBadRequest
    expectBodyContains := "Expected problem detail."

    expected := &problem.Problem{}
    expected.Status(expectedStatusCode)
    expected.Detail(expectBodyContains)

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(nil, expected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })

  t.Run("unexpected error", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    expectedStatusCode := http.StatusInternalServerError
    expectBodyContains := "An unexpected error occurred while processing your request"

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })

  t.Run("success with search", func(t *testing.T) {
    request.URL.RawQuery = request.URL.RawQuery + "&search=needle"

    expectedStatusCode := http.StatusOK

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "needle").Return(articles, nil)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.NotEmpty(t, recorder.Body.String())
    assert.Empty(t, recorder.Result().Cookies())
  })
}
