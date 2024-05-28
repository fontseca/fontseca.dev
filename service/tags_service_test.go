package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "testing"
)

func TestTagsService_Add(t *testing.T) {
  const routine = "Add"

  ctx := context.TODO()
  creation := &transfer.TagCreation{
    Name: "Consectetur! Adipiscing... Quis nostrud: ELIT?",
    ID:   "consectetur-adipiscing-quis-nostrud-elit",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.TagCreation{
      Name: " \n\t " + creation.Name + " \n\t ",
    }

    r := mocks.NewTagsRepository()

    r.On(routine, ctx, creation).Return(nil)
    r.On("Get", ctx).Return([]*model.Tag{{}, {}}, nil)

    err := NewTagsService(r).Add(ctx, dirty)

    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := mocks.NewTagsRepository()
    r.On(routine, ctx, mock.Anything).Return(unexpected)

    err := NewTagsService(r).Add(ctx, creation)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestTagsService_Get(t *testing.T) {
  const routine = "Get"

  ctx := context.TODO()

  t.Run("success without cache", func(t *testing.T) {
    expectedTags := []*model.Tag{{}, {}, {}}

    r := mocks.NewTagsRepository()
    r.On(routine, ctx).Return(expectedTags, nil)

    s := NewTagsService(r).(*tagsService)

    s.cache = nil

    tags, err := s.Get(ctx)

    assert.Equal(t, expectedTags, tags)
    assert.NoError(t, err)
  })

  t.Run("success with cache", func(t *testing.T) {
    expectedTags := []*model.Tag{{}, {}, {}}

    r := mocks.NewTagsRepository()
    r.AssertNotCalled(t, routine)

    s := NewTagsService(r).(*tagsService)

    s.cache = expectedTags

    tags, err := s.Get(ctx)

    assert.Equal(t, expectedTags, tags)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewTagsRepository()
    r.On(routine, ctx).Return(nil, unexpected)

    tags, err := NewTagsService(r).Get(ctx)

    assert.Nil(t, tags)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestTagsService_Update(t *testing.T) {
  const routine = "Update"

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

    r := mocks.NewTagsRepository()

    r.On(routine, ctx, id, update).Return(nil)
    r.On("Get", ctx).Return([]*model.Tag{{}, {}}, nil)

    err := NewTagsService(r).Update(ctx, id, dirty)

    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := mocks.NewTagsRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything).Return(unexpected)

    err := NewTagsService(r).Update(ctx, id, update)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestTagsService_Remove(t *testing.T) {
  const routine = "Remove"

  ctx := context.TODO()
  id := "id"

  t.Run("success", func(t *testing.T) {
    r := mocks.NewTagsRepository()
    r.On(routine, ctx, id).Return(nil)
    r.On("Get", ctx).Return([]*model.Tag{{}, {}}, nil)

    assert.NoError(t, NewTagsService(r).Remove(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewTagsRepository()
    r.On(routine, ctx, id).Return(unexpected)

    assert.ErrorIs(t, NewTagsService(r).Remove(ctx, id), unexpected)
  })
}
