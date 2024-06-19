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
  "github.com/gin-gonic/gin"
  "github.com/google/uuid"
  "log/slog"
  "net/http"
  "net/url"
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

  // SetSlug changes the slug of a published article.
  SetSlug(ctx context.Context, id, slug string) error

  // Publications retrieves a list of distinct months during which articles have been published.
  Publications(ctx context.Context) (publications []*transfer.Publication, err error)

  // Get retrieves all the articles that are either hidden or not. If
  // draftsOnly is true, then only retrieves all the ongoing drafts.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over non-hidden articles, so it attempts to find and
  // amass every article whose title contains any of the keywords
  // (if more than one) in needle.
  Get(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error)

  // GetOne retrieves one published article by the URL '/archive/:topic/:year/:month/:slug'.
  GetOne(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error)

  // GetByLink retrieves a draft by its shareable link.
  GetByLink(ctx context.Context, link string) (article *model.Article, err error)

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
  Revise(ctx context.Context, id string, revision *transfer.ArticleRevision) error

  // Release merges patch into the original article and published the
  // update immediately after merging.
  //
  // This method works only for patches.
  Release(ctx context.Context, id string) error

  // GetPatches retrieves all the ongoing patches of every article.
  GetPatches(ctx context.Context) (patches []*model.ArticlePatch, err error)

  // Close forces all caches be written.
  Close()
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

type archiveRepository struct {
  db                *sql.DB
  publicationsCache []*transfer.Publication
  articleViewsCache articleViewsCache
  done              chan struct{}
  mu                sync.RWMutex
  cleanOnce         sync.Once // for cleaning broken links once per share
}

func NewArchiveRepository(db *sql.DB) ArchiveRepository {
  r := &archiveRepository{
    db:                db,
    publicationsCache: []*transfer.Publication{},
    articleViewsCache: articleViewsCache{},
    done:              make(chan struct{}),
  }

  go r.cacheWriter()

  return r
}

// cacheWriter is a goroutine that writes articles view cache every midnight.
func (r *archiveRepository) cacheWriter() {
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

func (r *archiveRepository) Close() {
  close(r.done)
  r.writeViewsCache(context.TODO())
}

// closed checks if the repository has been closed, that is the
// Close method was invoked.
func (r *archiveRepository) closed() bool {
  select {
  default:
    return false
  case <-r.done:
    return true
  }
}

// writeViewsCache is a goroutine that writes the cached article views to the database;
// it should be run every midnight.
func (r *archiveRepository) writeViewsCache(ctx context.Context) {
  r.mu.RLock()

  if 0 == len(r.articleViewsCache) {
    r.mu.RUnlock()
    return
  }

  r.mu.RUnlock()

  slog.Info("writing article views cache to database")

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(err.Error())
    return
  }

  defer tx.Rollback()

  writeViewsQuery := `
  UPDATE "article"
     SET "views" = "views" + @views
   WHERE "uuid" = @uuid;`

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
        sql.Named("uuid", article),
        sql.Named("views", metadata.views))

      if nil != err {
        slog.Error(err.Error(), slog.String("article_uuid", article))
      }
    }()
  }

  writers.Wait()

  if !r.closed() {
    slog.Info("resetting article views cache")

    r.articleViewsCache = articleViewsCache{}
  }

  if err := tx.Commit(); nil != err {
    slog.Error(err.Error())
  }
}

// views returns the cached views of the given article.
func (r *archiveRepository) views(article string) int64 {
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
func (r *archiveRepository) incrementViews(ctx context.Context, article string) {
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
func (r *archiveRepository) cleanBrokenLinks() {
  r.cleanOnce.Do(func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    checkThereAreBrokenLinksQuery := `
    SELECT count (*)
      FROM "article_link"
     WHERE "expires_at" <= current_timestamp;`

    nbroken := 0

    err := r.db.QueryRowContext(ctx, checkThereAreBrokenLinksQuery).Scan(&nbroken)

    if nil != err {
      slog.Error(err.Error())
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
      slog.Error(err.Error())

      anew()
      if done() {
        return
      }

      goto again
    }

    defer tx.Rollback()

    removeBrokenLinksQuery := `
    DELETE FROM "article_link"
          WHERE "expires_at" <= current_timestamp;`

    ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    _, err = tx.ExecContext(ctx, removeBrokenLinksQuery)

    if nil != err {
      slog.Error(err.Error())

      anew()
      if done() {
        return
      }

      goto again
    }

    if err = tx.Commit(); nil != err {
      slog.Error(err.Error())
    }
  })
}

func (r *archiveRepository) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
  slog.Info("drafting new article", slog.String("title", creation.Title))

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

  var hasTopic bool

  hasTopicQuery := `
  SELECT count (1)
    FROM "article"
   WHERE "uuid" = $1
     AND "draft" IS TRUE
     AND "published_at" IS NULL
     AND "topic" IS NOT NULL
     AND "topic" <> '';`

  err = r.db.QueryRowContext(ctx, hasTopicQuery, id).Scan(&hasTopic)

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if !hasTopic {
    p := &problem.Problem{}
    p.Status(http.StatusBadRequest)
    p.Title("Could not publish draft.")
    p.Detail("Cannot publish a draft without making it belong to a topic first.")
    p.With("draft_uuid", id)
    return p
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

  slog.Info("publishing draft", slog.String("uuid", id))

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

  r.setPublicationsCache(ctx)

  return nil
}

func (r *archiveRepository) SetSlug(ctx context.Context, id, slug string) error {
  slog.Info("changing article slug", slog.String("article_uuid", id), slog.String("new_slug", slug))

  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  setSlugQuery := `
  UPDATE "article"
     SET "slug" = @slug
   WHERE "uuid" = @uuid
     AND "draft" IS FALSE
     AND "published_at" IS NOT NULL;`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, setSlugQuery, sql.Named("uuid", id), sql.Named("slug", slug))

  if nil != err && !errors.Is(sql.ErrNoRows, err) {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return problem.NewNotFound(id, "article")
  }

  if err := tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

func (r *archiveRepository) setPublicationsCache(ctx context.Context) {
  r.publicationsCache = nil
  r.publicationsCache, _ = r.Publications(ctx)
}

func (r *archiveRepository) Publications(ctx context.Context) (publications []*transfer.Publication, err error) {
  if 0 < len(r.publicationsCache) {
    return r.publicationsCache, nil
  }

  getPublicationsQuery := `
  SELECT cast(strftime('%m', "published_at") AS INTEGER) AS "month",
         cast(strftime('%Y', "published_at") AS INTEGER) AS "year"
    FROM "article"
   WHERE "draft" IS FALSE
     AND "published_at" IS NOT NULL
     AND "hidden" IS FALSE
     GROUP BY "month"
     ORDER BY "year" DESC, "month" DESC;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := r.db.QueryContext(ctx, getPublicationsQuery)

  if nil != err {
    slog.Error(err.Error())
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
      slog.Error(err.Error())
      return nil, err
    }

    publications = append(publications, &publication)
  }

  r.publicationsCache = publications

  return publications, nil
}

func (r *archiveRepository) Get(ctx context.Context, filter *transfer.ArticleFilter, hidden, draftsOnly bool) (articles []*transfer.Article, err error) {
  query := strings.Builder{}
  query.WriteString(`
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
               END`)

  if "" != filter.Search {
    for _, chunk := range strings.Fields(filter.Search) {
      if strings.Contains(chunk, "'") {
        chunk = strings.ReplaceAll(chunk, "'", "''")
      }

      query.WriteString("\nAND \"title\" LIKE '%" + chunk + "%'")
    }
  }

  query.WriteString(`
  ORDER BY "pinned" DESC, "published_at" DESC
  LIMIT @rpp
  OFFSET @rpp * (@page - 1);`)

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

  defer result.Close()

  URLBase := ""

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
      topicURL, err := url.JoinPath(URLBase, "archive", "topic")

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

func (r *archiveRepository) GetOne(ctx context.Context, request *transfer.ArticleRequest) (article *model.Article, err error) {
  requestArticleUUIDQuery := `
  SELECT "uuid"
    FROM "article"
   WHERE "draft" IS FALSE
     AND "published_at" IS NOT NULL
     AND "hidden" IS FALSE
     AND "topic" = @topic
     AND cast(strftime('%Y', "published_at") AS INTEGER) = @year
     AND cast(strftime('%m', "published_at") AS INTEGER) = @month
     AND "slug" = @slug;`

  var (
    year  = 0
    month = 0
  )

  if nil != request.Publication {
    year = request.Publication.Year
    month = int(request.Publication.Month)
  }

  var id string

  ctx1, cancel1 := context.WithTimeout(ctx, 10*time.Second)
  defer cancel1()

  err = r.db.QueryRowContext(ctx1, requestArticleUUIDQuery,
    sql.Named("topic", request.Topic),
    sql.Named("year", year),
    sql.Named("month", month),
    sql.Named("slug", request.Slug),
  ).Scan(&id)

  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  if "" == id {
    return nil, problem.NewNotFound(id, "article") // TODO: Do not return this kind of problem.
  }

  go r.incrementViews(ctx, id)

  return r.GetByID(ctx, id, false)
}

func (r *archiveRepository) GetByLink(ctx context.Context, link string) (article *model.Article, err error) {
  getByLinkQuery := `
  SELECT "article_uuid",
         "expires_at"
    FROM "article_link"
   WHERE "sharable_link" = $1;`

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
      slog.Error(err.Error())
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

  slog.Info("retrieving shared draft", slog.String("link", link))

  go r.incrementViews(ctx, id)

  return r.GetByID(ctx, id, true)
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

  ctx1, cancel1 := context.WithTimeout(ctx, 10*time.Second)
  defer cancel1()

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

  ctx2, cancel2 := context.WithTimeout(ctx, 10*time.Second)
  defer cancel2()

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

  r.db.QueryRowContext(ctx2, getArticleByUUIDQuery,
    sql.Named("uuid", id),
    sql.Named("is_draft", isDraft)).Scan(
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

  r.setPublicationsCache(ctx)

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

  r.setPublicationsCache(ctx)

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
  defer func() {
    if nil == err {
      r.mu.Lock()
      r.cleanOnce = sync.Once{}
      r.mu.Unlock()
    }
  }()

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
    if -1 == expiresAt.Compare(now) || 0 == expiresAt.Compare(now) { // link has expired
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

      slog.Info("shareable link expired, generating a new one",
        slog.String("elapsed", now.Sub(expiresAt).String()),
        slog.String("article_uuid", id))
    } else {
      return link, nil
    }
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

  ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
  defer cancel2()

  err = tx.QueryRowContext(ctx2, makeShareableLinkQuery,
    sql.Named("article_uuid", id),
    sql.Named("sharable_link", fmt.Sprintf("/archive/s/%x", hash))).
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

func (r *archiveRepository) Revise(ctx context.Context, id string, revision *transfer.ArticleRevision) error {
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

  if "" != revision.Topic {
    exists := false
    topicExistsQuery := `
    SELECT count (1)
      FROM "topic"
     WHERE "id" = $1;`

    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    err := r.db.QueryRowContext(ctx, topicExistsQuery, revision.Topic).Scan(&exists)

    if nil != err {
      slog.Error(err.Error())
      return err
    }

    if !exists {
      return problem.NewNotFound(revision.Topic, "topic")
    }
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
         "topic" = coalesce (nullif (@topic, ''), "topic"),
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
           "topic" = coalesce (nullif (@topic, ''), "topic"),
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
    sql.Named("topic", revision.Topic),
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
         "topic",
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
      &patch.TopicID,
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
         "topic" = coalesce(nullif(@topic, ''), "topic"),
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
    sql.Named("topic", patch.TopicID),
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
         "topic",
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
      &patch.TopicID,
      &patch.Content)

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    patches = append(patches, &patch)
  }

  return patches, nil
}
