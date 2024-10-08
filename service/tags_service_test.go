package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "testing"
)

type tagsRepositoryMockAPI struct {
  tagsRepositoryAPI
  t         *testing.T
  arguments []any
  returns   []any
  errors    error
  called    bool
}

type tagsRepositoryThenGetMock struct {
  tagsRepositoryMockAPI
}

func (mock *tagsRepositoryThenGetMock) List(context.Context) ([]*model.Tag, error) {
  return make([]*model.Tag, 2), nil
}

func (mock *tagsRepositoryThenGetMock) Create(_ context.Context, t *transfer.TagCreation) error {
  if nil != mock.errors {
    return mock.errors
  }

  require.Equal(mock.t, mock.arguments[1], t)
  return nil
}

func TestTagsService_Add(t *testing.T) {
  ctx := context.TODO()
  creation := &transfer.TagCreation{
    Name: "Consectetur! Adipiscing... Quis nostrud: ELIT?",
    ID:   "consectetur-adipiscing-quis-nostrud-elit",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.TagCreation{
      Name: " \n\t " + creation.Name + " \n\t ",
    }

    r := &tagsRepositoryThenGetMock{
      tagsRepositoryMockAPI{t: t, arguments: []any{nil, creation}},
    }
    err := NewTagsService(r).Create(ctx, dirty)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &tagsRepositoryThenGetMock{tagsRepositoryMockAPI{errors: unexpected}}
    err := NewTagsService(r).Create(ctx, creation)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *tagsRepositoryMockAPI) List(_ context.Context) ([]*model.Tag, error) {
  mock.called = true
  return mock.returns[0].([]*model.Tag), mock.errors
}

func TestTagsService_Get(t *testing.T) {
  ctx := context.TODO()

  t.Run("success without cache", func(t *testing.T) {
    expectedTags := []*model.Tag{{}, {}, {}}

    r := &tagsRepositoryMockAPI{returns: []any{expectedTags}}
    s := NewTagsService(r)
    s.cache = nil

    tags, err := s.List(ctx)

    assert.Equal(t, expectedTags, tags)
    assert.NoError(t, err)
  })

  t.Run("success with cache", func(t *testing.T) {
    expectedTags := []*model.Tag{{}, {}, {}}

    r := &tagsRepositoryMockAPI{}
    s := NewTagsService(r)

    s.cache = expectedTags

    tags, err := s.List(ctx)

    require.False(t, r.called)
    assert.Equal(t, expectedTags, tags)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &tagsRepositoryMockAPI{returns: []any{([]*model.Tag)(nil)}, errors: unexpected}
    tags, err := NewTagsService(r).List(ctx)

    assert.Nil(t, tags)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *tagsRepositoryThenGetMock) Update(_ context.Context, id string, t *transfer.TagUpdate) error {
  if nil != mock.errors {
    return mock.errors
  }

  require.Equal(mock.t, mock.arguments[1], id)
  require.Equal(mock.t, mock.arguments[2], t)
  return nil
}

func TestTagsService_Update(t *testing.T) {
  ctx := context.TODO()
  id := "consectetur-adipiscing-quis-nostrud-elit"

  update := &transfer.TagUpdate{
    ID:   "consectetur-adipiscing-quis-nostrud-elit",
    Name: "Consectetur! Adipiscing... Quis nostrud: ELIT?",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.TagUpdate{
      Name: " \n\t " + update.Name + " \n\t ",
    }

    r := &tagsRepositoryThenGetMock{tagsRepositoryMockAPI{t: t, arguments: []any{nil, id, update}}}
    err := NewTagsService(r).Update(ctx, id, dirty)

    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &tagsRepositoryThenGetMock{tagsRepositoryMockAPI{errors: unexpected}}
    err := NewTagsService(r).Update(ctx, id, update)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *tagsRepositoryThenGetMock) Remove(_ context.Context, id string) error {
  if nil != mock.errors {
    return mock.errors
  }

  require.Equal(mock.t, mock.arguments[1], id)
  return nil
}

func TestTagsService_Remove(t *testing.T) {
  ctx := context.TODO()
  id := "id"

  t.Run("success", func(t *testing.T) {
    r := &tagsRepositoryThenGetMock{tagsRepositoryMockAPI{t: t, arguments: []any{ctx, id}}}
    assert.NoError(t, NewTagsService(r).Remove(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &tagsRepositoryThenGetMock{tagsRepositoryMockAPI{errors: unexpected}}
    assert.ErrorIs(t, NewTagsService(r).Remove(ctx, id), unexpected)
  })
}
