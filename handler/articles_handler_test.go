package handler

import (
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "fontseca.dev/problem"
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

func TestArticlesHandler_GetHidden(t *testing.T) {
  const (
    routine = "GetHidden"
    method  = http.MethodGet
    target  = "/archive.articles.hidden.list"
  )

  request := httptest.NewRequest(method, target, nil)
  articles := []*model.Article{{}, {}, {}}

  t.Run("success without search", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, articles))

    s := mocks.NewArticlesService()
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(articles, nil)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(nil, expected)

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
    s.On(routine, mock.AnythingOfType("*gin.Context"), "").Return(nil, unexpected)

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).GetHidden)

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
    engine.GET(target, NewArticlesHandler(s).GetHidden)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.NotEmpty(t, recorder.Body.String())
    assert.Empty(t, recorder.Result().Cookies())
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
