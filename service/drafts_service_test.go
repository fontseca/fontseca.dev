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

type archiveRepositoryMockAPIForDrafts struct {
  archiveRepositoryAPIForDrafts
  t         *testing.T
  returns   []any
  arguments []any
  errors    error
  called    bool
}

func (mock *archiveRepositoryMockAPIForDrafts) Draft(_ context.Context, creation *transfer.ArticleCreation) (draft string, err error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], creation)
  }

  return mock.returns[0].(string), mock.errors
}

func TestDraftsService_Draft(t *testing.T) {
  id := uuid.New()
  ctx := context.TODO()
  creation := &transfer.ArticleCreation{
    Title:    "Consectetur! Adipiscing... Quis nostrud: ELIT?",
    Slug:     "consectetur-adipiscing-quis-nostrud-elit",
    ReadTime: 1,
    Content:  "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.ArticleCreation{
      Title:   " \n\t " + creation.Title + " \n\t ",
      Content: creation.Content,
    }

    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, creation}, returns: []any{id.String()}}
    insertedID, err := NewDraftsService(r).Draft(ctx, dirty)

    assert.NoError(t, err)
    assert.Equal(t, id, insertedID)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForDrafts{returns: []any{""}, errors: unexpected}
    s := NewDraftsService(r)

    insertedID, err := s.Draft(ctx, creation)

    assert.ErrorIs(t, err, unexpected)
    assert.Equal(t, uuid.Nil, insertedID)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) Publish(_ context.Context, draftID string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
  }

  return mock.errors
}

func TestDraftsService_Publish(t *testing.T) {
  ctx := context.TODO()
  id := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, id}}
    assert.NoError(t, NewDraftsService(r).Publish(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForDrafts{errors: unexpected}
    assert.ErrorIs(t, NewDraftsService(r).Publish(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForDrafts{}
    assert.Error(t, NewDraftsService(r).Publish(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) Get(_ context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (drafts []*transfer.Article, err error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], filter)
    require.Equal(mock.t, mock.arguments[2], hidden)
    require.Equal(mock.t, mock.arguments[3], draftsOnly)
  }

  return mock.returns[0].([]*transfer.Article), mock.errors
}

func TestDraftsService_Get(t *testing.T) {
  ctx := context.TODO()
  filter := &transfer.ArticleFilter{}

  t.Run("success", func(t *testing.T) {
    expectedDrafts := []*transfer.Article{{}, {}, {}}
    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, filter, false, true}, returns: []any{expectedDrafts}}
    drafts, err := NewDraftsService(r).Get(ctx, filter)
    assert.Equal(t, expectedDrafts, drafts)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForDrafts{returns: []any{([]*transfer.Article)(nil)}, errors: unexpected}
    _, err := NewDraftsService(r).Get(ctx, filter)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) GetByID(_ context.Context, draftID string, isDraft bool) (draft *model.Article, err error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
    require.Equal(mock.t, mock.arguments[2], isDraft)
  }

  return mock.returns[0].(*model.Article), mock.errors
}

func TestDraftsService_GetByID(t *testing.T) {
  ctx := context.TODO()
  id := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    expectedDraft := &model.Article{}
    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, id, true}, returns: []any{expectedDraft}}
    draft, err := NewDraftsService(r).GetByID(ctx, id)
    assert.Equal(t, expectedDraft, draft)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForDrafts{returns: []any{(*model.Article)(nil)}, errors: unexpected}
    draft, err := NewDraftsService(r).GetByID(ctx, id)
    assert.Nil(t, draft)
    assert.ErrorIs(t, err, unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForDrafts{}
    _, err := NewDraftsService(r).GetByID(ctx, id)
    require.False(t, r.called)
    assert.Error(t, err)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) AddTag(_ context.Context, draftID, tagID string, isDraft ...bool) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
    require.Equal(mock.t, mock.arguments[2], tagID)
    require.Equal(mock.t, mock.arguments[3], isDraft)
  }

  return mock.errors
}

func TestDraftsService_AddTag(t *testing.T) {
  ctx := context.TODO()
  draftUUID := uuid.New().String()
  tagID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, draftUUID, tagID, []bool{true}}}
    assert.NoError(t, NewDraftsService(r).AddTag(ctx, draftUUID, tagID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    draftUUID = "e4d06ba7-f086-47dc-9f5e"
    r := &archiveRepositoryMockAPIForDrafts{}
    assert.Error(t, NewDraftsService(r).AddTag(ctx, draftUUID, tagID))
    assert.False(t, r.called)
  })

  draftUUID = uuid.NewString()

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForDrafts{errors: unexpected}
    err := NewDraftsService(r).AddTag(ctx, draftUUID, uuid.NewString())
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) RemoveTag(_ context.Context, draftID, tagID string, isDraft ...bool) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
    require.Equal(mock.t, mock.arguments[2], tagID)
    require.Equal(mock.t, mock.arguments[3], isDraft)
  }

  return mock.errors
}

func TestDraftsService_RemoveTag(t *testing.T) {
  ctx := context.TODO()
  draftUUID := uuid.New().String()
  tagID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForDrafts{arguments: []any{ctx, draftUUID, tagID, []bool{true}}}
    assert.NoError(t, NewDraftsService(r).RemoveTag(ctx, draftUUID, tagID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    draftUUID = "e4d06ba7-f086-47dc-9f5e"
    r := &archiveRepositoryMockAPIForDrafts{}
    assert.Error(t, NewDraftsService(r).RemoveTag(ctx, draftUUID, tagID))
    assert.False(t, r.called)
  })

  draftUUID = uuid.NewString()

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &archiveRepositoryMockAPIForDrafts{errors: unexpected}
    err := NewDraftsService(r).RemoveTag(ctx, draftUUID, uuid.NewString())
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) Share(_ context.Context, draftID string) (link string, err error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
  }

  return mock.returns[0].(string), mock.errors
}

func TestDraftsService_Share(t *testing.T) {
  ctx := context.TODO()
  draftUUID := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    expectedLink := "link-to-resource"

    r := &archiveRepositoryMockAPIForDrafts{arguments: []any{ctx, draftUUID}, returns: []any{expectedLink}}

    link, err := NewDraftsService(r).Share(ctx, draftUUID)

    assert.Equal(t, expectedLink, link)
    assert.NoError(t, err)
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    draftUUID = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForDrafts{}

    link, err := NewDraftsService(r).Share(ctx, draftUUID)

    require.False(t, r.called)
    assert.Error(t, err)
    assert.Equal(t, "about:blank", link)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForDrafts{returns: []any{""}, errors: unexpected}

    link, err := NewDraftsService(r).Share(ctx, uuid.NewString())

    assert.Equal(t, "about:blank", link)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) Discard(_ context.Context, draftID string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
  }

  return mock.errors
}

func TestDraftsService_Discard(t *testing.T) {
  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, id}}

    assert.NoError(t, NewDraftsService(r).Discard(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForDrafts{errors: unexpected}

    assert.ErrorIs(t, NewDraftsService(r).Discard(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := &archiveRepositoryMockAPIForDrafts{}
    assert.Error(t, NewDraftsService(r).Discard(ctx, id))
    assert.False(t, r.called)
  })
}

func (mock *archiveRepositoryMockAPIForDrafts) Revise(_ context.Context, draftID string, revision *transfer.ArticleRevision) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], draftID)
    require.Equal(mock.t, mock.arguments[2], revision)
  }

  return mock.errors
}

func TestDraftsService_Revise(t *testing.T) {
  ctx := context.TODO()
  draftUUID := uuid.NewString()

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

    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, draftUUID, revision}}

    assert.NoError(t, NewDraftsService(r).Revise(ctx, draftUUID, dirty))
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

    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, draftUUID, revision}}

    assert.NoError(t, NewDraftsService(r).Revise(ctx, draftUUID, dirty))
  })

  t.Run("success: changing content", func(t *testing.T) {
    revision := &transfer.ArticleRevision{
      Content:  strings.Repeat("word ", 299) + "word",
      ReadTime: 2,
    }

    dirty := &transfer.ArticleRevision{
      Content: " \t\n " + revision.Content + " \t\n ",
    }

    r := &archiveRepositoryMockAPIForDrafts{t: t, arguments: []any{ctx, draftUUID, revision}}

    assert.NoError(t, NewDraftsService(r).Revise(ctx, draftUUID, dirty))
  })

  t.Run("nil parameter: revision", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForDrafts{}
    assert.ErrorContains(t, NewDraftsService(r).Revise(ctx, draftUUID, nil), "nil value")
    assert.False(t, r.called)
  })

  t.Run("wrong uuid: draftUUID", func(t *testing.T) {
    r := &archiveRepositoryMockAPIForDrafts{}
    assert.Error(t, NewDraftsService(r).Revise(ctx, "x", &transfer.ArticleRevision{}))
    assert.False(t, r.called)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &archiveRepositoryMockAPIForDrafts{errors: unexpected}

    assert.ErrorIs(t, NewDraftsService(r).Revise(ctx, draftUUID, &transfer.ArticleRevision{}), unexpected)
  })
}
