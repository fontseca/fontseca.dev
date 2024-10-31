package handler

import (
  "bytes"
  "context"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
  "time"
)

type meServiceMockAPI struct {
  meServiceAPI
  returns []any
  errors  error
  called  bool
}

func (mock *meServiceMockAPI) Get(context.Context) (*model.Me, error) {
  mock.called = true
  return mock.returns[0].(*model.Me), mock.errors
}

func (mock *meServiceMockAPI) Update(context.Context, *transfer.MeUpdate) error {
  mock.called = true
  return mock.errors
}

func TestMeHandler_Get(t *testing.T) {
  const method = http.MethodGet
  const target = "/me.get"

  t.Run("success", func(t *testing.T) {
    var me = model.Me{
      Username:     "Username",
      FirstName:    "FirstName",
      LastName:     "LastName",
      Summary:      "Summary",
      JobTitle:     "JobTitle",
      Email:        "Email",
      PhotoURL:     "PhotoURL",
      ResumeURL:    "ResumeURL",
      CodingSince:  2017,
      Company:      "Company",
      Location:     "Location",
      Hireable:     false,
      GitHubURL:    "GitHubURL",
      LinkedInURL:  "LinkedInURL",
      YouTubeURL:   "YouTubeURL",
      TwitterURL:   "TwitterURL",
      InstagramURL: "InstagramURL",
      CreatedAt:    time.Now(),
      UpdatedAt:    time.Now(),
    }
    var s = &meServiceMockAPI{returns: []any{&me}}
    var engine = gin.Default()
    engine.GET(target, NewMeHandler(s).Get)
    var recorder = httptest.NewRecorder()
    var request = httptest.NewRequest(method, target, nil)
    engine.ServeHTTP(recorder, request)
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, string(marshal(t, me)), recorder.Body.String())
  })
}

func TestMeHandler_SetPhoto(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.setPhoto"

  t.Run("success", func(t *testing.T) {
    var urls = []string{
      "https://picsum.photos/200/300",
      "http://www.picture.com/",
    }

    for _, url := range urls {
      url = strings.TrimSpace(url)
      var s = &meServiceMockAPI{returns: []any{true}}

      var request = httptest.NewRequest(method, target, nil)
      _ = request.ParseForm()
      request.PostForm.Set("photo_url", url)
      gin.SetMode(gin.ReleaseMode)
      var engine = gin.Default()
      engine.POST(target, NewMeHandler(s).SetPhoto)
      var recorder = httptest.NewRecorder()
      engine.ServeHTTP(recorder, request)

      assert.Equal(t, http.StatusNoContent, recorder.Code)
      assert.Empty(t, recorder.Body.String())
      assert.Empty(t, recorder.Header())
    }
  })
}

func (mock *meServiceMockAPI) SetHireable(context.Context, bool) error {
  mock.called = true
  return mock.errors
}

func TestMeHandler_SetResume(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.setResume"

  t.Run("success", func(t *testing.T) {
    var urls = []string{
      "https://picsum.photos/200/300",
      "http://www.picture.com/",
    }

    for _, url := range urls {
      url = strings.TrimSpace(url)
      var s = &meServiceMockAPI{returns: []any{true}}
      var request = httptest.NewRequest(method, target, nil)
      _ = request.ParseForm()
      request.PostForm.Set("resume_url", url)
      gin.SetMode(gin.ReleaseMode)
      var engine = gin.Default()
      engine.POST(target, NewMeHandler(s).SetResume)
      var recorder = httptest.NewRecorder()
      engine.ServeHTTP(recorder, request)

      assert.Equal(t, http.StatusNoContent, recorder.Code)
      assert.Empty(t, recorder.Body.String())
      assert.Empty(t, recorder.Header())
    }
  })
}

func TestMeHandler_SetHireable(t *testing.T) {
  const method = http.MethodPost
  const target = "/me.setHireable"

  t.Run("success", func(t *testing.T) {
    var s = &meServiceMockAPI{returns: []any{true}}
    var request = httptest.NewRequest(method, target, nil)
    _ = request.ParseForm()
    request.PostForm.Set("hireable", "true")
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.POST(target, NewMeHandler(s).SetHireable)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)

    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
    assert.Empty(t, recorder.Header())
  })

  t.Run("could not parse", func(t *testing.T) {
    var s = &meServiceMockAPI{}
    var request = httptest.NewRequest(method, target, nil)
    _ = request.ParseForm()
    request.PostForm.Set("hireable", "unparsable format")
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.POST(target, NewMeHandler(s).SetHireable)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)

    require.False(t, s.called)
    assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
    assert.NotEmpty(t, recorder.Body.String())
    assert.Contains(t, recorder.Body.String(), "Failure when parsing boolean value.")
  })
}

func TestMeHandler_Update(t *testing.T) {
  const (
    method = http.MethodPost
    target = "/me.update"
  )

  t.Run("success", func(t *testing.T) {
    var update = transfer.MeUpdate{
      Summary:      "Summary",
      JobTitle:     "JobTitle",
      Email:        "Email",
      Company:      "Company",
      Location:     "Location",
      GitHubURL:    "GitHubURL",
      LinkedInURL:  "LinkedInURL",
      YouTubeURL:   "YouTubeURL",
      TwitterURL:   "TwitterURL",
      InstagramURL: "InstagramURL",
    }
    var s = &meServiceMockAPI{returns: []any{true}}
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.POST(target, NewMeHandler(s).Set)
    var recorder = httptest.NewRecorder()
    var request = httptest.NewRequest(method, target, bytes.NewReader(marshal(t, update)))
    engine.ServeHTTP(recorder, request)

    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Empty(t, recorder.Body.String())
    assert.Equal(t, recorder.Header(), http.Header{"Content-Type": []string{"application/json"}})
  })
}
