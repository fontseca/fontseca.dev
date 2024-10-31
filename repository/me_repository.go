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
type MeRepository struct {
  cached *model.Me
  db     *sql.DB
  mu     sync.RWMutex
}

// NewMeRepository creates a new MeRepository instance associating db as its database.
func NewMeRepository(db *sql.DB) *MeRepository {
  return &MeRepository{nil, db, sync.RWMutex{}}
}

// registered checks if the sole record in the table "me"."me" is already set.
func (r *MeRepository) registered(ctx context.Context) bool {
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

// Register creates my profile record. It ensures that the record is
// created only once.
func (r *MeRepository) Register(ctx context.Context) {
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

// Get retrieves the information of my profile.
func (r *MeRepository) Get(ctx context.Context) (me *model.Me, err error) {
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

// Update updates the information of my profile.
func (r *MeRepository) Update(ctx context.Context, update *transfer.MeUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  updateMeQuery := `
    UPDATE "me"."me"
       SET "summary" = coalesce (nullif ($1, ''), "summary"),
           "job_title" = coalesce (nullif ($2, ''), "job_title"),
           "email" = coalesce (nullif ($3, ''), "email"),
           "photo_url" = coalesce (nullif ($4, ''), "photo_url"),
           "resume_url" = coalesce (nullif ($5, ''), "resume_url"),
           "company" = coalesce (nullif ($6, ''), "company"),
           "location" = coalesce (nullif ($7, ''), "location"),
           "github_url" = coalesce (nullif ($8, ''), "github_url"),
           "linkedin_url" = coalesce (nullif ($9, ''), "linkedin_url"),
           "youtube_url" = coalesce (nullif ($10, ''), "youtube_url"),
           "twitter_url" = coalesce (nullif ($11, ''), "twitter_url"),
           "instagram_url" = coalesce (nullif ($12, ''), "instagram_url"),
           "updated_at" = current_timestamp
     WHERE "username" = 'fontseca.dev';`

  slog.Info("updating 'me' object")

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, updateMeQuery,
    update.Summary,
    update.JobTitle,
    update.Email,
    update.PhotoURL,
    update.ResumeURL,
    update.Company,
    update.Location,
    update.GitHubURL,
    update.LinkedInURL,
    update.YouTubeURL,
    update.TwitterURL,
    update.InstagramURL)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  var affected, _ = result.RowsAffected()
  if 1 != affected {
    return nil
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  r.cache(ctx)

  return nil
}

// SetHireable defines whether I am currently hireable or not.
func (r *MeRepository) SetHireable(ctx context.Context, hireable bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  defer tx.Rollback()

  query := `
    UPDATE "me"."me"
       SET "hireable" = $1,
           "updated_at" = current_timestamp
     WHERE "username" = 'fontseca.dev';`

  slog.Info("updating 'me' object")

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, query, hireable)

  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  var affected, _ = result.RowsAffected()
  if 1 != affected {
    return nil
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  r.cache(ctx)

  return nil
}

func (r *MeRepository) cache(ctx context.Context) {
  r.mu.Lock()
  r.cached = nil
  r.mu.Unlock()
  r.cached, _ = r.Get(ctx)
}
