package handler

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "net/http"
  "net/http/httptest"
  "testing"
)

type articlesServiceMockAPI struct {
  articlesServiceAPI
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
}

func (mock *articlesServiceMockAPI) List(_ context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], filter)
  }

  return mock.returns[0].([]*transfer.Article), mock.errors
}

func TestArticlesHandler_Get(t *testing.T) {
  const (
    method = http.MethodGet
    target = "/archive.articles.list"
  )

  request := httptest.NewRequest(method, target, nil)
  articles := []*transfer.Article{{}, {}, {}}

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, articles))
    filter := &transfer.ArticleFilter{Page: 1, RPP: 20}

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), filter}, returns: []any{articles}}

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).List)

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

    s := &articlesServiceMockAPI{returns: []any{[]*transfer.Article(nil)}, errors: expected}

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).List)

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

    s := &articlesServiceMockAPI{returns: []any{[]*transfer.Article(nil)}, errors: unexpected}

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).List)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *articlesServiceMockAPI) ListHidden(_ context.Context, filter *transfer.ArticleFilter) (articles []*transfer.Article, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], filter)
  }

  return mock.returns[0].([]*transfer.Article), mock.errors
}

func TestArticlesHandler_GetHidden(t *testing.T) {
  const (
    method = http.MethodGet
    target = "/archive.articles.hidden.list"
  )

  request := httptest.NewRequest(method, target, nil)
  articles := []*transfer.Article{{}, {}, {}}

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, articles))
    filter := &transfer.ArticleFilter{Page: 1, RPP: 20}

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), filter}, returns: []any{articles}}

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).ListHidden)

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

    s := &articlesServiceMockAPI{returns: []any{[]*transfer.Article(nil)}, errors: expected}

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).ListHidden)

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

    s := &articlesServiceMockAPI{returns: []any{[]*transfer.Article(nil)}, errors: unexpected}

    engine := gin.Default()
    engine.GET(target, NewArticlesHandler(s).ListHidden)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *articlesServiceMockAPI) GetByID(_ context.Context, articleUUID string) (article *model.Article, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleUUID)
  }

  return mock.returns[0].(*model.Article), mock.errors
}

func TestArticlesHandler_GetByID(t *testing.T) {
  const (
    method = http.MethodGet
    target = "/archive.articles.get"
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

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}, returns: []any{article}}

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

    s := &articlesServiceMockAPI{returns: []any{(*model.Article)(nil)}, errors: expected}

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

    s := &articlesServiceMockAPI{returns: []any{(*model.Article)(nil)}, errors: unexpected}

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

func (mock *articlesServiceMockAPI) Hide(_ context.Context, articleID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesHandler_Hide(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.hide"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) Show(_ context.Context, articleID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesHandler_Show(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.show"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) Amend(_ context.Context, articleID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesHandler_Amend(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.amend"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) Remove(_ context.Context, articleID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesHandler_Remove(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.remove"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) Pin(_ context.Context, articleID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesHandler_Pin(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.pin"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) Unpin(_ context.Context, articleID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesHandler_Unpin(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.unpin"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) AddTag(_ context.Context, articleUUID, tagID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleUUID)
    require.Equal(mock.t, mock.arguments[2], tagID)
  }

  return mock.errors
}

func TestArticlesHandler_AddTag(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.tags.add"
  )

  request := httptest.NewRequest(method, target, nil)
  articleUUID := uuid.NewString()
  tagID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", articleUUID)
  request.PostForm.Add("tag_id", tagID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), articleUUID, tagID}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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

func (mock *articlesServiceMockAPI) RemoveTag(_ context.Context, articleUUID, tagID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleUUID)
    require.Equal(mock.t, mock.arguments[2], tagID)
  }

  return mock.errors
}

func TestArticlesHandler_RemoveTag(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.tags.remove"
  )

  request := httptest.NewRequest(method, target, nil)
  articleUUID := uuid.NewString()
  tagID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("article_uuid", articleUUID)
  request.PostForm.Add("tag_id", tagID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &articlesServiceMockAPI{t: t, arguments: []any{context.Background(), articleUUID, tagID}}

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

    s := &articlesServiceMockAPI{errors: expected}

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

    s := &articlesServiceMockAPI{errors: unexpected}

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
