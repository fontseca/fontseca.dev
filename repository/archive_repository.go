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

// ArchiveRepository is a common API for articles, article drafts
// and article patches.
//
// An article is a piece of writing about a particular subject in my
// website's archive. Naturally, every article has one or more tags
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
type ArchiveRepository interface {
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
  Get(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error)

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

  // AddTag adds a tag to the article. If the tag already
  // exists, it returns an error informing about a conflicting
  // state.
  AddTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error

  // RemoveTag removes a tag from the article. If the article has
  // no tag identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error

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

type archiveRepository struct {
  db *sql.DB
}

func NewArchiveRepository(db *sql.DB) ArchiveRepository {
  return &archiveRepository{db}
}

func (r *archiveRepository) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
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

func (r *archiveRepository) Publish(ctx context.Context, id string) error {
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

func (r *archiveRepository) Get(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error) {
  getArticlesQuery := `
  SELECT "uuid",
         "title",
         "slug",
         "topic",
         "pinned",
         "published_at"
    FROM "article"
   WHERE "draft" IS @drafts_only
     AND CASE WHEN @drafts_only
              THEN "published_at" IS NULL
              ELSE "published_at" IS NOT NULL
               AND "hidden" IS @hidden
               AND CASE WHEN @publication_year <> 0 AND @publication_month <> 0 
                   THEN
                        CAST(strftime('%Y', "published_at") AS INTEGER) = @publication_year AND
                        CAST(strftime('%m', "published_at") AS INTEGER) = @publication_month
                    ELSE TRUE END
               AND CASE WHEN @topic <> ""
                   THEN
                        "topic" = @topic
                   ELSE TRUE END
               END`

  if "" != filter.Search {
    searchAnnex := ""

    for _, chunk := range strings.Fields(filter.Search) {
      if strings.Contains(chunk, "'") {
        chunk = strings.ReplaceAll(chunk, "'", "''")
      }

      searchAnnex += fmt.Sprintf("\nAND \"title\" LIKE '%%%s%%'", chunk)
    }

    getArticlesQuery += searchAnnex
  }

  getArticlesQuery += `
  ORDER BY "pinned" DESC, "published_at" DESC
  LIMIT @rpp
  OFFSET @rpp * (@page - 1);`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var (
    year  = 0
    month = 0
  )

  if nil != filter.Publication {
    year = filter.Publication.Year
    month = int(filter.Publication.Month)
  }

  result, err := r.db.QueryContext(ctx, getArticlesQuery,
    sql.Named("needle", filter.Search),
    sql.Named("drafts_only", draftsOnly),
    sql.Named("hidden", hidden),
    sql.Named("page", filter.Page),
    sql.Named("rpp", filter.RPP),
    sql.Named("topic", filter.Topic),
    sql.Named("publication_year", year),
    sql.Named("publication_month", month),
  )

  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  articles = make([]*transfer.Article, 0)

  for result.Next() {
    var (
      article       transfer.Article
      slug          string
      nullableTopic sql.NullString
    )

    err = result.Scan(
      &article.UUID,
      &article.Title,
      &slug,
      &nullableTopic,
      &article.IsPinned,
      &article.PublishedAt,
    )

    if !draftsOnly {
      var publishedAt time.Time

      if nil != article.PublishedAt {
        publishedAt = *article.PublishedAt
      }

      topic := nullableTopic.String
      year := publishedAt.Year()
      month := int(publishedAt.Month())

      if "" == topic {
        topic = "none"
      }

      // The URL has the form: 'https://fontseca.dev/archive/:topic/:year/:month/:slug'.
      article.URL = fmt.Sprintf("https://fontseca.dev/archive/%s/%d/%d/%s", topic, year, month, slug)
    } else {
      article.URL = "about:blank"
    }

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    articles = append(articles, &article)
  }

  return articles, nil
}

func (r *archiveRepository) GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error) {
  getTagsQuery := `
     SELECT t."id",
            t."name",
            t.created_at,
            t.updated_at
       FROM "article_tag" at
  LEFT JOIN "tag" t
         ON at."tag_id" = t."id"
      WHERE "article_uuid" = @article_uuid;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx1, getTagsQuery, sql.Named("article_uuid", id))
  if err != nil {
    slog.Error(err.Error())
    return nil, err
  }

  defer result.Close()

  tags := make([]*model.Tag, 0)

  for result.Next() {
    var tag model.Tag

    err = result.Scan(
      &tag.ID,
      &tag.Name,
      &tag.CreatedAt,
      &tag.UpdatedAt,
    )

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    tags = append(tags, &tag)
  }

  getArticleByUUIDQuery := `
     SELECT a."uuid",
            a."title",
            a."author",
            a."slug",
            a."read_time",
            a."content",
            a."draft",
            a."pinned",
            a."drafted_at",
            a."published_at",
            a."updated_at",
            a."modified_at",
            t."id",
            t."name",
            t."created_at",
            t."updated_at"
       FROM "article" a
  LEFT JOIN "topic" t
         ON t."id" = a."topic" 
      WHERE "uuid" = @uuid
        AND "draft" IS @is_draft
        AND CASE WHEN @is_draft
                 THEN "published_at" IS NULL
                 ELSE "published_at" IS NOT NULL
                  AND "hidden" IS FALSE
                  END;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  row := r.db.QueryRowContext(ctx, getArticleByUUIDQuery, sql.Named("uuid", id), sql.Named("is_draft", isDraft))

  article = new(model.Article)

  if 0 < len(tags) {
    article.Tags = tags
  }

  var (
    nullableTopicID        sql.NullString
    nullableTopicName      sql.NullString
    nullableTopicCreatedAt sql.Null[time.Time]
    nullableTopicUpdatedAt sql.Null[time.Time]
  )

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
    &nullableTopicID,
    &nullableTopicName,
    &nullableTopicCreatedAt,
    &nullableTopicUpdatedAt,
  )

  if nullableTopicID.Valid {
    article.Topic = new(model.Topic)
    article.Topic.ID = nullableTopicID.String
    article.Topic.Name = nullableTopicName.String

    value, err := nullableTopicCreatedAt.Value()

    if nil != err {
      slog.Error(err.Error())
    } else {
      article.Topic.CreatedAt = value.(time.Time)
    }

    value, err = nullableTopicUpdatedAt.Value()

    if nil != err {
      slog.Error(err.Error())
    } else {
      article.Topic.UpdatedAt = value.(time.Time)
    }
  }

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

func (r *archiveRepository) Amend(ctx context.Context, id string) error {
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

func (r *archiveRepository) Remove(ctx context.Context, id string) error {
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

func (r *archiveRepository) AddTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error {
  var isArticleDraft bool

  if 0 < len(isDraft) {
    isArticleDraft = isDraft[0]
  }

  articleExistsQuery := `
  SELECT count (*)
    FROM "article"
   WHERE "uuid" = @uuid
     AND "draft" IS @is_draft
     AND CASE WHEN @is_draft
           THEN "published_at" IS NULL
           ELSE "published_at" IS NOT NULL
            END;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var articleExists bool

  err := r.db.QueryRowContext(ctx1, articleExistsQuery,
    sql.Named("uuid", articleID),
    sql.Named("is_draft", isArticleDraft)).
    Scan(&articleExists)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(err.Error())
      return err
    }
  }

  if !articleExists {
    if isArticleDraft {
      return problem.NewNotFound(articleID, "draft")
    }

    return problem.NewNotFound(articleID, "article")
  }

  tagExistsQuery := `
  SELECT count (*)
    FROM "tag"
   WHERE "id" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var tagExists bool

  err = r.db.QueryRowContext(ctx, tagExistsQuery, tagID).Scan(&tagExists)
  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(err.Error())
      return err
    }
  }

  if !tagExists {
    return problem.NewNotFound(tagID, "tag")
  }

  tagAlreadyExistsQuery := `
  SELECT count (*)
    FROM "article_tag"
   WHERE "article_uuid" = @article_uuid
     AND "tag_id" = @tag_id;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var tagAlreadyExists bool

  err = r.db.QueryRowContext(ctx, tagAlreadyExistsQuery,
    sql.Named("article_uuid", articleID),
    sql.Named("tag_id", tagID)).
    Scan(&tagAlreadyExists)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(err.Error())
      return err
    }
  }

  if tagAlreadyExists {
    p := problem.Problem{}
    p.Status(http.StatusConflict)
    p.Title("Could not add a tag.")

    detail := "This tag is already added to the current article."

    if isArticleDraft {
      detail = "This tag is already added to the current article draft."
    }

    p.Detail(detail)
    p.With("article_uuid", articleID)
    p.With("tag_id", tagID)

    return &p
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  addTagQuery := `
  INSERT INTO "article_tag" ("article_uuid", "tag_id")
                       VALUES (@article_uuid, @tag_id);`

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, addTagQuery,
    sql.Named("article_uuid", articleID),
    sql.Named("tag_id", tagID))

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    p := problem.Problem{}
    p.Status(http.StatusAccepted)
    p.Title("Could not add a tag.")

    detail := "Could not add tag to this article."

    if isArticleDraft {
      detail = "Could not add tag to this article draft."
    }

    p.Detail(detail)
    p.With("article_uuid", articleID)
    p.With("tag_id", tagID)

    return &p
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *archiveRepository) RemoveTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error {
  var isArticleDraft bool

  if 0 < len(isDraft) {
    isArticleDraft = isDraft[0]
  }

  articleExistsQuery := `
  SELECT count (*)
    FROM "article"
   WHERE "uuid" = @uuid
     AND "draft" IS @is_draft
     AND CASE WHEN @is_draft
           THEN "published_at" IS NULL
           ELSE "published_at" IS NOT NULL
            END;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var articleExists bool

  err := r.db.QueryRowContext(ctx1, articleExistsQuery,
    sql.Named("uuid", articleID),
    sql.Named("is_draft", isArticleDraft)).
    Scan(&articleExists)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(err.Error())
      return err
    }
  }

  if !articleExists {
    if isArticleDraft {
      return problem.NewNotFound(articleID, "draft")
    }

    return problem.NewNotFound(articleID, "article")
  }

  tagExistsQuery := `
  SELECT count (*)
    FROM "tag"
   WHERE "id" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var tagExists bool

  err = r.db.QueryRowContext(ctx, tagExistsQuery, tagID).Scan(&tagExists)
  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(err.Error())
      return err
    }
  }

  if !tagExists {
    return problem.NewNotFound(tagID, "tag")
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  removeTagQuery := `
  DELETE FROM "article_tag"
         WHERE "article_uuid" = @article_uuid
           AND "tag_id" = @tag_id;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, removeTagQuery,
    sql.Named("article_uuid", articleID),
    sql.Named("tag_id", tagID))

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    p := problem.Problem{}
    p.Status(http.StatusAccepted)
    p.Title("Could not remove a tag.")
    p.Detail("This article is no longer attached to this tag.")
    p.With("article_uuid", articleID)
    p.With("tag_id", tagID)

    return &p
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *archiveRepository) SetHidden(ctx context.Context, id string, hidden bool) error {
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

func (r *archiveRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
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

func (r *archiveRepository) Share(ctx context.Context, id string) (link string, err error) {
  assertIsArticlePatchQuery := `
  SELECT count(*)
    FROM "article_patch"
   WHERE "article_uuid" = @article_uuid;`

  var isArticlePatch bool

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err = r.db.QueryRowContext(ctx1, assertIsArticlePatchQuery, sql.Named("article_uuid", id)).Scan(&isArticlePatch)
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }

  if !isArticlePatch {
    assertIsArticleDraftQuery := `
    SELECT count (*)
      FROM "article"
     WHERE "uuid" = $1
       AND "draft" IS TRUE
       AND "published_at" IS NULL;`

    ctx1, cancel = context.WithTimeout(ctx, 2*time.Second)
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

func (r *archiveRepository) Discard(ctx context.Context, id string) error {
  isArticlePatchQuery := `
  SELECT count(*)
    FROM "article_patch"
   WHERE "article_uuid" = @article_uuid;`

  var isArticlePatch bool

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err := r.db.QueryRowContext(ctx1, isArticlePatchQuery, sql.Named("article_uuid", id)).Scan(&isArticlePatch)
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  discardPatchOrDraftQuery := `
  DELETE
    FROM "article"
   WHERE "uuid" = @uuid
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  if isArticlePatch {
    discardPatchOrDraftQuery = `
    DELETE
      FROM "article_patch"
     WHERE "article_uuid" = @uuid;`
  }

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, discardPatchOrDraftQuery, sql.Named("uuid", id))
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    if isArticlePatch {
      return problem.NewNotFound(id, "article patch")
    }

    return problem.NewNotFound(id, "draft")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *archiveRepository) Revise(ctx context.Context, id string, revision *transfer.ArticleUpdate) error {
  isArticlePatchQuery := `
  SELECT count(*)
    FROM "article_patch"
   WHERE "article_uuid" = @article_uuid;`

  var isArticlePatch bool

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err := r.db.QueryRowContext(ctx1, isArticlePatchQuery, sql.Named("article_uuid", id)).Scan(&isArticlePatch)
  if nil != err {
    slog.Error(err.Error())
    return err
  }

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
         "read_time" = CASE WHEN @read_time = "read_time"
                              OR @read_time IS NULL
                              OR @read_time = 0
                            THEN "read_time"
                            ELSE @read_time
                             END,
         "content" = coalesce (nullif (@content, ''), "content")
   WHERE "uuid" = @uuid
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  if isArticlePatch {
    reviseArticleQuery = `
    UPDATE "article_patch"
       SET "title" = coalesce (nullif (@title, ''), "title"),
           "slug" = coalesce (nullif (@slug, ''), "slug"),
           "read_time" = CASE WHEN @read_time = "read_time"
                                OR @read_time IS NULL
                                OR @read_time = 0
                              THEN "read_time"
                              ELSE @read_time
                               END,
           "content" = coalesce (nullif (@content, ''), "content")
     WHERE "article_uuid" = @uuid;`
  }

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, reviseArticleQuery,
    sql.Named("uuid", id),
    sql.Named("title", revision.Title),
    sql.Named("slug", revision.Slug),
    sql.Named("read_time", revision.ReadTime),
    sql.Named("content", revision.Content),
  )

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    if isArticlePatch {
      return problem.NewNotFound(id, "article patch")
    }

    return problem.NewNotFound(id, "draft")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *archiveRepository) Release(ctx context.Context, id string) error {
  getPatchQuery := `
  SELECT "article_uuid",
         "title",
         "slug",
         "read_time",
         "content"
    FROM "article_patch"
   WHERE "article_uuid" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var patch model.ArticlePatch

  err := r.db.QueryRowContext(ctx1, getPatchQuery, id).
    Scan(&patch.ArticleUUID,
      &patch.Title,
      &patch.Slug,
      &patch.ReadTime,
      &patch.Content,
    )

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      return problem.NewNotFound(id, "article patch")
    }

    slog.Error(err.Error())
    return err
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  releasePatchQuery := `
  UPDATE "article"
     SET "title" = coalesce(nullif(@title, ''), "title"),
         "slug" = coalesce(nullif(@slug, ''), "slug"),
         "read_time" = CASE WHEN @read_time = "read_time"
                              OR @read_time IS NULL
                              OR @read_time = 0
                            THEN "read_time"
                            ELSE @read_time
                             END,
         "content" = coalesce(nullif(@content, ''), "content"),
         "modified_at" = current_timestamp,
         "updated_at" = current_timestamp
   WHERE "uuid" = @uuid
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, releasePatchQuery,
    sql.Named("uuid", id),
    sql.Named("title", patch.Title),
    sql.Named("slug", patch.Slug),
    sql.Named("read_time", patch.ReadTime),
    sql.Named("content", patch.Content))

  if nil != err {
    slog.Error(err.Error())
    return nil
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  removePatchQuery := `
  DELETE FROM "article_patch"
        WHERE "article_uuid" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  _, err = tx.ExecContext(ctx1, removePatchQuery, id)
  if nil != err {
    slog.Error(err.Error())
    return nil
  }

  defer tx.Rollback()

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *archiveRepository) GetPatches(ctx context.Context) (patches []*model.ArticlePatch, err error) {
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
