package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "strings"
  "testing"
)

type archiveRepositoryMockAPIForPatches struct {
  archiveRepositoryAPIForPatches
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

func (mock *archiveRepositoryMockAPIForPatches) ListPatches(context.Context) ([]*model.ArticlePatch, error) {
  return mock.returns[0].([]*model.ArticlePatch), mock.errors
}

func TestPatchesService_Get(t *testing.T) {
  ctx := context.TODO()

  t.Run("success", func(t *testing.T) {
    expectedPatches := make([]*model.ArticlePatch, 3)

    r := &archiveRepositoryMockAPIForPatches{returns: []any{expectedPatches}}

    articles, err := NewPatchesService(r).List(ctx)

    assert.Equal(t, expectedPatches, articles)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForPatches{returns: []any{[]*model.ArticlePatch(nil)}, errors: unexpected}

    articles, err := NewPatchesService(r).List(ctx)

    assert.Nil(t, articles)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForPatches) Revise(_ context.Context, patchID string, revision *transfer.ArticleRevision) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
    require.Equal(mock.t, mock.arguments[2], revision)
  }

  return mock.errors
}

func TestPatchesService_Revise(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    revision := &transfer.ArticleRevision{
      Title:    "Consectetur! Adipiscing... Quis nostrud: ELIT?",
      Slug:     "consectetur-adipiscing-quis-nostrud-elit",
      ReadTime: 11,
      Content:  strings.Repeat("word ", 1999) + "word",
    }

    dirty := &transfer.ArticleRevision{
      Title:   " \t\n " + revision.Title + " \t\n ",
      Content: " \t\n " + revision.Content + " \t\n ",
    }

    r := &archiveRepositoryMockAPIForPatches{t: t, arguments: []any{ctx, id, revision}}
    assert.NoError(t, NewPatchesService(r).Revise(ctx, id, dirty))
  })

  t.Run("success: changing title", func(t *testing.T) {
    revision := &transfer.ArticleRevision{
      Title:    "Consectetur-Adipiscing!!... Quis nostrud: ELIT??? +-'\"",
      Slug:     "consectetur-adipiscing-quis-nostrud-elit",
      ReadTime: 1,
    }

    dirty := &transfer.ArticleRevision{
      Title: " \t\n " + revision.Title + " \t\n ",
    }

    r := &archiveRepositoryMockAPIForPatches{t: t, arguments: []any{ctx, id, revision}}
    assert.NoError(t, NewPatchesService(r).Revise(ctx, id, dirty))
  })

  t.Run("success: changing content", func(t *testing.T) {
    revision := &transfer.ArticleRevision{
      Content:  strings.Repeat("word ", 299) + "word",
      ReadTime: 2,
    }

    dirty := &transfer.ArticleRevision{
      Content: " \t\n " + revision.Content + " \t\n ",
    }

    r := &archiveRepositoryMockAPIForPatches{t: t, arguments: []any{ctx, id, revision}}
    assert.NoError(t, NewPatchesService(r).Revise(ctx, id, dirty))
  })

  t.Run("nil parameter: revision", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForPatches{}
    assert.ErrorContains(t, NewPatchesService(r).Revise(ctx, id, nil), "nil value")
    assert.False(t, r.called)
  })

  t.Run("wrong uuid: id", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForPatches{}
    assert.Error(t, NewPatchesService(r).Revise(ctx, "x", &transfer.ArticleRevision{}))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForPatches{errors: unexpected}
    assert.ErrorIs(t, NewPatchesService(r).Revise(ctx, id, &transfer.ArticleRevision{}), unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForPatches) Share(_ context.Context, patchID string) (string, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
  }

  return mock.returns[0].(string), mock.errors
}

func TestPatchesService_Share(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    expectedLink := "link-to-resource"

    r := &archiveRepositoryMockAPIForPatches{t: t, arguments: []any{ctx, id}, returns: []any{expectedLink}}
    link, err := NewPatchesService(r).Share(ctx, id)

    assert.Equal(t, expectedLink, link)
    assert.NoError(t, err)
  })

  t.Run("wrong patch uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForPatches{}

    link, err := NewPatchesService(r).Share(ctx, id)
    require.False(t, r.called)
    assert.Error(t, err)
    assert.Equal(t, "about:blank", link)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForPatches{returns: []any{""}, errors: unexpected}
    link, err := NewPatchesService(r).Share(ctx, uuid.NewString())
    assert.Equal(t, "about:blank", link)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForPatches) Discard(_ context.Context, patchID string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
  }

  return mock.errors
}

func TestPatchesService_Discard(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForPatches{t: t, arguments: []any{ctx, id}}

    assert.NoError(t, NewPatchesService(r).Discard(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForPatches{errors: unexpected}
    assert.ErrorIs(t, NewPatchesService(r).Discard(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForPatches{}
    assert.Error(t, NewPatchesService(r).Discard(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForPatches) Release(_ context.Context, patchID string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], patchID)
  }

  return mock.errors
}

func TestPatchesService_Release(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForPatches{t: t, arguments: []any{ctx, id}}

    assert.NoError(t, NewPatchesService(r).Release(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForPatches{errors: unexpected}
    assert.ErrorIs(t, NewPatchesService(r).Release(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForPatches{}
    assert.Error(t, NewPatchesService(r).Release(ctx, id))
    assert.False(t, r.called)
  })
}
