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

type draftsServiceMockAPI struct {
  draftsServiceAPI
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
}

func (mock *draftsServiceMockAPI) Draft(_ context.Context, creation *transfer.ArticleCreation) (insertedUUID uuid.UUID, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], creation)
  }

  return mock.returns[0].(uuid.UUID), mock.errors
}

func TestDraftsHandler_Start(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.start"
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

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), creation}, returns: []any{id}}

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

    s := &draftsServiceMockAPI{returns: []any{uuid.Nil}, errors: expected}

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

    s := &draftsServiceMockAPI{returns: []any{uuid.Nil}, errors: unexpected}

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

func (mock *draftsServiceMockAPI) Publish(_ context.Context, draftUUID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
  }

  return mock.errors
}

func TestDraftsHandler_Publish(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.publish"
  )

  id := uuid.NewString()

  request := httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

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

    s := &draftsServiceMockAPI{errors: expected}

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

    s := &draftsServiceMockAPI{errors: unexpected}

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

func (mock *draftsServiceMockAPI) Get(_ context.Context, filter *transfer.ArticleFilter) (drafts []*transfer.Article, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], filter)
  }

  return mock.returns[0].([]*transfer.Article), mock.errors
}

func TestDraftsHandler_Get(t *testing.T) {
  const (
    method = http.MethodGet
    target = "/archive.drafts.list"
  )

  request := httptest.NewRequest(method, target, nil)
  drafts := []*transfer.Article{{}, {}, {}}

  t.Run("success without search", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, drafts))
    filter := &transfer.ArticleFilter{Page: 1, RPP: 20}

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), filter}, returns: []any{drafts}}

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

    s := &draftsServiceMockAPI{returns: []any{[]*transfer.Article(nil)}, errors: expected}

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

    s := &draftsServiceMockAPI{returns: []any{[]*transfer.Article(nil)}, errors: unexpected}

    engine := gin.Default()
    engine.GET(target, NewDraftsHandler(s).Get)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *draftsServiceMockAPI) GetByID(_ context.Context, draftUUID string) (draft *model.Article, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
  }

  return mock.returns[0].(*model.Article), mock.errors
}

func TestDraftsHandler_GetByID(t *testing.T) {
  const (
    method = http.MethodGet
    target = "/archive.drafts.info"
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

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), id}, returns: []any{draft}}

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

    s := &draftsServiceMockAPI{returns: []any{(*model.Article)(nil)}, errors: expected}

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

    s := &draftsServiceMockAPI{returns: []any{(*model.Article)(nil)}, errors: unexpected}

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

func (mock *draftsServiceMockAPI) AddTag(_ context.Context, draftUUID, tagID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
    require.Equal(mock.t, mock.arguments[2], tagID)
  }

  return mock.errors
}

func TestDraftsHandler_AddTag(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.tags.add"
  )

  request := httptest.NewRequest(method, target, nil)
  draftUUID := uuid.NewString()
  tagID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", draftUUID)
  request.PostForm.Add("tag_id", tagID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), draftUUID, tagID}}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).AddTag)

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

    s := &draftsServiceMockAPI{errors: expected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).AddTag)

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

    s := &draftsServiceMockAPI{errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).AddTag)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *draftsServiceMockAPI) RemoveTag(_ context.Context, draftUUID, tagID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
    require.Equal(mock.t, mock.arguments[2], tagID)
  }

  return mock.errors
}

func TestDraftsHandler_RemoveTag(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.tags.remove"
  )

  request := httptest.NewRequest(method, target, nil)
  draftUUID := uuid.NewString()
  tagID := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", draftUUID)
  request.PostForm.Add("tag_id", tagID)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), draftUUID, tagID}}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).RemoveTag)

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

    s := &draftsServiceMockAPI{errors: expected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).RemoveTag)

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

    s := &draftsServiceMockAPI{errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).RemoveTag)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *draftsServiceMockAPI) Share(_ context.Context, draftUUID string) (link string, err error) {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
  }

  return mock.returns[0].(string), mock.errors
}

func TestDraftsHandler_Share(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.share"
    link   = "/link/to/draft"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusOK
    expectedBody := string(marshal(t, gin.H{"shareable_link": link}))

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), id}, returns: []any{link}}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Share)

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

    s := &draftsServiceMockAPI{returns: []any{"about:blank"}, errors: expected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Share)

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

    s := &draftsServiceMockAPI{returns: []any{"about:blank"}, errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Share)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *draftsServiceMockAPI) Discard(_ context.Context, draftUUID string) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
  }

  return mock.errors
}

func TestDraftsHandler_Discard(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.discard"
  )

  request := httptest.NewRequest(method, target, nil)
  id := uuid.NewString()

  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", id)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), id}}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Discard)

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

    s := &draftsServiceMockAPI{errors: expected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Discard)

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

    s := &draftsServiceMockAPI{errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Discard)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}

func (mock *draftsServiceMockAPI) Revise(_ context.Context, draftUUID string, revision *transfer.ArticleRevision) error {
  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftUUID)
    require.Equal(mock.t, mock.arguments[2], revision)
  }

  return mock.errors
}

func TestDraftsHandler_Revise(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/archive.drafts.revise"
  )

  revision := &transfer.ArticleRevision{
    Title:   "Title",
    Content: "Content",
  }

  id := uuid.NewString()

  request := httptest.NewRequest(method, target, nil)
  _ = request.ParseForm()

  request.PostForm.Add("draft_uuid", id)
  request.PostForm.Add("title", revision.Title)
  request.PostForm.Add("content", revision.Content)

  t.Run("success", func(t *testing.T) {
    expectedStatusCode := http.StatusNoContent

    s := &draftsServiceMockAPI{t: t, arguments: []any{context.Background(), id, revision}}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Revise)

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

    s := &draftsServiceMockAPI{errors: expected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Revise)

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

    s := &draftsServiceMockAPI{errors: unexpected}

    engine := gin.Default()
    engine.POST(target, NewDraftsHandler(s).Revise)

    recorder := httptest.NewRecorder()

    engine.ServeHTTP(recorder, request)

    assert.Equal(t, expectedStatusCode, recorder.Code)
    assert.Contains(t, recorder.Body.String(), expectBodyContains)
    assert.Empty(t, recorder.Result().Cookies())
    assert.Contains(t, recorder.Result().Header.Get("Content-Type"), "application/problem+json")
  })
}
