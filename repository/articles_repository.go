package repository

import (
  "context"
  "crypto/sha1"
  "database/sql"
  "errors"
  "fmt"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "log/slog"
  "net/http"
  "strings"
  "time"
)

// ArticlesRepository is a common API for articles, article drafts
// and article patches.
//
// An article is a piece of writing about a particular subject in my
// website's archive. Naturally, every article has one or more topics
// that are inherent to the discussion of the article.
//
// An article draft, or just draft, is a rough version of an article
// that is not yet in its final form nor is it published. The main
// purpose of a draft is to provide a description of the main facts
// or points involved in the subject. This draft is improved by
// making revisions. You can share a draft in order to get feedback.
// Once the draft seems enticing and complete, it is published.
//
// When an article needs to be improved or amended, an article patch
// is created internally to record every enhancement made to the main
// article. The article is still available online during this process.
// An article patch, or simply patch, is an internal entity that points
// to the article it's been improving; it is a temporal place to store
// updates and enhancements. The patch is also shared to get feedback
// and improved by revisions. Once the patch is coherent and complete,
// it is released and physically merged to the original article.
//
// The draft and the article are both referenced by the same UUID. The
// patch is a completely different object that points to an article.
// Since an article can only have one patch at a time, by using the
// article's UUID, you can access any patch it currently has.
type ArticlesRepository interface {
  // Draft starts the creation process of an article. It returns the
  // UUID of the draft that was created.
  //
  // To draft an article, only its title is required, other fields
  // are completely optional and can be added in an eventual revision.
  Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error)

  // Publish makes a draft publicly available.
  //
  // Invoking Publish on an already published article or a patch has
  // no effect.
  Publish(ctx context.Context, id string) error

  // Get retrieves all the articles that are either hidden or not. If
  // draftsOnly is true, then only retrieves all the ongoing drafts.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over non-hidden articles, so it attempts to find and
  // amass every article whose title contains any of the keywords
  // (if more than one) in needle.
  Get(ctx context.Context, needle string, hidden, draftsOnly bool) (articles []*model.Article, err error)

  // GetByID retrieves one article (or article draft) by its UUID.
  GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error)

  // Amend starts the process to update an article. To amend the article,
  // a public copy of it is kept available to everyone while a patch
  // is created to store any revision made to the article.
  //
  // If the article is still a draft, or it's already being amended,
  // any call to this method has no effect.
  Amend(ctx context.Context, id string) error

  // Remove completely removes an article and any patch it currently
  // has from the database. If the article is a draft, calling Remove
  // has no effect on it whatsoever.
  //
  // If you want to remove a draft, use Discard instead.
  Remove(ctx context.Context, id string) error

  // AddTopic adds a topic to the article. If the topic already
  // exists, it returns an error informing about a conflicting
  // state.
  AddTopic(ctx context.Context, articleID, topicID string) error

  // RemoveTopic removes a topic from the article. If the article has
  // no topic identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTopic(ctx context.Context, articleID, topicID string) error

  // SetHidden hides or shows an article depending on the value of hidden.
  SetHidden(ctx context.Context, id string, hidden bool) error

  // SetPinned pins or unpins an article depending on the value of pinned.
  SetPinned(ctx context.Context, id string, pinned bool) error

  // Share creates a shareable link for a draft or a patch. Only users
  // with that link can see the progress and provide feedback.
  //
  // A shareable link does not make an article public. This link will
  // eventually expire after a certain amount of time.
  Share(ctx context.Context, id string) (link string, err error)

  // Discard completely drops a draft; otherwise if called on a patch
  // it discards it but keeps the original article.
  //
  // This method has no effect on an article.
  Discard(ctx context.Context, id string) error

  // Revise adds a correction or inclusion to a draft or patch in order
  // to correct or improve it.
  Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error

  // Release merges patch into the original article and published the
  // update immediately after merging.
  //
  // This method works only for patches.
  Release(ctx context.Context, id string) error

  // GetPatches retrieves all the ongoing patches of every article.
  GetPatches(ctx context.Context) (patches []*model.ArticlePatch, err error)
}

type articlesRepository struct {
  db *sql.DB
}

func NewArticlesRepository(db *sql.DB) ArticlesRepository {
  return &articlesRepository{db}
}

func (r *articlesRepository) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }

  defer tx.Rollback()

  draftArticleQuery := `
  INSERT INTO "article" ("title", "author", "slug", "read_time", "content")
                 VALUES (@title, 'fontseca.dev', @slug, @read_time, @content)
              RETURNING "uuid";`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result := tx.QueryRowContext(ctx, draftArticleQuery,
    sql.Named("title", creation.Title),
    sql.Named("slug", creation.Slug),
    sql.Named("read_time", creation.ReadTime),
    sql.Named("content", creation.Content),
  )

  if err = result.Scan(&id); nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }

  return id, nil
}

func (r *articlesRepository) Publish(ctx context.Context, id string) error {
  isArticleDraftQuery := `
  SELECT "draft" IS TRUE
     AND "published_at" IS NULL
    FROM "article"
   WHERE "uuid" = @uuid;`

  ctx1, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()

  row := r.db.QueryRowContext(ctx1, isArticleDraftQuery, sql.Named("uuid", id))

  var isArticleDraft bool

  err := row.Scan(&isArticleDraft)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      err = problem.NewNotFound(id, "draft")
    } else {
      slog.Error(err.Error())
    }

    return err
  }

  if !isArticleDraft {
    p := problem.Problem{}
    p.Title("Article already published.")
    p.Status(http.StatusConflict)
    p.Detail("Cannot publish an article that is already published.")
    p.With("article_uuid", id)

    return &p
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  publishArticleDraftQuery := `
  UPDATE "article"
     SET "draft" = FALSE,
         "published_at" = current_timestamp
   WHERE "uuid" = @uuid;`

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, publishArticleDraftQuery, sql.Named("uuid", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "draft")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) Get(ctx context.Context, needle string, hidden, draftsOnly bool) (articles []*model.Article, err error) {
  getArticlesQuery := `
  SELECT "uuid",
         "title",
         "author",
         "slug",
         "read_time",
         "content",
         "draft",
         "pinned",
         "drafted_at",
         "published_at",
         "updated_at",
         "modified_at"
    FROM "article"
   WHERE "draft" IS @drafts_only
     AND CASE WHEN @drafts_only
              THEN "published_at" IS NULL
              ELSE "published_at" IS NOT NULL
               AND "hidden" IS @hidden
               END`

  if "" != needle {
    searchAnnex := ""

    for _, chunk := range strings.Fields(needle) {
      if strings.Contains(chunk, "'") {
        chunk = strings.ReplaceAll(chunk, "'", "''")
      }

      searchAnnex += fmt.Sprintf("\nAND \"title\" LIKE '%%%s%%'", chunk)
    }

    getArticlesQuery += searchAnnex
  }

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getArticlesQuery,
    sql.Named("needle", needle), sql.Named("drafts_only", draftsOnly), sql.Named("hidden", hidden))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  articles = make([]*model.Article, 0)

  for result.Next() {
    var article model.Article

    err = result.Scan(
      &article.UUID,
      &article.Title,
      &article.Author,
      &article.Slug,
      &article.ReadTime,
      &article.Content,
      &article.IsDraft,
      &article.IsPinned,
      &article.DraftedAt,
      &article.PublishedAt,
      &article.UpdatedAt,
      &article.ModifiedAt,
    )

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    articles = append(articles, &article)
  }

  return articles, nil
}

func (r *articlesRepository) GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error) {
  getArticleByUUIDQuery := `
  SELECT "uuid",
         "title",
         "author",
         "slug",
         "read_time",
         "content",
         "draft",
         "pinned",
         "drafted_at",
         "published_at",
         "updated_at",
         "modified_at"
    FROM "article"
   WHERE "uuid" = @uuid
     AND "draft" IS @is_draft
     AND CASE WHEN @is_draft
              THEN "published_at" IS NULL
              ELSE "published_at" IS NOT NULL
               AND "hidden" IS FALSE
               END;`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  row := r.db.QueryRowContext(ctx, getArticleByUUIDQuery, sql.Named("uuid", id), sql.Named("is_draft", isDraft))

  article = new(model.Article)

  err = row.Scan(
    &article.UUID,
    &article.Title,
    &article.Author,
    &article.Slug,
    &article.ReadTime,
    &article.Content,
    &article.IsDraft,
    &article.IsPinned,
    &article.DraftedAt,
    &article.PublishedAt,
    &article.UpdatedAt,
    &article.ModifiedAt,
  )

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      recordType := "article"

      if isDraft {
        recordType = "draft"
      }

      err = problem.NewNotFound(id, recordType)
    } else {
      slog.Error(err.Error())
    }

    return nil, err
  }

  return article, nil
}

func (r *articlesRepository) Amend(ctx context.Context, id string) error {
  articleExistsQuery := `
  SELECT "uuid"
    FROM "article"
   WHERE "uuid" = @uuid
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  err := r.db.QueryRowContext(ctx1, articleExistsQuery, sql.Named("uuid", id)).Scan(&id)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      return problem.NewNotFound(id, "article")
    }

    slog.Error(err.Error())
    return err
  }

  isBeenAmendedQuery := `
    SELECT count (1)
      FROM "article_patch"
     WHERE "article_uuid" = @article_uuid;`

  var isBeenAmended bool

  ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err = r.db.QueryRowContext(ctx2, isBeenAmendedQuery, sql.Named("article_uuid", id)).Scan(&isBeenAmended)
  if nil != err {
    slog.Error(err.Error())
  }

  if isBeenAmended {
    p := problem.Problem{}
    p.Title("Article is currently been amended.")
    p.Detail("Could not amend article because there is an ongoing update.")
    p.Status(http.StatusConflict)
    p.With("article_uuid", id)

    return &p
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  amendArticleQuery := `
  INSERT INTO "article_patch" ("article_uuid",
                               "title",
                               "slug",
                               "content")
                       VALUES (@uuid,
                               NULL,
                               NULL,
                               NULL);`

  ctx, cancel = context.WithTimeout(ctx, 4*time.Second)
  defer cancel()

  _, err = tx.ExecContext(ctx, amendArticleQuery, sql.Named("uuid", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  removeArticleQuery := `
  DELETE FROM "article"
        WHERE "uuid" = @uuid
          AND "draft" IS FALSE
          AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, removeArticleQuery, sql.Named("uuid", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) AddTopic(ctx context.Context, articleID, topicID string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) RemoveTopic(ctx context.Context, articleID, topicID string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) SetHidden(ctx context.Context, id string, hidden bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  setHiddenQuery := `
  UPDATE "article"
     SET "hidden" = @hidden
   WHERE "uuid" = @uuid
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, setHiddenQuery, sql.Named("uuid", id), sql.Named("hidden", hidden))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  setPinnedQuery := `
  UPDATE "article"
     SET "pinned" = @pinned
   WHERE "uuid" = @uuid
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, setPinnedQuery, sql.Named("uuid", id), sql.Named("pinned", pinned))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) Share(ctx context.Context, id string) (link string, err error) {
  assertIsArticleDraftQuery := `
  SELECT count (*)
    FROM "article"
   WHERE "uuid" = $1
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  var isDraft bool

  err = r.db.QueryRowContext(ctx1, assertIsArticleDraftQuery, id).Scan(&isDraft)
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }

  if !isDraft {
    return "", problem.NewNotFound(id, "draft")
  }

  tryToGetCurrentLinkWithExpirationTimeQuery := `
  SELECT "sharable_link",
         "expires_at"
    FROM "article_link"
   WHERE "article_uuid" = $1;`

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  var expiresAt time.Time

  err = r.db.QueryRowContext(ctx, tryToGetCurrentLinkWithExpirationTimeQuery, id).Scan(&link, &expiresAt)
  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(err.Error())
      return "", err
    }
  }

  if "" != link {
    now := time.Now()
    if -1 == expiresAt.Compare(now) || 0 == expiresAt.Compare(now) {
      removeObsoleteLinkQuery := `
      DELETE FROM "article_link"
            WHERE "article_uuid" = $1;`

      ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
      defer cancel()

      _, err = r.db.ExecContext(ctx, removeObsoleteLinkQuery, id)
      if nil != err {
        slog.Error(err.Error())
        return "", err
      }

      p := problem.Problem{}
      p.Status(http.StatusGone)
      p.Title("Invalid shareable link.")
      p.Detail("This shareable link is no longer valid. Try creating a new one.")
      p.With("shareable_link", link)

      return "", &p
    }

    return link, nil
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }

  defer tx.Rollback()

  makeShareableLinkQuery := `
  INSERT INTO "article_link" ("article_uuid", "sharable_link")
                      VALUES (@article_uuid, @sharable_link)
    RETURNING "sharable_link";`

  data := fmt.Sprintf("%s at %s", id, time.Now().String())
  hash := sha1.Sum([]byte(data))

  err = tx.QueryRowContext(ctx, makeShareableLinkQuery,
    sql.Named("article_uuid", id), sql.Named("sharable_link", fmt.Sprintf("/archive/s/%x", hash))).
    Scan(&link)

  if nil != err {
    slog.Error(err.Error())
    return "", nil
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return "", err
  }

  return link, nil
}

func (r *articlesRepository) Discard(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  discardDraftQuery := `
  DELETE
    FROM "article"
   WHERE "uuid" = @uuid
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, discardDraftQuery, sql.Named("uuid", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "draft")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  reviseArticleQuery := `
  UPDATE "article"
     SET "title" = coalesce (nullif (@title, ''), "title"),
         "slug" = coalesce (nullif (@slug, ''), "slug"),
         "content" = coalesce (nullif (@content, ''), "content")
   WHERE "uuid" = @uuid
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  result, err := tx.ExecContext(ctx, reviseArticleQuery,
    sql.Named("uuid", id),
    sql.Named("title", revision.Title),
    sql.Named("slug", revision.Slug),
    sql.Named("content", revision.Content),
  )

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "draft")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *articlesRepository) Release(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) GetPatches(ctx context.Context) (patches []*model.ArticlePatch, err error) {
  getPatchesQuery := `
  SELECT "article_uuid",
         "title",
         "slug",
         "content"
    FROM "article_patch";`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getPatchesQuery)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  defer result.Close()

  patches = make([]*model.ArticlePatch, 0)

  for result.Next() {
    var patch model.ArticlePatch

    err = result.Scan(
      &patch.ArticleUUID,
      &patch.Title,
      &patch.Slug,
      &patch.Content)

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    patches = append(patches, &patch)
  }

  return patches, nil
}
