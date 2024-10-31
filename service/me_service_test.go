package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "testing"
  "time"
)

type meRepositoryMockAPI struct {
  repository.MeRepository
  returns []any
  errors  error
  called  bool
}

func (mock *meRepositoryMockAPI) Get(context.Context) (*model.Me, error) {
  mock.called = true
  return mock.returns[0].(*model.Me), mock.errors
}

func TestMeService_Get(t *testing.T) {
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
    var r = &meRepositoryMockAPI{returns: []any{&me}}
    var ctx = context.Background()
    res, err := NewMeService(r).Get(ctx)
    require.NotNil(t, res)
    assert.NoError(t, err)
    assert.Equal(t, expected, *res)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &meRepositoryMockAPI{returns: []any{(*model.Me)(nil)}, errors: unexpected}
    var ctx = context.Background()
    res, err := NewMeService(r).Get(ctx)
    assert.ErrorIs(t, err, unexpected)
    assert.Nil(t, res)
  })
}

func (mock *meRepositoryMockAPI) Update(context.Context, *transfer.MeUpdate) error {
  mock.called = true
  return mock.errors
}

func TestMeService_Update(t *testing.T) {
  t.Run("success", func(t *testing.T) {
    var expected = transfer.MeUpdate{
      Summary:      "Summary",
      JobTitle:     "JobTitle",
      Email:        "Email@email.com",
      PhotoURL:     "https://www.PhotoURL.com",
      ResumeURL:    "https://www.ResumeURL.com",
      Company:      "Company",
      Location:     "Location",
      Hireable:     false,
      GitHubURL:    "https://www.GitHubURL.com/",
      LinkedInURL:  "https://www.LinkedInURL.com/",
      YouTubeURL:   "https://www.YouTubeURL.com/",
      TwitterURL:   "https://www.TwitterURL.com/",
      InstagramURL: "https://www.InstagramURL.com/",
    }
    var dirty = transfer.MeUpdate{
      Summary:      " \n\t " + expected.Summary + " \n\t ",
      JobTitle:     " \n\t " + expected.JobTitle + " \n\t ",
      Email:        " \n\t " + expected.Email + " \n\t ",
      PhotoURL:     " \n\t " + expected.PhotoURL + " \n\t ",
      ResumeURL:    " \n\t " + expected.ResumeURL + " \n\t ",
      Company:      " \n\t " + expected.Company + " \n\t ",
      Location:     " \n\t " + expected.Location + " \n\t ",
      Hireable:     expected.Hireable,
      GitHubURL:    " \n\t " + expected.GitHubURL + " \n\t ",
      LinkedInURL:  " \n\t " + expected.LinkedInURL + " \n\t ",
      YouTubeURL:   " \n\t " + expected.YouTubeURL + " \n\t ",
      TwitterURL:   " \n\t " + expected.TwitterURL + " \n\t ",
      InstagramURL: " \n\t " + expected.InstagramURL + " \n\t ",
    }
    var r = &meRepositoryMockAPI{returns: []any{true}}
    var ctx = context.Background()
    err := NewMeService(r).Update(ctx, &dirty)
    assert.NoError(t, err)
  })

  t.Run("error on nil update", func(t *testing.T) {
    var r = &meRepositoryMockAPI{}
    var ctx = context.Background()
    err := NewMeService(r).Update(ctx, nil)
    assert.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: update")
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &meRepositoryMockAPI{returns: []any{false}, errors: unexpected}
    var ctx = context.Background()
    err := NewMeService(r).Update(ctx, new(transfer.MeUpdate))
    assert.ErrorIs(t, err, unexpected)
  })
}
