package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "testing"
)

type archiveRepositoryMockAPIForArticles struct {
  archiveRepositoryAPIForArticles
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

func (mock *archiveRepositoryMockAPIForArticles) List(_ context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], filter)
    require.Equal(mock.t, mock.arguments[2], hidden)
    require.Equal(mock.t, mock.arguments[3], draftsOnly)
  }

  return mock.returns[0].([]*transfer.Article), mock.errors
}

func TestArticlesService_Get(t *testing.T) {
  ctx := context.TODO()
  filter := &transfer.ArticleFilter{}

  t.Run("success", func(t *testing.T) {
    expectedArticles := []*transfer.Article{{}, {}, {}}

    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, filter, false, false}, returns: []any{expectedArticles}}
    articles, err := NewArticlesService(r).List(ctx, filter)

    assert.Equal(t, expectedArticles, articles)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForArticles{returns: []any{([]*transfer.Article)(nil)}, errors: unexpected}
    _, err := NewArticlesService(r).List(ctx, filter)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestArticlesService_GetHidden(t *testing.T) {
  ctx := context.TODO()
  filter := &transfer.ArticleFilter{}

  t.Run("success", func(t *testing.T) {
    expectedArticles := []*transfer.Article{{}, {}, {}}

    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, filter, true, false}, returns: []any{expectedArticles}}
    articles, err := NewArticlesService(r).ListHidden(ctx, filter)

    assert.Equal(t, expectedArticles, articles)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForArticles{returns: []any{([]*transfer.Article)(nil)}, errors: unexpected}
    _, err := NewArticlesService(r).ListHidden(ctx, filter)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) GetByID(_ context.Context, draftID string, isDraft bool) (article *model.Article, err error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
    require.Equal(mock.t, mock.arguments[2], isDraft)
  }

  return mock.returns[0].(*model.Article), mock.errors
}

func TestArticlesService_GetByID(t *testing.T) {
  ctx := context.TODO()
  id := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    expectedArticle := &model.Article{}

    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id, false}, returns: []any{expectedArticle}}
    article, err := NewArticlesService(r).GetByID(ctx, id)

    assert.Equal(t, expectedArticle, article)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForArticles{returns: []any{(*model.Article)(nil)}, errors: unexpected}
    article, err := NewArticlesService(r).GetByID(ctx, id)

    assert.Nil(t, article)
    assert.ErrorIs(t, err, unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForArticles{}
    _, err := NewArticlesService(r).GetByID(ctx, id)

    require.False(t, r.called)
    assert.Error(t, err)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) SetHidden(_ context.Context, articleID string, hidden bool) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
    require.Equal(mock.t, mock.arguments[2], hidden)
  }

  return mock.errors
}

func TestArticlesService_Hide(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id, true}}

    assert.NoError(t, NewArticlesService(r).Hide(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    assert.ErrorIs(t, NewArticlesService(r).Hide(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).Hide(ctx, id))
    assert.False(t, r.called)
  })
}

func TestArticlesService_Show(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id, false}}
    assert.NoError(t, NewArticlesService(r).Show(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    assert.ErrorIs(t, NewArticlesService(r).Show(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).Show(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) Amend(_ context.Context, articleID string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesService_Amend(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id}}
    assert.NoError(t, NewArticlesService(r).Amend(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    assert.ErrorIs(t, NewArticlesService(r).Amend(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).Amend(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) Remove(_ context.Context, articleID string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
  }

  return mock.errors
}

func TestArticlesService_Remove(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id}}

    assert.NoError(t, NewArticlesService(r).Remove(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    assert.ErrorIs(t, NewArticlesService(r).Remove(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).Remove(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) SetPinned(_ context.Context, articleID string, pinned bool) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
    require.Equal(mock.t, mock.arguments[2], pinned)
  }

  return mock.errors
}

func TestArticlesService_Pin(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id, true}}
    assert.NoError(t, NewArticlesService(r).Pin(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    assert.ErrorIs(t, NewArticlesService(r).Pin(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"
    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).Pin(ctx, id))
    assert.False(t, r.called)
  })
}

func TestArticlesService_Unpin(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, id, false}}
    assert.NoError(t, NewArticlesService(r).Unpin(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    assert.ErrorIs(t, NewArticlesService(r).Unpin(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).Unpin(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) AddTag(_ context.Context, draftID, tagID string, isDraft ...bool) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
    require.Equal(mock.t, mock.arguments[2], tagID)
    require.Equal(mock.t, mock.arguments[3], isDraft)
  }

  return mock.errors
}

func TestArticlesService_AddTag(t *testing.T) {
  ctx := context.TODO()
  articleUUID := uuid.New().String()
  tagID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, articleUUID, tagID, []bool(nil)}}
    assert.NoError(t, NewArticlesService(r).AddTag(ctx, articleUUID, tagID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    articleUUID = "e4d06ba7-f086-47dc-9f5e"
    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).AddTag(ctx, articleUUID, tagID))
    assert.False(t, r.called)
  })

  articleUUID = uuid.NewString()

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    err := NewArticlesService(r).AddTag(ctx, articleUUID, uuid.NewString())
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForArticles) RemoveTag(_ context.Context, articleID, tagID string, isDraft ...bool) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], articleID)
    require.Equal(mock.t, mock.arguments[2], tagID)
    require.Equal(mock.t, mock.arguments[3], isDraft)
  }

  return mock.errors
}

func TestArticlesService_RemoveTag(t *testing.T) {
  ctx := context.TODO()
  articleUUID := uuid.New().String()
  tagID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{t: t, arguments: []any{ctx, articleUUID, tagID, []bool(nil)}}
    assert.NoError(t, NewArticlesService(r).RemoveTag(ctx, articleUUID, tagID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForArticles{}
    assert.Error(t, NewArticlesService(r).RemoveTag(ctx, "e4d06ba7-f086-47dc-9f5e", tagID))
    assert.False(t, r.called)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForArticles{errors: unexpected}
    err := NewArticlesService(r).RemoveTag(ctx, articleUUID, uuid.NewString())
    assert.ErrorIs(t, err, unexpected)
  })
}
