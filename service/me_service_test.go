package service

import (
  "context"
  "errors"
  "fontseca/mocks"
  "fontseca/model"
  "fontseca/transfer"
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

func TestMeService_Update(t *testing.T) {
  const routine = "Update"

  t.Run("success", func(t *testing.T) {
    var expected = transfer.MeUpdate{
      Summary:      "Summary",
      JobTitle:     "JobTitle",
      Email:        "Email",
      Company:      "Company",
      Location:     "Location",
      Hireable:     false,
      GitHubURL:    "GitHubURL",
      LinkedInURL:  "LinkedInURL",
      YouTubeURL:   "YouTubeURL",
      TwitterURL:   "TwitterURL",
      InstagramURL: "InstagramURL",
    }
    var dirty = transfer.MeUpdate{
      Summary:      " \n\t " + expected.Summary + " \n\t ",
      JobTitle:     " \n\t " + expected.JobTitle + " \n\t ",
      Email:        " \n\t " + expected.Email + " \n\t ",
      Company:      " \n\t " + expected.Company + " \n\t ",
      Location:     " \n\t " + expected.Location + " \n\t ",
      Hireable:     expected.Hireable,
      GitHubURL:    " \n\t " + expected.GitHubURL + " \n\t ",
      LinkedInURL:  " \n\t " + expected.LinkedInURL + " \n\t ",
      YouTubeURL:   " \n\t " + expected.YouTubeURL + " \n\t ",
      TwitterURL:   " \n\t " + expected.TwitterURL + " \n\t ",
      InstagramURL: " \n\t " + expected.InstagramURL + " \n\t ",
    }
    var r = mocks.NewMeRepository()
    var ctx = context.Background()
    r.On(routine, ctx, &expected).Return(true, nil)
    res, err := NewMeService(r).Update(ctx, &dirty)
    assert.NoError(t, err)
    assert.True(t, res)
  })

  t.Run("error on nil update", func(t *testing.T) {
    var r = mocks.NewMeRepository()
    var ctx = context.Background()
    r.AssertNotCalled(t, routine)
    res, err := NewMeService(r).Update(ctx, nil)
    assert.ErrorContains(t, err, "nil value for parameter: update")
    assert.False(t, res)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = mocks.NewMeRepository()
    var ctx = context.Background()
    r.On(routine, mock.Anything, mock.Anything).Return(false, unexpected)
    res, err := NewMeService(r).Update(ctx, new(transfer.MeUpdate))
    assert.ErrorIs(t, err, unexpected)
    assert.False(t, res)
  })
}
