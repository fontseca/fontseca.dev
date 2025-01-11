package repository

import (
  "context"
  "crypto/sha256"
  "database/sql"
  "errors"
  "fmt"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "log/slog"
  "net/http"
  "net/url"
  "os"
  "slices"
  "strconv"
  "strings"
  "sync"
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
type ArchiveRepository struct {
  db                *sql.DB
  publicationsCache []*transfer.Publication
  articleViewsCache articleViewsCache
  done              chan struct{}
  mu                sync.RWMutex
  cleanOnce         sync.Once // for cleaning broken links once per share
}

// visitor is the IP address of an article reader.
type visitor string

// entry is the metadata needed for an article views caching.
type entry struct {
  watchers map[visitor]struct{}
  views    int64
}

// articleViewsCache is the article views cache abstraction.
type articleViewsCache map[string]*entry

func NewArchiveRepository(db *sql.DB) *ArchiveRepository {
  r := &ArchiveRepository{
    db:                db,
    publicationsCache: []*transfer.Publication{},
    articleViewsCache: articleViewsCache{},
    done:              make(chan struct{}),
  }

  go r.cacheWriter()

  return r
}

// cacheWriter is a goroutine that writes articles view cache every midnight.
func (r *ArchiveRepository) cacheWriter() {
  var (
    now           = time.Now()
    midnight      = time.Date(now.Year(), now.Month(), 1+now.Day(), 0, 0, 0, 0, now.Location())
    untilMidnight = midnight.Sub(now)
  )

  <-time.After(untilMidnight)

  var (
    ticker    = time.NewTicker(1) // Since we're at midnight, we need the first tick right now!
    firstTick = true
  )

  defer ticker.Stop()

  for {
    select {
    case <-r.done:
      return
    case <-ticker.C:
      if firstTick {
        ticker.Reset(24 * time.Hour) // Now keep ticking every midnight.
        firstTick = false
      }

      r.writeViewsCache(context.TODO())
    }
  }
}

// Close forces all caches be written.
func (r *ArchiveRepository) Close(ctx context.Context) {
  close(r.done)
  r.writeViewsCache(ctx)
}

// closed checks if the repository has been closed, that is the
// Close method was invoked.
func (r *ArchiveRepository) closed() bool {
  select {
  default:
    return false
  case <-r.done:
    return true
  }
}

// writeViewsCache is a goroutine that writes the cached article views to the database;
// it should be run every midnight.
func (r *ArchiveRepository) writeViewsCache(ctx context.Context) {
  r.mu.RLock()

  if 0 == len(r.articleViewsCache) {
    r.mu.RUnlock()
    return
  }

  r.mu.RUnlock()

  slog.Info("writing article views cache to database")

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return
  }

  defer tx.Rollback()

  writeViewsQuery := `
  UPDATE "archive"."article"
     SET "views" = "views" + $2::INTEGER
   WHERE "uuid" = $1;`

  writers := sync.WaitGroup{}
  ctx, cancel := context.WithTimeout(ctx, time.Minute)
  defer cancel()

  for article, metadata := range r.articleViewsCache {
    writers.Add(1)

    go func() {
      r.mu.RLock()
      defer writers.Done()
      defer r.mu.RUnlock()

      slog.Info("writing cached view(s)",
        slog.Int64("count", metadata.views),
        slog.String("article_uuid", article))

      _, err := tx.ExecContext(ctx, writeViewsQuery,
        article,
        metadata.views)

      if nil != err {
        slog.Error(getErrMsg(err), slog.String("article_uuid", article))
      }
    }()
  }

  writers.Wait()

  if !r.closed() {
    slog.Info("resetting article views cache")

    r.articleViewsCache = articleViewsCache{}
  }

  if err := tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
  }
}

// views returns the cached views of the given article.
func (r *ArchiveRepository) views(article string) int64 {
  r.mu.RLock()
  defer r.mu.RUnlock()

  metadata, cached := r.articleViewsCache[article]

  if cached {
    return metadata.views
  }

  return 0
}

// VisitorKey is the key that I use to get the remote IP
// address which represents a visitor.
const VisitorKey string = "visitor-ip"

// incrementViews increments the views counter of the given article,
func (r *ArchiveRepository) incrementViews(ctx context.Context, article string) {
  r.mu.Lock()
  defer r.mu.Unlock()

  var (
    value         = ctx.Value(VisitorKey)
    ip    visitor = ""
  )

  if "" != value {
    ip = visitor(value.(string))
  }

  if _, isCached := r.articleViewsCache[article]; !isCached {
    slog.Info("caching article views count", slog.String("article_uuid", article))

    watchers := make(map[visitor]struct{})
    watchers[ip] = struct{}{}

    r.articleViewsCache[article] = &entry{
      watchers: watchers,
      views:    1,
    }

    return
  }

  if "" == ip {
    r.articleViewsCache[article].views++
    return
  }

  if _, hasViewed := r.articleViewsCache[article].watchers[ip]; !hasViewed {
    r.articleViewsCache[article].watchers[ip] = struct{}{}
    r.articleViewsCache[article].views++
  }
}

// cleanBrokenLinks is a goroutine that cleans up shareable links that
// are no longer valid because have already expired.
func (r *ArchiveRepository) cleanBrokenLinks() {
  r.cleanOnce.Do(func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    checkThereAreBrokenLinksQuery := `
    SELECT count (*)
      FROM "archive"."article_link"
     WHERE "expires_at" <= current_timestamp;`

    nbroken := 0

    err := r.db.QueryRowContext(ctx, checkThereAreBrokenLinksQuery).Scan(&nbroken)

    if nil != err {
      slog.Error(getErrMsg(err))
    }

    if 0 == nbroken {
      return
    }

    attempts := 1
    anew := func() { attempts++ }
    done := func() bool { return attempts > 3 }

  again:
    slog.Info("trying to clear all broken shareable links",
      slog.Int("broken_links", nbroken),
      slog.Int("attempt", attempts))

    tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})

    if nil != err {
      slog.Error(getErrMsg(err))

      anew()
      if done() {
        return
      }

      goto again
    }

    defer tx.Rollback()

    removeBrokenLinksQuery := `
    DELETE FROM "archive"."article_link"
          WHERE "expires_at" <= current_timestamp;`

    ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    _, err = tx.ExecContext(ctx, removeBrokenLinksQuery)

    if nil != err {
      slog.Error(getErrMsg(err))

      anew()
      if done() {
        return
      }

      goto again
    }

    if err = tx.Commit(); nil != err {
      slog.Error(getErrMsg(err))
    }
  })
}

// Draft starts the creation process of an article. It returns the
// UUID of the draft that was created.
//
// To draft an article, only its title is required, other fields
// are completely optional and can be added in an eventual revision.
func (r *ArchiveRepository) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
  slog.Info("drafting new article", slog.String("title", creation.Title))

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return uuid.Nil.String(), err
  }

  defer tx.Rollback()

  draftArticleQuery := `
  INSERT INTO "archive"."article" ("title",
                                   "author",
                                   "slug",
                                   "read_time",
                                   "content",
                                   "summary",
                                   "cover_url",
                                   "cover_caption")
                 VALUES ($1,
                         'fontseca.dev',
                         $2,
                         $3,
                         coalesce(nullif($4, ''), 'no content'),
                         coalesce(nullif($5, ''), 'no summary'),
                         coalesce(nullif($6, ''), 'about:blank'),
                         nullif($7, ''))
              RETURNING "uuid";`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result := tx.QueryRowContext(ctx, draftArticleQuery,
    &creation.Title,
    &creation.Slug,
    &creation.ReadTime,
    &creation.Content,
    &creation.Summary,
    &creation.CoverURL,
    &creation.CoverCap,
  )

  if err = result.Scan(&id); nil != err {
    if strings.Contains(err.Error(), `duplicate key value violates unique constraint "article_slug_key"`) {
      var p problem.Problem
      p.Type(problem.TypeDuplicateKey)
      p.Status(http.StatusConflict)
      p.Title("Duplicate draft title.")
      p.Detail("The provided article draft title is already registered. Try using a different one.")
      p.With("draft_title", creation.Title)
      return "", &p
    }
    slog.Error(getErrMsg(err))
    return uuid.Nil.String(), err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return uuid.Nil.String(), err
  }

  return id, nil
}

// Publish makes a draft publicly available.
//
// Invoking Publish on an already published article or a patch has
// no effect.
func (r *ArchiveRepository) Publish(ctx context.Context, id string) error {
  isArticleDraftQuery := `
  SELECT "draft" IS TRUE
     AND "published_at" IS NULL
    FROM "archive"."article"
   WHERE "uuid" = $1;`

  ctx1, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()

  row := r.db.QueryRowContext(ctx1, isArticleDraftQuery, id)

  var isArticleDraft bool

  err := row.Scan(&isArticleDraft)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      err = problem.NewNotFound(id, "draft")
    } else {
      slog.Error(getErrMsg(err))
    }

    return err
  }

  if !isArticleDraft {
    p := problem.Problem{}
    p.Type(problem.TypeActionAlreadyCompleted)
    p.Title("Article already published.")
    p.Status(http.StatusConflict)
    p.Detail("Cannot publish an article that is already published.")
    p.With("article_uuid", id)

    return &p
  }

  var hasTopic bool

  hasTopicQuery := `
  SELECT count (1)
    FROM "archive"."article"
   WHERE "uuid" = $1
     AND "draft" IS TRUE
     AND "published_at" IS NULL
     AND "topic" IS NOT NULL
     AND "topic" <> '';`

  err = r.db.QueryRowContext(ctx, hasTopicQuery, id).Scan(&hasTopic)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if !hasTopic {
    p := &problem.Problem{}
    p.Type(problem.TypeActionRefused)
    p.Status(http.StatusBadRequest)
    p.Title("Could not publish draft.")
    p.Detail("Cannot publish a draft without making it belong to a topic first.")
    p.With("draft_uuid", id)
    return p
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  publishArticleDraftQuery := `
  UPDATE "archive"."article"
     SET "draft" = FALSE,
         "published_at" = current_timestamp
   WHERE "uuid" = $1;`

  slog.Info("publishing draft", slog.String("uuid", id))

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, publishArticleDraftQuery, id)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "draft")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  r.setPublicationsCache(ctx)

  return nil
}

// SetSlug changes the slug of a published article.
func (r *ArchiveRepository) SetSlug(ctx context.Context, id, slug string) error {
  slog.Info("changing article slug", slog.String("article_uuid", id), slog.String("new_slug", slug))

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  setSlugQuery := `
  UPDATE "archive"."article"
     SET "slug" = $2
   WHERE "uuid" = $1
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, setSlugQuery, id, slug)

  if nil != err && !errors.Is(sql.ErrNoRows, err) {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err := tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

func (r *ArchiveRepository) setPublicationsCache(ctx context.Context) {
  r.publicationsCache = nil
  r.publicationsCache, _ = r.Publications(ctx)
}

// Publications retrieves a list of distinct months during which articles have been published.
func (r *ArchiveRepository) Publications(ctx context.Context) (publications []*transfer.Publication, err error) {
  if 0 < len(r.publicationsCache) {
    return r.publicationsCache, nil
  }

  getPublicationsQuery := `
  SELECT extract(MONTH FROM "published_at")::INTEGER AS "month",
         extract(YEAR FROM "published_at")::INTEGER AS "year"
    FROM "archive"."article"
   WHERE "draft" IS FALSE
     AND "published_at" IS NOT NULL
     AND "hidden" IS FALSE
     GROUP BY "month", "year"
     ORDER BY "year" DESC, "month" DESC;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getPublicationsQuery)

  if nil != err {
    slog.Error(getErrMsg(err))
    return nil, err
  }

  defer result.Close()

  publications = make([]*transfer.Publication, 0)

  for result.Next() {
    var publication transfer.Publication

    err = result.Scan(
      &publication.Month,
      &publication.Year,
    )

    if nil != err {
      slog.Error(getErrMsg(err))
      return nil, err
    }

    publications = append(publications, &publication)
  }

  r.publicationsCache = publications

  return publications, nil
}

// List retrieves all the articles that are either hidden or not. If
// draftsOnly is true, then only retrieves all the ongoing drafts.
//
// If filter.Search is a non-empty string, then List behaves like a search
// function over non-hidden articles, so it attempts to find and
// amass every article whose title contains any of the keywords
// (if more than one) in the search string.
func (r *ArchiveRepository) List(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error) {
  query := strings.Builder{}
  query.WriteString(`
  SELECT "uuid",
         "title",
         "slug",
         "topic",
         "pinned",
         "published_at"
    FROM "archive"."article" a`)

  if "" != filter.Tag {
    query.WriteString(`
    INNER JOIN "archive"."article_tag" t ON t."article_uuid" = a."uuid"`)
  }

  query.WriteString(`
   WHERE "draft" = $1
     AND CASE WHEN $1 = TRUE
              THEN "published_at" IS NULL
              ELSE "published_at" IS NOT NULL
               AND "hidden" = $2
               AND CASE WHEN $6 <> 0 AND $7 <> 0 
                   THEN
                        extract(YEAR FROM "published_at")::INTEGER = $6 AND
                        extract(MONTH FROM "published_at")::INTEGER = $7
                    ELSE TRUE END
               AND CASE WHEN $5 <> ''
                   THEN
                        "topic" = $5
                   ELSE TRUE END
               END`)

  if "" != filter.Tag {
    query.WriteString(`
               AND t."tag_id" = $8`)
  } else {
    query.WriteString(`
               AND length($8) >= 0`)
  }

  if "" != filter.Search {
    for _, chunk := range strings.Fields(filter.Search) {
      if strings.Contains(chunk, "'") {
        chunk = strings.ReplaceAll(chunk, "'", "''")
      }

      query.WriteString("\nAND lower(\"title\") LIKE '%" + strings.ToLower(chunk) + "%'")
    }
  }

  query.WriteString(`
  ORDER BY "pinned" DESC, "published_at" DESC
  LIMIT $4
  OFFSET $4 * ($3 - 1);`)

  var (
    year  = 0
    month = 0
  )

  if nil != filter.Publication {
    year = filter.Publication.Year
    month = int(filter.Publication.Month)
  }

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, query.String(),
    draftsOnly,
    hidden,
    filter.Page,
    filter.RPP,
    filter.Topic,
    year,
    month,
    filter.Tag,
  )

  if nil != err {
    slog.Error(getErrMsg(err))
    return nil, err
  }

  defer result.Close()

  URLBase := "/"

  if gin.ReleaseMode == strings.TrimSpace(os.Getenv("SERVER_MODE")) {
    value := ctx.Value(gin.ContextKey)
    if nil != value {
      c := value.(*gin.Context)

      if nil != c {
        schema := "http"

        if nil != c.Request.TLS {
          schema = "https"
        }

        URLBase = schema + "://" + c.Request.Host
      }
    }
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

    topic := nullableTopic.String

    if "" == topic {
      article.Topic = nil
    } else {
      // The topic URL has the form: '.../archive/:topic'.
      topicURL, err := url.JoinPath(URLBase, "archive", topic)

      if nil == err {
        article.Topic = &struct {
          ID  string `json:"id"`
          URL string `json:"url"`
        }{
          ID:  topic,
          URL: topicURL,
        }
      } else {
        slog.Error(err.Error())
      }
    }

    if !draftsOnly && nil != article.Topic {
      var publishedAt time.Time

      if nil != article.PublishedAt {
        publishedAt = *article.PublishedAt
      }

      year := publishedAt.Year()
      month := int(publishedAt.Month())

      // The URL has the form: '.../archive/:topic/:year/:month/:slug'.
      u, err := url.JoinPath(
        URLBase,
        "archive",
        article.Topic.ID,
        strconv.Itoa(year),
        strconv.Itoa(month),
        slug)

      if nil == err {
        article.URL = u
      } else {
        slog.Error(err.Error())
      }
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

// Get retrieves one published article by the URL '/archive/:topic/:year/:month/:slug'.
func (r *ArchiveRepository) Get(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error) {
  requestArticleUUIDQuery := `
  SELECT "uuid"
    FROM "archive"."article"
   WHERE "draft" IS FALSE
     AND "published_at" IS NOT NULL
     AND "hidden" IS FALSE
     AND "topic" = $1
     AND extract(YEAR FROM "published_at")::INTEGER = $2
     AND extract(MONTH FROM "published_at")::INTEGER = $3
     AND "slug" = $4;`

  var (
    year  = 0
    month = 0
  )

  if nil != request.Publication {
    year = request.Publication.Year
    month = int(request.Publication.Month)
  }

  var id string

  ctx1, cancel1 := context.WithTimeout(ctx, 5*time.Second)
  defer cancel1()

  err = r.db.QueryRowContext(ctx1, requestArticleUUIDQuery,
    request.Topic,
    year,
    month,
    request.Slug,
  ).Scan(&id)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
    }

    return nil, err
  }

  if "" == id {
    return nil, problem.NewNotFound(id, "article") // TODO: Do not return this kind of problem.
  }

  go r.incrementViews(ctx, id)

  return r.GetByID(ctx, id, false)
}

// GetByLink retrieves a draft by its shareable link.
func (r *ArchiveRepository) GetByLink(ctx context.Context, link string) (article *model.Article, err error) {
  getByLinkQuery := `
  SELECT "article_uuid",
         "expires_at"
    FROM "archive"."article_link"
   WHERE "shareable_link" = $1;`

  var (
    id           string
    expiresAtStr string
  )

  ctx1, cancel1 := context.WithTimeout(ctx, 5*time.Second)
  defer cancel1()

  err = r.db.QueryRowContext(ctx1, getByLinkQuery, link).Scan(&id, &expiresAtStr)

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      go r.cleanBrokenLinks() // TODO: Improve this side effect by using a background worker.

      p := &problem.Problem{}
      p.Status(http.StatusGone)
      p.Title("Orphan shareable link.")
      p.Detail("No article was found referenced by this shareable link; it might have been either removed or blocked.")
      p.With("shareable_link", link)
      err = p
    } else {
      slog.Error(getErrMsg(err))
    }

    return nil, err
  }

  expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)

  if nil != err {
    slog.Error(err.Error())
    expiresAt = time.Time{}
  }

  now := time.Now()
  if -1 == expiresAt.Compare(now) || 0 == expiresAt.Compare(now) {
    go r.cleanBrokenLinks() // TODO: Improve this side effect by using a background worker.

    p := &problem.Problem{}
    p.Status(http.StatusGone)
    p.Title("Broken shareable link.")
    p.Detail("This shareable link is no longer valid because it has expired.")
    p.With("shareable_link", link)

    return nil, p
  }

  go r.incrementViews(ctx, id)

  return r.GetByID(ctx, id, true)
}

// GetByID retrieves one article (or article draft) by its UUID.
func (r *ArchiveRepository) GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error) {
  getTagsQuery := `
     SELECT t."id",
            t."name",
            t.created_at,
            t.updated_at
       FROM "archive"."article_tag" at
  LEFT JOIN "archive"."tag" t
         ON at."tag_id" = t."id"
      WHERE "article_uuid" = $1;`

  ctx1, cancel1 := context.WithTimeout(ctx, 5*time.Second)
  defer cancel1()

  result, err := r.db.QueryContext(ctx1, getTagsQuery, id)
  if err != nil {
    slog.Error(getErrMsg(err))
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
      slog.Error(getErrMsg(err))
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
            a."views",
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
       FROM "archive"."article" a
  LEFT JOIN "archive"."topic" t
         ON t."id" = a."topic" 
      WHERE "uuid" = $1
        AND "draft" = $2
        AND CASE WHEN $2 = TRUE
                 THEN "published_at" IS NULL
                 ELSE "published_at" IS NOT NULL
                  AND "hidden" IS FALSE
                  END;`

  ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
  defer cancel2()

  article = new(model.Article)

  if 0 < len(tags) {
    slices.SortFunc(tags, func(a, b *model.Tag) int {
      return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
    })

    article.Tags = tags
  }

  var (
    nullableTopicID        sql.NullString
    nullableTopicName      sql.NullString
    nullableTopicCreatedAt sql.Null[time.Time]
    nullableTopicUpdatedAt sql.Null[time.Time]
  )

  err = r.db.QueryRowContext(ctx2, getArticleByUUIDQuery, id, isDraft).Scan(
    &article.UUID,
    &article.Title,
    &article.Author,
    &article.Slug,
    &article.ReadTime,
    &article.Views,
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

  article.Views += r.views(article.UUID.String())

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
      slog.Error(getErrMsg(err))
    }

    return nil, err
  }

  return article, nil
}

// Amend starts the process to update an article. To amend the article,
// a public copy of it is kept available to everyone while a patch
// is created to store any revision made to the article.
//
// If the article is still a draft, or it's already being amended,
// any call to this method has no effect.
func (r *ArchiveRepository) Amend(ctx context.Context, id string) error {
  articleExistsQuery := `
  SELECT "uuid"
    FROM "archive"."article"
   WHERE "uuid" = $1
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  err := r.db.QueryRowContext(ctx1, articleExistsQuery, id).Scan(&id)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      return problem.NewNotFound(id, "article")
    }

    slog.Error(getErrMsg(err))
    return err
  }

  isBeenAmendedQuery := `
    SELECT count (1)
      FROM "archive"."article_patch"
     WHERE "article_uuid" = $1;`

  var isBeenAmended bool

  ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err = r.db.QueryRowContext(ctx2, isBeenAmendedQuery, id).Scan(&isBeenAmended)
  if nil != err {
    slog.Error(getErrMsg(err))
  }

  if isBeenAmended {
    p := problem.Problem{}
    p.Type(problem.TypeActionRefused)
    p.Title("Article is currently been amended.")
    p.Detail("Could not amend article because there is an ongoing update.")
    p.Status(http.StatusConflict)
    p.With("article_uuid", id)

    return &p
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  amendArticleQuery := `
  INSERT INTO  "archive"."article_patch"
                              ("article_uuid",
                               "title",
                               "slug",
                               "content")
                       VALUES ($1,
                               NULL,
                               NULL,
                               NULL);`

  slog.Info("starting article amendment", slog.String("article_uuid", id))

  ctx, cancel = context.WithTimeout(ctx, 4*time.Second)
  defer cancel()

  _, err = tx.ExecContext(ctx, amendArticleQuery, id)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Remove completely removes an article and any patch it currently
// has from the database. If the article is a draft, calling Remove
// has no effect on it whatsoever.
//
// If you want to remove a draft, use Discard instead.
func (r *ArchiveRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  removeArticleQuery := `
  DELETE FROM "archive"."article"
        WHERE "uuid" = $1
          AND "draft" IS FALSE
          AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, removeArticleQuery, id)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  r.setPublicationsCache(ctx)

  return nil
}

// AddTag adds a tag to the article. If the tag already
// exists, it returns an error informing about a conflicting
// state.
func (r *ArchiveRepository) AddTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error {
  var isArticleDraft bool

  if 0 < len(isDraft) {
    isArticleDraft = isDraft[0]
  }

  articleExistsQuery := `
  SELECT count (*)
    FROM "archive"."article"
   WHERE "uuid" = $1
     AND "draft" = $2
     AND CASE WHEN $2 = TRUE
           THEN "published_at" IS NULL
           ELSE "published_at" IS NOT NULL
            END;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var articleExists bool

  err := r.db.QueryRowContext(ctx1, articleExistsQuery,
    articleID,
    isArticleDraft).
    Scan(&articleExists)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
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
    FROM "archive"."tag"
   WHERE "id" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var tagExists bool

  err = r.db.QueryRowContext(ctx, tagExistsQuery, tagID).Scan(&tagExists)
  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
      return err
    }
  }

  if !tagExists {
    return problem.NewNotFound(tagID, "tag")
  }

  tagAlreadyExistsQuery := `
  SELECT count (*)
    FROM "archive"."article_tag"
   WHERE "article_uuid" = $1
     AND "tag_id" = $2;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var tagAlreadyExists bool

  err = r.db.QueryRowContext(ctx, tagAlreadyExistsQuery,
    articleID,
    tagID).
    Scan(&tagAlreadyExists)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
      return err
    }
  }

  if tagAlreadyExists {
    p := problem.Problem{}
    p.Type(problem.TypeDuplicateKey)
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
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  addTagQuery := `
  INSERT INTO "archive"."article_tag" ("article_uuid", "tag_id")
                       VALUES ($1, $2);`

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, addTagQuery,
    articleID,
    tagID)

  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// RemoveTag removes a tag from the article. If the article has
// no tag identified by its UUID, it returns an error indication
// a not found state.
func (r *ArchiveRepository) RemoveTag(ctx context.Context, articleID, tagID string, isDraft ...bool) error {
  var isArticleDraft bool

  if 0 < len(isDraft) {
    isArticleDraft = isDraft[0]
  }

  articleExistsQuery := `
  SELECT count (*)
    FROM "archive"."article"
   WHERE "uuid" = $1
     AND "draft" = $2
     AND CASE WHEN $2 = TRUE
           THEN "published_at" IS NULL
           ELSE "published_at" IS NOT NULL
            END;`

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var articleExists bool

  err := r.db.QueryRowContext(ctx1, articleExistsQuery,
    articleID,
    isArticleDraft).
    Scan(&articleExists)

  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
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
    FROM "archive"."tag"
   WHERE "id" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var tagExists bool

  err = r.db.QueryRowContext(ctx, tagExistsQuery, tagID).Scan(&tagExists)
  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
      return err
    }
  }

  if !tagExists {
    return problem.NewNotFound(tagID, "tag")
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  removeTagQuery := `
  DELETE FROM "archive"."article_tag"
         WHERE "article_uuid" = $1
           AND "tag_id" = $2;`

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, removeTagQuery,
    articleID,
    tagID)

  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// SetHidden hides or shows an article depending on the value of hidden.
func (r *ArchiveRepository) SetHidden(ctx context.Context, id string, hidden bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  setHiddenQuery := `
  UPDATE "archive"."article"
     SET "hidden" = $2
   WHERE "uuid" = $1
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, setHiddenQuery, id, hidden)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  r.setPublicationsCache(ctx)

  return nil
}

// SetPinned pins or unpins an article depending on the value of pinned.
func (r *ArchiveRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  setPinnedQuery := `
  UPDATE "archive"."article"
     SET "pinned" = $2
   WHERE "uuid" = $1
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, setPinnedQuery, id, pinned)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Share creates a shareable link for a draft or a patch. Only users
// with that link can see the progress and provide feedback.
//
// A shareable link does not make an article public. This link will
// eventually expire after a certain amount of time.
func (r *ArchiveRepository) Share(ctx context.Context, id string) (link string, err error) {
  defer func() {
    if nil == err {
      r.mu.Lock()
      r.cleanOnce = sync.Once{}
      r.mu.Unlock()
    }
  }()

  assertIsArticlePatchQuery := `
  SELECT count(*)
    FROM "archive"."article_patch"
   WHERE "article_uuid" = $1;`

  var isArticlePatch bool

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err = r.db.QueryRowContext(ctx1, assertIsArticlePatchQuery, id).Scan(&isArticlePatch)
  if nil != err {
    slog.Error(getErrMsg(err))
    return "", err
  }

  if !isArticlePatch {
    assertIsArticleDraftQuery := `
    SELECT count (*)
      FROM "archive"."article"
     WHERE "uuid" = $1
       AND "draft" IS TRUE
       AND "published_at" IS NULL;`

    ctx1, cancel = context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    var isDraft bool

    err = r.db.QueryRowContext(ctx1, assertIsArticleDraftQuery, id).Scan(&isDraft)
    if nil != err {
      slog.Error(getErrMsg(err))
      return "", err
    }

    if !isDraft {
      return "", problem.NewNotFound(id, "draft")
    }
  }

  tryToGetCurrentLinkWithExpirationTimeQuery := `
  SELECT "shareable_link",
         "expires_at"
    FROM "archive"."article_link"
   WHERE "article_uuid" = $1;`

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  var expiresAt time.Time

  err = r.db.QueryRowContext(ctx, tryToGetCurrentLinkWithExpirationTimeQuery, id).Scan(&link, &expiresAt)
  if nil != err {
    if !errors.Is(err, sql.ErrNoRows) {
      slog.Error(getErrMsg(err))
      return "", err
    }
  }

  if "" != link {
    now := time.Now()
    if -1 == expiresAt.Compare(now) || 0 == expiresAt.Compare(now) { // link has expired
      removeObsoleteLinkQuery := `
      DELETE FROM "archive"."article_link"
            WHERE "article_uuid" = $1;`

      ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
      defer cancel()

      _, err = r.db.ExecContext(ctx, removeObsoleteLinkQuery, id)
      if nil != err {
        slog.Error(getErrMsg(err))
        return "", err
      }

      slog.Info("shareable link expired, generating a new one",
        slog.String("elapsed", now.Sub(expiresAt).String()),
        slog.String("article_uuid", id))
    } else {
      return link, nil
    }
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(getErrMsg(err))
    return "", err
  }

  defer tx.Rollback()

  makeShareableLinkQuery := `
  INSERT INTO "archive"."article_link" ("article_uuid", "shareable_link")
                      VALUES ($1, $2)
    RETURNING "shareable_link";`

  data := fmt.Sprintf("%s at %s", id, time.Now().String())
  hash := sha256.Sum256([]byte(data))

  ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
  defer cancel2()

  err = tx.QueryRowContext(ctx2, makeShareableLinkQuery,
    id,
    fmt.Sprintf("/archive/sharing/%x", hash)).
    Scan(&link)

  if nil != err {
    slog.Error(getErrMsg(err))
    return "", nil
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return "", err
  }

  return link, nil
}

// Discard completely drops a draft; otherwise if called on a patch
// it discards it but keeps the original article.
//
// This method has no effect on an article.
func (r *ArchiveRepository) Discard(ctx context.Context, id string) error {
  isArticlePatchQuery := `
  SELECT count(*)
    FROM "archive"."article_patch"
   WHERE "article_uuid" = $1;`

  var isArticlePatch bool

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err := r.db.QueryRowContext(ctx1, isArticlePatchQuery, id).Scan(&isArticlePatch)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  discardPatchOrDraftQuery := `
  DELETE
    FROM "archive"."article"
   WHERE "uuid" = $1
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  if isArticlePatch {
    discardPatchOrDraftQuery = `
    DELETE
      FROM "archive"."article_patch"
     WHERE "article_uuid" = $1;`

    slog.Info("discarding amendment", slog.String("article_uuid", id))
  } else {
    slog.Info("discarding drafted article", slog.String("draft_uuid", id))
  }

  ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, discardPatchOrDraftQuery, id)
  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Revise adds a correction or inclusion to a draft or patch in order
// to correct or improve it.
func (r *ArchiveRepository) Revise(ctx context.Context, id string, revision *transfer.ArticleRevision) error {
  isArticlePatchQuery := `
  SELECT count(*)
    FROM "archive"."article_patch"
   WHERE "article_uuid" = $1;`

  var isArticlePatch bool

  ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err := r.db.QueryRowContext(ctx1, isArticlePatchQuery, id).Scan(&isArticlePatch)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  if "" != revision.Topic {
    exists := false
    topicExistsQuery := `
    SELECT count (1)
      FROM "archive"."topic"
     WHERE "id" = $1;`

    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    err := r.db.QueryRowContext(ctx, topicExistsQuery, revision.Topic).Scan(&exists)

    if nil != err {
      slog.Error(getErrMsg(err))
      return err
    }

    if !exists {
      return problem.NewNotFound(revision.Topic, "topic")
    }
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  reviseArticleQuery := `
  UPDATE "archive"."article"
     SET "title" = coalesce (nullif ($2, ''), "title"),
         "slug" = coalesce (nullif ($3, ''), "slug"),
         "topic" = coalesce (nullif ($4, ''), "topic"),
         "read_time" = CASE WHEN $5 = "read_time"
                              OR $5 IS NULL
                              OR $5 = 0
                            THEN "read_time"
                            ELSE $5
                             END,
         "content" = coalesce (nullif ($6, ''), "content")
   WHERE "uuid" = $1
     AND "draft" IS TRUE
     AND "published_at" IS NULL;`

  if isArticlePatch {
    reviseArticleQuery = `
    UPDATE "archive"."article_patch"
       SET "title" = coalesce (nullif ($2, ''), "title"),
           "slug" = coalesce (nullif ($3, ''), "slug"),
           "topic" = coalesce (nullif ($4, ''), "topic"),
           "read_time" = CASE WHEN $5 = "read_time"
                                OR $5 IS NULL
                                OR $5 = 0
                              THEN "read_time"
                              ELSE $5
                               END,
           "content" = coalesce (nullif ($6, ''), "content")
     WHERE "article_uuid" = $1;`
  }

  ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, reviseArticleQuery,
    id,
    revision.Title,
    revision.Slug,
    revision.Topic,
    revision.ReadTime,
    revision.Content,
  )

  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// Release merges patch into the original article and published the
// update immediately after merging.
//
// This method works only for patches.
func (r *ArchiveRepository) Release(ctx context.Context, id string) error {
  getPatchQuery := `
  SELECT "article_uuid",
         "title",
         "slug",
         "topic",
         "read_time",
         "content"
    FROM "archive"."article_patch"
   WHERE "article_uuid" = $1;`

  slog.Info("releasing article amendment", slog.String("article_uuid", id))

  ctx1, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var patch model.ArticlePatch

  err := r.db.QueryRowContext(ctx1, getPatchQuery, id).
    Scan(&patch.ArticleUUID,
      &patch.Title,
      &patch.Slug,
      &patch.TopicID,
      &patch.ReadTime,
      &patch.Content,
    )

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      return problem.NewNotFound(id, "article patch")
    }

    slog.Error(getErrMsg(err))
    return err
  }

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  releasePatchQuery := `
  UPDATE "archive"."article"
     SET "title" = coalesce(nullif($2, ''), "title"),
         "slug" = coalesce(nullif($3, ''), "slug"),
         "topic" = coalesce(nullif($4, ''), "topic"),
         "read_time" = CASE WHEN $5 = "read_time"
                              OR $5 IS NULL
                              OR $5 = 0
                            THEN "read_time"
                            ELSE $5
                             END,
         "content" = coalesce(nullif($6, ''), "content"),
         "modified_at" = current_timestamp,
         "updated_at" = current_timestamp
   WHERE "uuid" = $1
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx1, releasePatchQuery,
    id,
    patch.Title,
    patch.Slug,
    patch.TopicID,
    patch.ReadTime,
    patch.Content)

  if nil != err {
    slog.Error(getErrMsg(err))
    return nil
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  removePatchQuery := `
  DELETE FROM "archive"."article_patch"
        WHERE "article_uuid" = $1;`

  ctx1, cancel = context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  _, err = tx.ExecContext(ctx1, removePatchQuery, id)
  if nil != err {
    slog.Error(getErrMsg(err))
    return nil
  }

  defer tx.Rollback()

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// ListPatches retrieves all the ongoing patches of every article.
func (r *ArchiveRepository) ListPatches(ctx context.Context) (patches []*model.ArticlePatch, err error) {
  getPatchesQuery := `
  SELECT "article_uuid",
         "title",
         "slug",
         "topic",
         "content"
    FROM "archive"."article_patch";`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getPatchesQuery)
  if nil != err {
    slog.Error(getErrMsg(err))
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
      &patch.TopicID,
      &patch.Content)

    if nil != err {
      slog.Error(getErrMsg(err))
      return nil, err
    }

    patches = append(patches, &patch)
  }

  return patches, nil
}
