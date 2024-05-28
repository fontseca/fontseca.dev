package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/transfer"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "testing"
)

func TestTopicsService_Add(t *testing.T) {
  const routine = "Add"

  ctx := context.TODO()
  creation := &transfer.TopicCreation{
    Name: "Consectetur! Adipiscing... Quis nostrud: ELIT?",
    ID:   "consectetur-adipiscing-quis-nostrud-elit",
  }

  t.Run("success", func(t *testing.T) {
    dirty := &transfer.TopicCreation{
      Name: " \n\t " + creation.Name + " \n\t ",
    }

    r := mocks.NewTopicsRepository()
    r.On(routine, ctx, creation).Return(nil)

    err := NewTopicsService(r).Add(ctx, dirty)

    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")
    r := mocks.NewTopicsRepository()
    r.On(routine, ctx, mock.Anything).Return(unexpected)

    err := NewTopicsService(r).Add(ctx, creation)
    assert.ErrorIs(t, err, unexpected)
  })
}
