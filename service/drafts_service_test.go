package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "strings"
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

func TestDraftsService_Get(t *testing.T) {
  const routine = "Get"

  ctx := context.TODO()

  t.Run("success", func(t *testing.T) {
    expectedDrafts := make([]*model.Article, 3)

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, "", false, true).Return(expectedDrafts, nil)

    drafts, err := NewDraftsService(r).Get(ctx, "\n \t \n")

    assert.Equal(t, expectedDrafts, drafts)
    assert.NoError(t, err)
  })

  t.Run("success with search", func(t *testing.T) {
    expectedDrafts := make([]*model.Article, 3)
    expectedNeedle := "20 www xxx yyy zzz zzz"

    needle := ">> = 20 www? xxx! yyy... zzz_zzz \" ' ° <<"

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, expectedNeedle, false, true).Return(expectedDrafts, nil)

    drafts, err := NewDraftsService(r).Get(ctx, needle)

    assert.Equal(t, expectedDrafts, drafts)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, unexpected)

    _, err := NewDraftsService(r).Get(ctx, "")
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestDraftsService_GetByID(t *testing.T) {
  const routine = "GetByID"

  ctx := context.TODO()
  id := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    expectedDraft := &model.Article{}

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id, true).Return(expectedDraft, nil)

    draft, err := NewDraftsService(r).GetByID(ctx, id)

    assert.Equal(t, expectedDraft, draft)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(nil, unexpected)

    draft, err := NewDraftsService(r).GetByID(ctx, id)

    assert.Nil(t, draft)
    assert.ErrorIs(t, err, unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    _, err := NewDraftsService(r).GetByID(ctx, id)

    assert.Error(t, err)
  })
}

func TestDraftsService_AddTopic(t *testing.T) {
  const routine = "AddTopic"

  ctx := context.TODO()
  draftUUID := uuid.New().String()
  topicID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, draftUUID, topicID, []bool{true}).Return(nil)

    assert.NoError(t, NewDraftsService(r).AddTopic(ctx, draftUUID, topicID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    draftUUID = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewDraftsService(r).AddTopic(ctx, draftUUID, topicID))
  })

  draftUUID = uuid.NewString()

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    err := NewDraftsService(r).AddTopic(ctx, draftUUID, uuid.NewString())

    assert.ErrorIs(t, err, unexpected)
  })
}

func TestDraftsService_RemoveTopic(t *testing.T) {
  const routine = "RemoveTopic"

  ctx := context.TODO()
  draftUUID := uuid.New().String()
  topicID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, draftUUID, topicID, []bool{true}).Return(nil)

    assert.NoError(t, NewDraftsService(r).RemoveTopic(ctx, draftUUID, topicID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    draftUUID = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewDraftsService(r).RemoveTopic(ctx, draftUUID, topicID))
  })

  draftUUID = uuid.NewString()

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    err := NewDraftsService(r).RemoveTopic(ctx, draftUUID, uuid.NewString())

    assert.ErrorIs(t, err, unexpected)
  })
}

func TestDraftsService_Share(t *testing.T) {
  const routine = "Share"

  ctx := context.TODO()
  draftUUID := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    expectedLink := "link-to-resource"

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, draftUUID).Return(expectedLink, nil)

    link, err := NewDraftsService(r).Share(ctx, draftUUID)

    assert.Equal(t, expectedLink, link)
    assert.NoError(t, err)
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    draftUUID = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    link, err := NewDraftsService(r).Share(ctx, draftUUID)

    assert.Error(t, err)
    assert.Equal(t, "about:blank", link)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return("", unexpected)

    link, err := NewDraftsService(r).Share(ctx, uuid.NewString())

    assert.Equal(t, "about:blank", link)
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestDraftsService_Discard(t *testing.T) {
  const routine = "Discard"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id).Return(nil)

    assert.NoError(t, NewDraftsService(r).Discard(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewDraftsService(r).Discard(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewDraftsService(r).Discard(ctx, id))
  })
}

func TestDraftsService_Revise(t *testing.T) {
  const routine = "Revise"
  ctx := context.TODO()
  draftUUID := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    revision := &transfer.ArticleUpdate{
      Title:    "Consectetur! Adipiscing... Quis nostrud: ELIT?",
      Slug:     "consectetur-adipiscing-quis-nostrud-elit",
      ReadTime: 11,
      Content:  strings.Repeat("word ", 1999) + "word",
    }

    dirty := &transfer.ArticleUpdate{
      Title:   " \t\n " + revision.Title + " \t\n ",
      Content: " \t\n " + revision.Content + " \t\n ",
    }

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, draftUUID, revision).Return(nil)

    assert.NoError(t, NewDraftsService(r).Revise(ctx, draftUUID, dirty))
  })

  t.Run("success: changing title", func(t *testing.T) {
    revision := &transfer.ArticleUpdate{
      Title:    "Consectetur-Adipiscing!!... Quis nostrud: ELIT??? +-'\"",
      Slug:     "consectetur-adipiscing-quis-nostrud-elit",
      ReadTime: 1,
    }

    dirty := &transfer.ArticleUpdate{
      Title: " \t\n " + revision.Title + " \t\n ",
    }

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, revision).Return(nil)

    assert.NoError(t, NewDraftsService(r).Revise(ctx, draftUUID, dirty))
  })

  t.Run("success: changing content", func(t *testing.T) {
    revision := &transfer.ArticleUpdate{
      Content:  strings.Repeat("word ", 299) + "word",
      ReadTime: 2,
    }

    dirty := &transfer.ArticleUpdate{
      Content: " \t\n " + revision.Content + " \t\n ",
    }

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, revision).Return(nil)

    assert.NoError(t, NewDraftsService(r).Revise(ctx, draftUUID, dirty))
  })

  t.Run("nil parameter: revision", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)
    assert.ErrorContains(t, NewDraftsService(r).Revise(ctx, draftUUID, nil), "nil value")
  })

  t.Run("wrong uuid: draftUUID", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)
    assert.Error(t, NewDraftsService(r).Revise(ctx, "x", &transfer.ArticleUpdate{}))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewDraftsService(r).Revise(ctx, draftUUID, &transfer.ArticleUpdate{}), unexpected)
  })
}
