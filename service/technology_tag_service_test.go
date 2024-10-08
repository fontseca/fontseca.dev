package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "strings"
  "testing"
)

type technologyTagRepositoryMockAPI struct {
  technologyTagRepositoryAPI
  t         *testing.T
  arguments []any
  returns   []any
  errors    error
  called    bool
}

func (mock *technologyTagRepositoryMockAPI) List(context.Context) ([]*model.TechnologyTag, error) {
  return mock.returns[0].([]*model.TechnologyTag), mock.errors
}

func TestTechnologyTagService_Get(t *testing.T) {
  t.Run("success", func(t *testing.T) {
    var r = &technologyTagRepositoryMockAPI{returns: []any{[]*model.TechnologyTag{}}}
    var ctx = context.Background()
    res, err := NewTechnologyTagService(r).List(ctx)
    assert.NotNil(t, res)
    assert.NoError(t, err)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &technologyTagRepositoryMockAPI{returns: []any{([]*model.TechnologyTag)(nil)}, errors: unexpected}
    var ctx = context.Background()
    res, err := NewTechnologyTagService(r).List(ctx)
    assert.Nil(t, res)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *technologyTagRepositoryMockAPI) Create(_ context.Context, t *transfer.TechnologyTagCreation) (string, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], t)
  }

  return mock.returns[0].(string), mock.errors
}

func TestTechnologyTagService_Add(t *testing.T) {
  var creation = &transfer.TechnologyTagCreation{Name: "Technology Tag Name"}
  var ctx = context.Background()

  t.Run("success", func(t *testing.T) {
    var dirty = &transfer.TechnologyTagCreation{Name: "  \n\t\n  " + creation.Name + "  \n\t\n  "}
    var id = uuid.New().String()
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, creation},
      returns:   []any{id},
      errors:    nil,
    }
    res, err := NewTechnologyTagService(r).Create(ctx, dirty)
    assert.Equal(t, id, res)
    assert.NoError(t, err)
  })

  t.Run("error on nil creation", func(t *testing.T) {
    var r = &technologyTagRepositoryMockAPI{}
    res, err := NewTechnologyTagService(r).Create(ctx, nil)
    require.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: creation")
    assert.Empty(t, res)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &technologyTagRepositoryMockAPI{
      returns: []any{""},
      errors:  unexpected,
    }
    res, err := NewTechnologyTagService(r).Create(ctx, creation)
    assert.Empty(t, res)
    assert.ErrorIs(t, err, unexpected)
  })

  t.Run("max len (64) exceeded", func(t *testing.T) {
    var p = problem.NewValidation([3]string{"name", "max", "64"})
    var r = &technologyTagRepositoryMockAPI{}
    creation.Name = strings.Repeat("x", 65)
    res, err := NewTechnologyTagService(r).Create(ctx, creation)
    require.False(t, r.called)
    assert.ErrorAs(t, err, &p)
    assert.Empty(t, res)
  })
}

func (mock *technologyTagRepositoryMockAPI) Exists(_ context.Context, id string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
  }

  return mock.errors
}

func TestTechnologyTagService_Exists(t *testing.T) {
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success: does exist", func(t *testing.T) {
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id},
      errors:    nil,
    }
    err := NewTechnologyTagService(r).Exists(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("success: does not exist", func(t *testing.T) {
    var p = problem.NewNotFound(id, "technology_tag")
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id},
      errors:    p,
    }
    err := NewTechnologyTagService(r).Exists(ctx, id)
    assert.ErrorAs(t, err, &p)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &technologyTagRepositoryMockAPI{errors: unexpected}
    err := NewTechnologyTagService(r).Exists(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}

func (mock *technologyTagRepositoryMockAPI) Update(_ context.Context, id string, t *transfer.TechnologyTagUpdate) (bool, error) {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
    require.Equal(mock.t, mock.arguments[2], t)
  }

  return mock.returns[0].(bool), mock.errors
}

func TestTechnologyTagService_Update(t *testing.T) {
  var update = &transfer.TechnologyTagUpdate{Name: strings.Repeat("x", 64)}
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success", func(t *testing.T) {
    var dirty = &transfer.TechnologyTagUpdate{Name: "  \n\t\n  " + update.Name + "  \n\t\n  "}
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id, update},
      returns:   []any{true},
      errors:    nil,
    }
    res, err := NewTechnologyTagService(r).Update(ctx, id, dirty)
    assert.True(t, res)
    assert.NoError(t, err)
  })

  t.Run("error on nil update", func(t *testing.T) {
    var r = &technologyTagRepositoryMockAPI{}
    res, err := NewTechnologyTagService(r).Update(ctx, id, nil)
    require.False(t, r.called)
    assert.ErrorContains(t, err, "nil value for parameter: update")
    assert.False(t, res)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &technologyTagRepositoryMockAPI{
      returns: []any{false},
      errors:  unexpected,
    }
    res, err := NewTechnologyTagService(r).Update(ctx, id, update)
    assert.False(t, res)
    assert.ErrorIs(t, err, unexpected)
  })

  t.Run("max len (64) exceeded", func(t *testing.T) {
    var p = problem.NewValidation([3]string{"name", "max", "64"})
    var r = &technologyTagRepositoryMockAPI{}
    update.Name = strings.Repeat("x", 65)
    res, err := NewTechnologyTagService(r).Update(ctx, id, update)
    require.False(t, r.called)
    assert.ErrorAs(t, err, &p)
    assert.False(t, res)
  })
}

func (mock *technologyTagRepositoryMockAPI) Remove(_ context.Context, id string) error {
  mock.called = true

  if nil != mock.t {
    require.Equal(mock.t, mock.arguments[1], id)
  }

  return mock.errors
}

func TestTechnologyTagService_Remove(t *testing.T) {
  var ctx = context.Background()
  var id = uuid.New().String()

  t.Run("success: does exist", func(t *testing.T) {
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id},
      errors:    nil,
    }
    err := NewTechnologyTagService(r).Remove(ctx, id)
    assert.NoError(t, err)
  })

  t.Run("success: does not exist", func(t *testing.T) {
    var p = problem.NewNotFound(id, "technology_tag")
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id},
      errors:    p,
    }
    err := NewTechnologyTagService(r).Remove(ctx, id)
    assert.ErrorAs(t, err, &p)
  })

  t.Run("got an error", func(t *testing.T) {
    var unexpected = errors.New("unexpected error")
    var r = &technologyTagRepositoryMockAPI{
      t:         t,
      arguments: []any{ctx, id},
      errors:    unexpected,
    }
    err := NewTechnologyTagService(r).Remove(ctx, id)
    assert.ErrorIs(t, err, unexpected)
  })
}
