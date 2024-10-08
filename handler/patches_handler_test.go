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

type patchesServiceMockAPI struct {
  patchesServiceAPI
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
}

func (mock *patchesServiceMockAPI) List(context.Context) (patches []*model.ArticlePatch, err error) {
  return mock.returns[0].([]*model.ArticlePatch), mock.errors
}

func TestPatchesHandler_Get(t *testing.T) {
  const (
    method = http.MethodGet
    target = "/archive.articles.patches.list"
  )

  request := httptest.NewRequest(method, target, nil)
  patches := []*model.ArticlePatch{{}, {}, {}}

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, patches))

    s := &patchesServiceMockAPI{returns: []any{patches}}

    engine := gin.Default()
    engine.GET(target, NewPatchesHandler(s).List)

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

    s := &patchesServiceMockAPI{returns: []any{[]*model.ArticlePatch(nil)}, errors: expected}

    engine := gin.Default()
    engine.GET(target, NewPatchesHandler(s).List)

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

    s := &patchesServiceMockAPI{returns: []any{[]*model.ArticlePatch(nil)}, errors: unexpected}

    engine := gin.Default()
    engine.GET(target, NewPatchesHandler(s).List)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *patchesServiceMockAPI) Revise(_ context.Context, patchID string, revision *transfer.ArticleRevision) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
    require.Equal(mock.t, mock.arguments[2], revision)
  }

  return mock.errors
}

func TestPatchesHandler_Revise(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.revise"
  )

  revision := &transfer.ArticleRevision{
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

    s := &patchesServiceMockAPI{t: t, arguments: []any{context.Background(), id, revision}}

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

    s := &patchesServiceMockAPI{errors: expected}

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

    s := &patchesServiceMockAPI{errors: unexpected}

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

func (mock *patchesServiceMockAPI) Share(_ context.Context, patchID string) (link string, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
  }

  return mock.returns[0].(string), mock.errors
}

func TestPatchesHandler_Share(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.patches.share"
    link   = "/link/to/draft"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("patch_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, gin.H{"shareable_link": link}))

    s := &patchesServiceMockAPI{t: t, arguments: []any{context.Background(), id}, returns: []any{link}}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Share)

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

    s := &patchesServiceMockAPI{returns: []any{"about:blank"}, errors: expected}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Share)

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

    s := &patchesServiceMockAPI{returns: []any{"about:blank"}, errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Share)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *patchesServiceMockAPI) Discard(_ context.Context, patchID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
  }

  return mock.errors
}

func TestPatchesHandler_Discard(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.patches.discard"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("patch_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &patchesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Discard)

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

    s := &patchesServiceMockAPI{errors: expected}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Discard)

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

    s := &patchesServiceMockAPI{errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Discard)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *patchesServiceMockAPI) Release(_ context.Context, patchID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
  }

  return mock.errors
}

func TestPatchesHandler_Release(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.articles.patches.release"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("patch_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &patchesServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Release)

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

    s := &patchesServiceMockAPI{errors: expected}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Release)

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

    s := &patchesServiceMockAPI{errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewPatchesHandler(s).Release)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
