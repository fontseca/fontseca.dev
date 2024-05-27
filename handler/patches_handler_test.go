package handler

import (
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "net/http"
  "net/http/httptest"
  "testing"
)

func TestPatchesHandler_Get(t *testing.T) {
  const (
    routine = "Get"
    method  = http.MethodGet
    target  = "/archive.articles.patches.list"
  )

  request := httptest.NewRequest(method, target, nil)
  patches := []*model.ArticlePatch{{}, {}, {}}

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, patches))

    s := mocks.NewPatchesService()
    s.On(routine, mock.AnythingOfType("*gin.Context")).Return(patches, nil)

    engine := gin.Default()
    engine.GET(target, NewPatchesHandler(s).Get)

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

    s := mocks.NewPatchesService()
    s.On(routine, mock.AnythingOfType("*gin.Context")).Return(nil, expected)

    engine := gin.Default()
    engine.GET(target, NewPatchesHandler(s).Get)

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

    s := mocks.NewPatchesService()
    s.On(routine, mock.AnythingOfType("*gin.Context")).Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewPatchesHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestPatchesHandler_Revise(t *testing.T) {
  const (
    routine = "Revise"
    method  = http.MethodPost
    target  = "/archive.articles.revise"
  )

  revision := &transfer.ArticleUpdate{
    Title:   "Title",
    Content: "Content",
  }

  id := uuid.NewString()

  request := httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  request.PostForm.Add("patch_uuid", id)
  request.PostForm.Add("title", revision.Title)
  request.PostForm.Add("content", revision.Content)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewPatchesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, revision).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Revise)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Empty(t, recorder.Body)
    assert.Empty(t, recorder.Result().Cookies())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    expectedStatusCode := http.StatusBadRequest
    expectBodyContains := "Expected problem detail."

    expected := &problem.Problem{}
    expected.Status(expectedStatusCode)
    expected.Detail(expectBodyContains)

    s := mocks.NewPatchesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, revision).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Revise)

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

    s := mocks.NewPatchesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id, revision).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Revise)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
