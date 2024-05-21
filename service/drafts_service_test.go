package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "testing"
)

func TestDraftsService_Draft(t *testing.T) {
  const routine = "Draft"
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

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, creation).Return(id.String(), nil)

    insertedID, err := NewDraftsService(r).Draft(ctx, dirty)

    assert.NoError(t, err)
    assert.Equal(t, id, insertedID)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, mock.Anything).Return("", unexpected)
    s := NewDraftsService(r)

    insertedID, err := s.Draft(ctx, creation)

    assert.ErrorIs(t, err, unexpected)
    assert.Equal(t, uuid.Nil, insertedID)
  })
}

func TestDraftsService_Publish(t *testing.T) {
  const routine = "Publish"

  ctx := context.TODO()
  id := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id).Return(nil)

    assert.NoError(t, NewDraftsService(r).Publish(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewDraftsService(r).Publish(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewDraftsService(r).Publish(ctx, id))
  })
}
