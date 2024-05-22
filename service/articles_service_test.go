package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "testing"
)

func TestArticlesService_Get(t *testing.T) {
  const routine = "Get"

  ctx := context.TODO()

  t.Run("success", func(t *testing.T) {
    expectedDrafts := make([]*model.Article, 3)

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, "", false, false).Return(expectedDrafts, nil)

    drafts, err := NewArticlesService(r).Get(ctx, "\n \t \n")

    assert.Equal(t, expectedDrafts, drafts)
    assert.NoError(t, err)
  })

  t.Run("success with search", func(t *testing.T) {
    expectedDrafts := make([]*model.Article, 3)
    expectedNeedle := "20 www xxx yyy zzz zzz"

    needle := ">> = 20 www? xxx! yyy... zzz_zzz \" ' ° <<"

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, expectedNeedle, false, false).Return(expectedDrafts, nil)

    drafts, err := NewArticlesService(r).Get(ctx, needle)

    assert.Equal(t, expectedDrafts, drafts)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, unexpected)

    _, err := NewArticlesService(r).Get(ctx, "")
    assert.ErrorIs(t, err, unexpected)
  })
}
