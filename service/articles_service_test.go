package service

import (
  "context"
  "errors"
  "fontseca.dev/mocks"
  "fontseca.dev/model"
  "github.com/google/uuid"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "testing"
)

func TestArticlesService_Get(t *testing.T) {
  const routine = "Get"

  ctx := context.TODO()

  t.Run("success", func(t *testing.T) {
    expectedArticles := make([]*model.Article, 3)

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, "", false, false).Return(expectedArticles, nil)

    articles, err := NewArticlesService(r).Get(ctx, "\n \t \n")

    assert.Equal(t, expectedArticles, articles)
    assert.NoError(t, err)
  })

  t.Run("success with search", func(t *testing.T) {
    expectedArticles := make([]*model.Article, 3)
    expectedNeedle := "20 www xxx yyy zzz zzz"

    needle := ">> = 20 www? xxx! yyy... zzz_zzz \" ' ° <<"

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, expectedNeedle, false, false).Return(expectedArticles, nil)

    articles, err := NewArticlesService(r).Get(ctx, needle)

    assert.Equal(t, expectedArticles, articles)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, unexpected)

    _, err := NewArticlesService(r).Get(ctx, "")
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestArticlesService_GetHidden(t *testing.T) {
  const routine = "Get"

  ctx := context.TODO()

  t.Run("success", func(t *testing.T) {
    expectedArticles := make([]*model.Article, 3)

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, "", true, false).Return(expectedArticles, nil)

    articles, err := NewArticlesService(r).GetHidden(ctx, "\n \t \n")

    assert.Equal(t, expectedArticles, articles)
    assert.NoError(t, err)
  })

  t.Run("success with search", func(t *testing.T) {
    expectedArticles := make([]*model.Article, 3)
    expectedNeedle := "20 www xxx yyy zzz zzz"

    needle := ">> = 20 www? xxx! yyy... zzz_zzz \" ' ° <<"

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, expectedNeedle, true, false).Return(expectedArticles, nil)

    articles, err := NewArticlesService(r).GetHidden(ctx, needle)

    assert.Equal(t, expectedArticles, articles)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, unexpected)

    _, err := NewArticlesService(r).GetHidden(ctx, "")
    assert.ErrorIs(t, err, unexpected)
  })
}

func TestArticlesService_GetByID(t *testing.T) {
  const routine = "GetByID"

  ctx := context.TODO()
  id := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    expectedArticle := &model.Article{}

    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id, true).Return(expectedArticle, nil)

    article, err := NewArticlesService(r).GetByID(ctx, id)

    assert.Equal(t, expectedArticle, article)
    assert.NoError(t, err)
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(nil, unexpected)

    article, err := NewArticlesService(r).GetByID(ctx, id)

    assert.Nil(t, article)
    assert.ErrorIs(t, err, unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    _, err := NewArticlesService(r).GetByID(ctx, id)

    assert.Error(t, err)
  })
}

func TestArticlesService_Hide(t *testing.T) {
  const routine = "SetHidden"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id, true).Return(nil)

    assert.NoError(t, NewArticlesService(r).Hide(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewArticlesService(r).Hide(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).Hide(ctx, id))
  })
}

func TestArticlesService_Show(t *testing.T) {
  const routine = "SetHidden"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id, false).Return(nil)

    assert.NoError(t, NewArticlesService(r).Show(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewArticlesService(r).Show(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).Show(ctx, id))
  })
}

func TestArticlesService_Amend(t *testing.T) {
  const routine = "Amend"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id).Return(nil)

    assert.NoError(t, NewArticlesService(r).Amend(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewArticlesService(r).Amend(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).Amend(ctx, id))
  })
}

func TestArticlesService_Remove(t *testing.T) {
  const routine = "Remove"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id).Return(nil)

    assert.NoError(t, NewArticlesService(r).Remove(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewArticlesService(r).Remove(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).Remove(ctx, id))
  })
}

func TestArticlesService_Pin(t *testing.T) {
  const routine = "SetPinned"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id, true).Return(nil)

    assert.NoError(t, NewArticlesService(r).Pin(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewArticlesService(r).Pin(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).Pin(ctx, id))
  })
}

func TestArticlesService_Unpin(t *testing.T) {
  const routine = "SetPinned"

  ctx := context.TODO()
  id := uuid.NewString()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, id, false).Return(nil)

    assert.NoError(t, NewArticlesService(r).Unpin(ctx, id))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    assert.ErrorIs(t, NewArticlesService(r).Unpin(ctx, id), unexpected)
  })

  t.Run("wrong uuid", func(t *testing.T) {
    id = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).Unpin(ctx, id))
  })
}

func TestArticlesService_AddTopic(t *testing.T) {
  const routine = "AddTopic"

  ctx := context.TODO()
  articleUUID := uuid.New().String()
  topicUUID := uuid.New().String()

  t.Run("success", func(t *testing.T) {
    r := mocks.NewArchiveRepository()
    r.On(routine, ctx, articleUUID, topicUUID).Return(nil)

    assert.NoError(t, NewArticlesService(r).AddTopic(ctx, articleUUID, topicUUID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    articleUUID = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).AddTopic(ctx, articleUUID, topicUUID))
  })

  t.Run("wrong draft uuid", func(t *testing.T) {
    articleUUID = topicUUID
    topicUUID = "e4d06ba7-f086-47dc-9f5e"

    r := mocks.NewArchiveRepository()
    r.AssertNotCalled(t, routine)

    assert.Error(t, NewArticlesService(r).AddTopic(ctx, articleUUID, topicUUID))
  })

  t.Run("gets a repository failure", func(t *testing.T) {
    unexpected := errors.New("unexpected error")

    r := mocks.NewArchiveRepository()
    r.On(routine, mock.Anything, mock.Anything, mock.Anything).Return(unexpected)

    err := NewArticlesService(r).AddTopic(ctx, articleUUID, uuid.NewString())

    assert.ErrorIs(t, err, unexpected)
  })
}
