package handler

import (
  "errors"
  "fontseca.dev/mocks"
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

func TestDraftsHandler_Start(t *testing.T) {
  const (
    routine = "Draft"
    method  = http.MethodPost
    target  = "/archive.drafts.start"
  )

  creation := &transfer.ArticleCreation{
    Title:   "Title",
    Content: "Content",
  }

  id := uuid.New()

  request := httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  request.PostForm.Add("title", creation.Title)
  request.PostForm.Add("content", creation.Content)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusCreated
    expectedResponse := string(marshal(t, gin.H{"draft_uuid": id.String()}))

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return(id, nil)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Start)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Equal(t, expectedResponse, recorder.Body.String())
    assert.Empty(t, recorder.Result().Cookies())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    expectedStatusCode := http.StatusBadRequest
    expectBodyContains := "Expected problem detail."

    expected := &problem.Problem{}
    expected.Status(expectedStatusCode)
    expected.Detail(expectBodyContains)

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return(uuid.Nil, expected)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Start)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })

  t.Run("expected problem detail", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    expectedStatusCode := http.StatusInternalServerError
    expectBodyContains := "An unexpected error occurred while processing your request"

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), creation).Return(uuid.Nil, unexpected)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Start)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
