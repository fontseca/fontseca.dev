package handler

import (
  "fontseca/mocks"
  "fontseca/model"
  "fontseca/transfer"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
  "time"
)

func TestMeHandler_Get(t *testing.T) {
  const routine = "Get"
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
    var s = mocks.NewMeService()
    s.On(routine, mock.AnythingOfType("*gin.Context")).Return(&me, nil)
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
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.setPhoto"

  t.Run("success", func(t *testing.T) {
    var urls = []string{
      "  \t\n\t  ",
      "https://picsum.photos/200/300",
      "http://www.picture.com/",
    }

    for _, url := range urls {
      url = strings.TrimSpace(url)
      var expected = transfer.MeUpdate{PhotoURL: url}

      var s = mocks.NewMeService()
      s.On(routine, mock.AnythingOfType("*gin.Context"), &expected).Return(true, nil)

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

func TestMeHandler_SetResume(t *testing.T) {
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.setResume"

  t.Run("success", func(t *testing.T) {
    var urls = []string{
      "  \t\n\t  ",
      "https://picsum.photos/200/300",
      "http://www.picture.com/",
    }

    for _, url := range urls {
      url = strings.TrimSpace(url)
      var expected = transfer.MeUpdate{ResumeURL: url}

      var s = mocks.NewMeService()
      s.On(routine, mock.AnythingOfType("*gin.Context"), &expected).Return(true, nil)

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
  const routine = "Update"
  const method = http.MethodPost
  const target = "/me.setHireable"

  t.Run("success", func(t *testing.T) {
    var s = mocks.NewMeService()
    var expected = transfer.MeUpdate{Hireable: true}
    s.On(routine, mock.AnythingOfType("*gin.Context"), &expected).Return(true, nil)

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
    var s = mocks.NewMeService()
    s.AssertNotCalled(t, routine)

    var request = httptest.NewRequest(method, target, nil)
    _ = request.ParseForm()
    request.PostForm.Set("hireable", "unparsable format")
    gin.SetMode(gin.ReleaseMode)
    var engine = gin.Default()
    engine.POST(target, NewMeHandler(s).SetHireable)
    var recorder = httptest.NewRecorder()
    engine.ServeHTTP(recorder, request)

    assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
    assert.NotEmpty(t, recorder.Body.String())
    assert.Contains(t, recorder.Body.String(), "Failure when parsing boolean value.")
  })
}
