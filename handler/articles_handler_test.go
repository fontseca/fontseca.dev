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

func TestArticlesHandler_Get(t *testing.T) {
  const (
    routine = "Get"
    method  = http.MethodGet
    target  = "/archive.articles.list"
  )

  request := httptest.NewRequest(method, target, nil)
  articles := []*transfer.Article{{}, {}, {}}

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, articles))

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*transfer.ArticleFilter")).Return(articles, nil)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*transfer.ArticleFilter")).Return(nil, expected)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*transfer.ArticleFilter")).Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_GetHidden(t *testing.T) {
  const (
    routine = "GetHidden"
    method  = http.MethodGet
    target  = "/archive.articles.hidden.list"
  )

  request := httptest.NewRequest(method, target, nil)
  articles := []*transfer.Article{{}, {}, {}}

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, articles))

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*transfer.ArticleFilter")).Return(articles, nil)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetHidden)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*transfer.ArticleFilter")).Return(nil, expected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetHidden)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*transfer.ArticleFilter")).Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetHidden)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_GetByID(t *testing.T) {
  const (
    routine = "GetByID"
    method  = http.MethodGet
    target  = "/archive.articles.info"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  article := &model.Article{
    UUID: uuid.MustParse(id),
  }

  request.URL.RawQuery = request.URL.RawQuery + "&article_uuid=" + id

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, article))

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(article, nil)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetByID)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil, expected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetByID)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetByID)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_Hide(t *testing.T) {
  const (
    routine = "Hide"
    method  = http.MethodPost
    target  = "/archive.articles.hide"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Hide)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Hide)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Hide)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_Show(t *testing.T) {
  const (
    routine = "Show"
    method  = http.MethodPost
    target  = "/archive.articles.show"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Show)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Show)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Show)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_Amend(t *testing.T) {
  const (
    routine = "Amend"
    method  = http.MethodPost
    target  = "/archive.articles.amend"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Amend)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Amend)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Amend)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_Remove(t *testing.T) {
  const (
    routine = "Remove"
    method  = http.MethodPost
    target  = "/archive.articles.remove"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Remove)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Remove)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Remove)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_Pin(t *testing.T) {
  const (
    routine = "Pin"
    method  = http.MethodPost
    target  = "/archive.articles.pin"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Pin)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Pin)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Pin)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
func TestArticlesHandler_Unpin(t *testing.T) {
  const (
    routine = "Unpin"
    method  = http.MethodPost
    target  = "/archive.articles.unpin"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Unpin)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Unpin)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), id).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).Unpin)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_AddTag(t *testing.T) {
  const (
    routine = "AddTag"
    method  = http.MethodPost
    target  = "/archive.articles.tags.add"
  )

  request := httptest.NewRequest(method, target, nil)
  articleUUID := uuid.NewString()
  tagID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", articleUUID)
  request.PostForm.Add("tag_id", tagID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), articleUUID, tagID).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).AddTag)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), articleUUID, tagID).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).AddTag)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), articleUUID, tagID).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).AddTag)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func TestArticlesHandler_RemoveTag(t *testing.T) {
  const (
    routine = "RemoveTag"
    method  = http.MethodPost
    target  = "/archive.articles.tags.remove"
  )

  request := httptest.NewRequest(method, target, nil)
  articlesUUID := uuid.NewString()
  tagID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", articlesUUID)
  request.PostForm.Add("tag_id", tagID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), articlesUUID, tagID).Return(nil)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).RemoveTag)

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

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), articlesUUID, tagID).Return(expected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).RemoveTag)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), articlesUUID, tagID).Return(unexpected)

    engine := gin.Default()
    engine.POST(target, NewArticlesHandler(s).RemoveTag)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
