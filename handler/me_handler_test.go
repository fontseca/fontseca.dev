package handler

import (
  "fontseca/mocks"
  "fontseca/model"
  "github.com/gin-gonic/gin"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "net/http"
  "net/http/httptest"
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
