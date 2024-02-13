package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "github.com/stretchr/testify/require"
  "testing"
  "time"
)

func TestMeService_Get(t *testing.T) {
  const routine = "Get"

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
      Hireable:     true,
      GitHubURL:    "GitHubURL",
      LinkedInURL:  "LinkedInURL",
      YouTubeURL:   "YouTubeURL",
      TwitterURL:   "TwitterURL",
      InstagramURL: "InstagramURL",
      CreatedAt:    time.Now(),
      UpdatedAt:    time.Now(),
    }
    var expected = me
    var r = mocks.NewMeRepository()
    var ctx = context.Background()
    r.On(routine, ctx).Return(&me, nil)
    res, err := NewMeService(r).Get(ctx)
    require.NotNil(t, res)
    assert.NoError(t, err)
    assert.Equal(t, expected, *res)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewMeRepository()
    var ctx = context.Background()
    r.On(routine, mock.Anything).Return(nil, unexpected)
    res, err := NewMeService(r).Get(ctx)
    assert.ErrorIs(t, err, unexpected)
    assert.Nil(t, res)
  })
}
