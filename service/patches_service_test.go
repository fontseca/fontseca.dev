package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "github.com/stretchr/testify/assert"
  "testing"
)

func TestPatchesService_Get(t *testing.T) {
  const routine = "GetPatches"

  ctx := context.TODO()

  t.Run("success", func(t *testing.T) {
    expectedPatches := make([]*model.ArticlePatch, 3)

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx).Return(expectedPatches, nil)

    articles, err := NewPatchesService(r).Get(ctx)

    assert.Equal(t, expectedPatches, articles)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx).Return(nil, unexpected)

    articles, err := NewPatchesService(r).Get(ctx)

    assert.Nil(t, articles)
    assert.ErrorIs(t, err, unexpected)
  })
}
