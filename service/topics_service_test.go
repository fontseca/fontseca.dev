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

type topicsRepositoryMockAPI struct {
  topicsRepositoryAPI
  t         *testing.T
  arguments []any
  returns   []any
  errors    error
  called    bool
}

type topicsRepositoryThenGetMock struct {
  topicsRepositoryMockAPI
}

func (mock *topicsRepositoryThenGetMock) List(context.Context) ([]*model.Topic, error) {
  return make([]*model.Topic, 2), nil
}

func (mock *topicsRepositoryThenGetMock) Create(_ context.Context, t *transfer.TopicCreation) error {
  if nil != mock.errors {
    return mock.errors
  }

  require.Equal(mock.t, mock.arguments[1], t)
  return nil
}

func TestTopicsService_Add(t *testing.T) {
  ctx := context.TODO()
  creation := &transfer.TopicCreation{
    Name: "Consectetur! Adipiscing... Quis nostrud: ELIT?",
    ID:   "consectetur-adipiscing-quis-nostrud-elit",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.TopicCreation{
      Name: " \n\t " + creation.Name + " \n\t ",
    }

    r := &topicsRepositoryThenGetMock{
      topicsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, creation},
      },
    }
    err := NewTopicsService(r).Create(ctx, dirty)

    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &topicsRepositoryThenGetMock{
      topicsRepositoryMockAPI{
        errors: unexpected,
      },
    }
    err := NewTopicsService(r).Create(ctx, creation)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *topicsRepositoryMockAPI) List(context.Context) ([]*model.Topic, error) {
  mock.called = true
  return mock.returns[0].([]*model.Topic), mock.errors
}

func TestTopicsService_Get(t *testing.T) {
  ctx := context.TODO()

  t.Run("success without cache", func(t *testing.T) {
    expectedTopics := []*model.Topic{{}, {}, {}}
    r := &topicsRepositoryMockAPI{returns: []any{expectedTopics}}
    s := NewTopicsService(r)
    s.cache = nil
    topics, err := s.List(ctx)

    assert.Equal(t, expectedTopics, topics)
    assert.NoError(t, err)
  })

  t.Run("success with cache", func(t *testing.T) {
    expectedTopics := []*model.Topic{{}, {}, {}}

    r := &topicsRepositoryMockAPI{}
    s := NewTopicsService(r)
    s.cache = expectedTopics
    topics, err := s.List(ctx)

    require.False(t, r.called)
    assert.Equal(t, expectedTopics, topics)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := &topicsRepositoryMockAPI{returns: []any{([]*model.Topic)(nil)}, errors: unexpected}
    topics, err := NewTopicsService(r).List(ctx)

    assert.Nil(t, topics)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *topicsRepositoryThenGetMock) Update(_ context.Context, id string, t *transfer.TopicUpdate) error {
  if nil != mock.errors {
    return mock.errors
  }

  require.Equal(mock.t, mock.arguments[1], id)
  require.Equal(mock.t, mock.arguments[2], t)
  return nil
}

func TestTopicsService_Update(t *testing.T) {
  ctx := context.TODO()
  id := "consectetur-adipiscing-quis-nostrud-elit"

  update := &transfer.TopicUpdate{
    ID:   "consectetur-adipiscing-quis-nostrud-elit",
    Name: "Consectetur! Adipiscing... Quis nostrud: ELIT?",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.TopicUpdate{
      Name: " \n\t " + update.Name + " \n\t ",
    }

    r := &topicsRepositoryThenGetMock{
      topicsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, update},
      },
    }
    err := NewTopicsService(r).Update(ctx, id, dirty)

    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &topicsRepositoryThenGetMock{
      topicsRepositoryMockAPI{
        t:         t,
        arguments: []any{ctx, id, update},
        errors:    unexpected,
      },
    }
    err := NewTopicsService(r).Update(ctx, id, update)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *topicsRepositoryThenGetMock) Remove(_ context.Context, id string) error {
  if nil != mock.errors {
    return mock.errors
  }

  require.Equal(mock.t, mock.arguments[1], id)
  return nil
}

func TestTopicsService_Remove(t *testing.T) {
  ctx := context.TODO()
  id := "id"

  t.Run("success", func(t *testing.T) {
    r := &topicsRepositoryThenGetMock{topicsRepositoryMockAPI{t: t, arguments: []any{ctx, id}}}
    assert.NoError(t, NewTopicsService(r).Remove(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := &topicsRepositoryThenGetMock{topicsRepositoryMockAPI{errors: unexpected}}
    assert.ErrorIs(t, NewTopicsService(r).Remove(ctx, id), unexpected)
  })
}
