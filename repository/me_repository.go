package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "log/slog"
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
  db *sql.DB
}

// NewMeRepository creates a new MeRepository instance associating db as its database.
func NewMeRepository(db *sql.DB) MeRepository {
  return &meRepositoryImpl{db}
}

func (r *meRepositoryImpl) registered(ctx context.Context) bool {
  var exists bool
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var err = r.db.
    QueryRowContext(ctx, `SELECT count (1) FROM "me";`).
    Scan(&exists)
  if nil != err {
    slog.Error(err.Error())
    return false
  }
  return exists
}

func (r *meRepositoryImpl) Register(ctx context.Context) {
  if r.registered(ctx) {
    return
  }
  var query = `
  INSERT INTO "me" ("summary",
                    "email",
                    "company",
                    "location")
            VALUES ('No summary provided.',
                    'email@example.com',
                    'None',
                    'Unknown');`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  _, err := r.db.ExecContext(ctx, query)
  if nil != err {
    slog.Error(err.Error())
  }
}

func (r *meRepositoryImpl) Get(ctx context.Context) (me *model.Me, err error) {
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, `SELECT * FROM "me";`)
  me = new(model.Me)
  err = row.Scan(
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
    slog.Error(err.Error())
    return me, err
  }
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
    slog.Error(err.Error())
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
  var query = `
    UPDATE "me"
       SET "summary" = coalesce (nullif (@new_summary, ''), @current_summary),
           "job_title" = coalesce (nullif (@new_job_title, ''), @current_job_title),
           "email" = coalesce (nullif (@new_email, ''), @current_email),
           "photo_url" = coalesce (nullif (@new_photo_url, ''), @current_photo_url),
           "resume_url" = coalesce (nullif (@new_resume_url, ''), @current_resume_url),
           "company" = coalesce (nullif (@new_company, ''), @current_company),
           "location" = coalesce (nullif (@new_location, ''), @current_location),
           "hireable" = @new_hireable,
           "github_url" = coalesce (nullif (@new_github_url, ''), @current_github_url),
           "linkedin_url" = coalesce (nullif (@new_linkedin_url, ''), @current_linkedin_url),
           "youtube_url" = coalesce (nullif (@new_youtube_url, ''), @current_youtube_url),
           "twitter_url" = coalesce (nullif (@new_twitter_url, ''), @current_twitter_url),
           "instagram_url" = coalesce (nullif (@new_instagram_url, ''), @current_instagram_url),
           "updated_at" = current_timestamp
     WHERE "username" = "fontseca.dev";`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query,
    sql.Named("new_summary", update.Summary), sql.Named("current_summary", current.Summary),
    sql.Named("new_job_title", update.JobTitle), sql.Named("current_job_title", current.JobTitle),
    sql.Named("new_email", update.Email), sql.Named("current_email", current.Email),
    sql.Named("new_photo_url", update.PhotoURL), sql.Named("current_photo_url", current.PhotoURL),
    sql.Named("new_resume_url", update.ResumeURL), sql.Named("current_resume_url", current.ResumeURL),
    sql.Named("new_company", update.Company), sql.Named("current_company", current.Company),
    sql.Named("new_location", update.Location), sql.Named("current_location", current.Location),
    sql.Named("new_hireable", update.Hireable),
    sql.Named("new_github_url", update.GitHubURL), sql.Named("current_github_url", current.GitHubURL),
    sql.Named("new_linkedin_url", update.LinkedInURL), sql.Named("current_linkedin_url", current.LinkedInURL),
    sql.Named("new_youtube_url", update.YouTubeURL), sql.Named("current_youtube_url", current.YouTubeURL),
    sql.Named("new_twitter_url", update.TwitterURL), sql.Named("current_twitter_url", current.TwitterURL),
    sql.Named("new_instagram_url", update.InstagramURL), sql.Named("current_instagram_url", current.InstagramURL))
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  var affected, _ = result.RowsAffected()
  if 1 != affected {
    return false, nil
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return false, err
  }
  return true, nil
}
