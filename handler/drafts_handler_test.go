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

  t.Run("unexpected error", func(t *testing.T) {
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

func TestDraftsHandler_Publish(t *testing.T) {
  const (
    routine = "Publish"
    method  = http.MethodPost
    target  = "/archive.drafts.publish"
  )

  id := uuid.NewString()

  request := httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Publish)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Empty(t, recorder.Body.String())
    assert.Empty(t, recorder.Result().Cookies())
  })

  t.Run("expected problem detail", func(t *testing.T) {
    expectedStatusCode := http.StatusBadRequest
    expectBodyContains := "Expected problem detail."

    expected := &problem.Problem{}
    expected.Status(expectedStatusCode)
    expected.Detail(expectBodyContains)

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Publish)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Publish)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestDraftsHandler_Get(t *testing.T) {
  const (
    routine = "Get"
    method  = http.MethodGet
    target  = "/archive.drafts.list"
  )

  request := httptest.NewRequest(method, target, nil)
  drafts := []*model.Article{{}, {}, {}}

  t.Run("success without search", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, drafts))

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(drafts, nil)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).Get)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(nil, expected)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).Get)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).Get)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "needle").Return(drafts, nil)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.NotEmpty(t, recorder.Body.String())
    assert.Empty(t, recorder.Result().Cookies())
  })
}

func TestDraftsHandler_GetByID(t *testing.T) {
  const (
    routine = "GetByID"
    method  = http.MethodGet
    target  = "/archive.drafts.info"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  draft := &model.Article{
    UUID: uuid.MustParse(id),
  }

  request.URL.RawQuery = request.URL.RawQuery + "&draft_uuid=" + id

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, draft))

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(draft, nil)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).GetByID)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil, expected)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).GetByID)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).GetByID)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestDraftsHandler_AddTopic(t *testing.T) {
  const (
    routine = "AddTopic"
    method  = http.MethodPost
    target  = "/archive.drafts.topics.add"
  )

  request := httptest.NewRequest(method, target, nil)
  draftUUID := uuid.NewString()
  topicUUID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", draftUUID)
  request.PostForm.Add("topic_uuid", topicUUID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), draftUUID, topicUUID).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).AddTopic)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), draftUUID, topicUUID).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).AddTopic)

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

    s := mocks.NewDraftsService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), draftUUID, topicUUID).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).AddTopic)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
