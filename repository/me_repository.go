package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "log/slog"
  "sync"
  "time"
)

// MeRepository is an abstraction of the database that provides
// actions for model.Me management.
type MeRepository interface {
  // Register creates my profile record. It ensures that the record is
  // created only once.
  Register(ctx context.Context)

  // Get retrieves the information of my profile.
  Get(ctx context.Context) (me *model.Me, err error)

  // Update updates the information of my profile.
  Update(ctx context.Context, update *transfer.MeUpdate) (ok bool, err error)
}

type meRepositoryImpl struct {
  cached *model.Me
  db     *sql.DB
  mu     sync.RWMutex
}

// NewMeRepository creates a new MeRepository instance associating db as its database.
func NewMeRepository(db *sql.DB) MeRepository {
  return &meRepositoryImpl{nil, db, sync.RWMutex{}}
}

func (r *meRepositoryImpl) registered(ctx context.Context) bool {
  var exists bool
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var err = r.db.
    QueryRowContext(ctx, `SELECT count (1) FROM "me"."me";`).
    Scan(&exists)
  if nil != err {
    slog.Error(getErrMsg(err))
    return false
  }
  return exists
}

func (r *meRepositoryImpl) Register(ctx context.Context) {
  if r.registered(ctx) {
    return
  }

  registerMeQuery := `
  INSERT INTO "me"."me" ("summary",
                         "email",
                         "company",
                         "location",
                         "job_title")
                 VALUES ('Empty.',
                         'nobody@unknown.com',
                         'Unknown',
                         'Unknown',
                         'Unknown');`

  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()

  _, err := r.db.ExecContext(ctx, registerMeQuery)

  if nil != err {
    slog.Error(getErrMsg(err))
  }

  r.cache(ctx)
}

func (r *meRepositoryImpl) Get(ctx context.Context) (me *model.Me, err error) {
  r.mu.RLock()
  if nil != r.cached {
    r.mu.RUnlock()
    return r.cached, nil
  }
  r.mu.RUnlock()

  getMeQuery := `
  SELECT "username",
         "first_name",
         "last_name",
         "summary",
         "job_title",
         "email",
         "photo_url",
         "resume_url",
         "coding_since",
         "company",
         "location",
         "hireable",
         "github_url",
         "linkedin_url",
         "youtube_url",
         "twitter_url",
         "instagram_url",
         "created_at",
         "updated_at"
    FROM "me"."me";`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  me = new(model.Me)
  err = r.db.QueryRowContext(ctx, getMeQuery).Scan(
    &me.Username,
    &me.FirstName,
    &me.LastName,
    &me.Summary,
    &me.JobTitle,
    &me.Email,
    &me.PhotoURL,
    &me.ResumeURL,
    &me.CodingSince,
    &me.Company,
    &me.Location,
    &me.Hireable,
    &me.GitHubURL,
    &me.LinkedInURL,
    &me.YouTubeURL,
    &me.TwitterURL,
    &me.InstagramURL,
    &me.CreatedAt,
    &me.UpdatedAt)

  if nil != err {
    slog.Error(getErrMsg(err))
    return me, err
  }

  r.mu.Lock()
  r.cached = me
  r.mu.Unlock()
  return me, nil
}

func (r *meRepositoryImpl) updatable(current *model.Me, update *transfer.MeUpdate) bool {
  if ("" == update.Summary || update.Summary == current.Summary) &&
    ("" == update.JobTitle || update.JobTitle == current.JobTitle) &&
    ("" == update.Email || update.Email == current.Email) &&
    ("" == update.PhotoURL || update.PhotoURL == current.PhotoURL) &&
    ("" == update.ResumeURL || update.ResumeURL == current.ResumeURL) &&
    ("" == update.Company || update.Company == current.Company) &&
    ("" == update.Location || update.Location == current.Location) &&
    (update.Hireable == current.Hireable) &&
    ("" == update.GitHubURL || update.GitHubURL == current.GitHubURL) &&
    ("" == update.LinkedInURL || update.LinkedInURL == current.LinkedInURL) &&
    ("" == update.YouTubeURL || update.YouTubeURL == current.YouTubeURL) &&
    ("" == update.InstagramURL || update.InstagramURL == current.InstagramURL) {
    return false
  }
  return true
}

func (r *meRepositoryImpl) Update(ctx context.Context, update *transfer.MeUpdate) (ok bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return false, err
  }

  defer tx.Rollback()

  current, err := r.Get(ctx)

  if nil != err {
    return false, err
  }

  if updatable := r.updatable(current, update); !updatable {
    return false, nil
  }

  updateMeQuery := `
    UPDATE "me"."me"
       SET "summary" = coalesce (nullif ($1, ''), $2),
           "job_title" = coalesce (nullif ($3, ''), $4),
           "email" = coalesce (nullif ($5, ''), $6),
           "photo_url" = coalesce (nullif ($7, ''), $8),
           "resume_url" = coalesce (nullif ($9, ''), $10),
           "company" = coalesce (nullif ($11, ''), $12),
           "location" = coalesce (nullif ($13, ''), $14),
           "hireable" = $15,
           "github_url" = coalesce (nullif ($16, ''), $17),
           "linkedin_url" = coalesce (nullif ($18, ''), $19),
           "youtube_url" = coalesce (nullif ($20, ''), $21),
           "twitter_url" = coalesce (nullif ($22, ''), $23),
           "instagram_url" = coalesce (nullif ($24, ''), $25),
           "updated_at" = current_timestamp
     WHERE "username" = 'fontseca.dev';`

  slog.Info("updating 'me' object")

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, updateMeQuery,
    update.Summary, current.Summary,
    update.JobTitle, current.JobTitle,
    update.Email, current.Email,
    update.PhotoURL, current.PhotoURL,
    update.ResumeURL, current.ResumeURL,
    update.Company, current.Company,
    update.Location, current.Location,
    update.Hireable,
    update.GitHubURL, current.GitHubURL,
    update.LinkedInURL, current.LinkedInURL,
    update.YouTubeURL, current.YouTubeURL,
    update.TwitterURL, current.TwitterURL,
    update.InstagramURL, current.InstagramURL)

  if nil != err {
    slog.Error(getErrMsg(err))
    return false, err
  }

  var affected, _ = result.RowsAffected()
  if 1 != affected {
    return false, nil
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return false, err
  }

  r.cache(ctx)

  return true, nil
}

func (r *meRepositoryImpl) cache(ctx context.Context) {
  r.mu.Lock()
  r.cached = nil
  r.mu.Unlock()
  r.cached, _ = r.Get(ctx)
}
