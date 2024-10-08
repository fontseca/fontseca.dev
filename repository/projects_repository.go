package repository

import (
  "context"
  "database/sql"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "log/slog"
  "strings"
  "time"
)

// ProjectsRepository provides methods for interacting with project data in the database.
type ProjectsRepository struct {
  db *sql.DB
}

func NewProjectsRepository(db *sql.DB) *ProjectsRepository {
  return &ProjectsRepository{db}
}

// Get retrieves a slice of project types.
func (r *ProjectsRepository) Get(ctx context.Context, archived bool) (projects []*model.Project, err error) {
  var getProjectsQuery = `
     SELECT p."uuid",
            p."name",
            p."slug",
            p."homepage",
            p."language",
            p."summary",
            p."read_time",
            p."content",
            p."first_image_url",
            p."second_image_url",
            p."github_url",
            p."collection_url",
            p."playground_url",
            p."playable",
            p."archived",
            p."finished",
            p."created_at",
            p."updated_at",
            string_agg (tt."name", ',')
       FROM "projects"."project" p
  LEFT JOIN "projects"."project_tag" ptt
         ON ptt."project_uuid" = p."uuid"
  LEFT JOIN "projects"."tag" tt
         ON tt."uuid" = ptt."technology_tag_uuid"
      WHERE p."archived" = $1
   GROUP BY p."uuid"
   ORDER BY p."created_at" DESC;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  rows, err := r.db.QueryContext(ctx, getProjectsQuery, archived)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }
  defer rows.Close()
  projects = make([]*model.Project, 0)
  for rows.Next() {
    var (
      project = new(model.Project)
      tags    *string
    )
    err = rows.Scan(
      &project.UUID,
      &project.Name,
      &project.Slug,
      &project.Homepage,
      &project.Language,
      &project.Summary,
      &project.ReadTime,
      &project.Content,
      &project.FirstImageURL,
      &project.SecondImageURL,
      &project.GitHubURL,
      &project.CollectionURL,
      &project.PlaygroundURL,
      &project.Playable,
      &project.Archived,
      &project.Finished,
      &project.CreatedAt,
      &project.UpdatedAt,
      &tags)
    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }
    if nil != tags && "" != *tags {
      project.TechnologyTags = strings.Split(*tags, ",")
    }
    projects = append(projects, project)
  }
  return projects, nil
}

func (r *ProjectsRepository) doGetByID(ctx context.Context, id string, ignoreArchived bool) (project *model.Project, err error) {
  var getProjectByIDQuery = `
     SELECT p."uuid",
            p."name",
            p."slug",
            p."homepage",
            p."language",
            p."summary",
            p."read_time",
            p."content",
            p."first_image_url",
            p."second_image_url",
            p."github_url",
            p."collection_url",
            p."playground_url",
            p."playable",
            p."archived",
            p."finished",
            p."created_at",
            p."updated_at",
            string_agg (tt."name", ',')
       FROM "projects"."project" p
  LEFT JOIN "projects"."project_tag" ptt
         ON ptt."project_uuid" = p."uuid"
  LEFT JOIN "projects"."tag" tt
         ON tt."uuid" = ptt."technology_tag_uuid"
      WHERE p."uuid" = $1
        AND p."archived" = $2
   GROUP BY p."uuid";`

  ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
  defer cancel()

  project = new(model.Project)
  var tags *string
  err = r.db.QueryRowContext(ctx, getProjectByIDQuery, id, ignoreArchived).
    Scan(
      &project.UUID,
      &project.Name,
      &project.Slug,
      &project.Homepage,
      &project.Language,
      &project.Summary,
      &project.ReadTime,
      &project.Content,
      &project.FirstImageURL,
      &project.SecondImageURL,
      &project.GitHubURL,
      &project.CollectionURL,
      &project.PlaygroundURL,
      &project.Playable,
      &project.Archived,
      &project.Finished,
      &project.CreatedAt,
      &project.UpdatedAt,
      &tags)

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      err = problem.NewNotFound(id, "project")
    } else {
      slog.Error(err.Error())
    }

    return nil, err
  }

  if nil != tags && "" != *tags {
    project.TechnologyTags = strings.Split(*tags, ",")
  }

  return project, nil
}

// GetByID retrieves a project type by its UUID.
func (r *ProjectsRepository) GetByID(ctx context.Context, id string) (project *model.Project, err error) {
  return r.doGetByID(ctx, id, false)
}

// GetBySlug retrieves a project type by its slug.
func (r *ProjectsRepository) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  var getProjectBySlugQuery = `
     SELECT p."uuid",
            p."name",
            p."slug",
            p."homepage",
            p."language",
            p."summary",
            p."read_time",
            p."content",
            p."first_image_url",
            p."second_image_url",
            p."github_url",
            p."collection_url",
            p."playground_url",
            p."playable",
            p."archived",
            p."finished",
            p."created_at",
            p."updated_at",
            string_agg (tt."name", ',')
       FROM "projects"."project" p
  LEFT JOIN "projects"."project_tag" ptt
         ON ptt."project_uuid" = p."uuid"
  LEFT JOIN "projects"."tag" tt
         ON tt."uuid" = ptt."technology_tag_uuid"
      WHERE p."archived" IS FALSE
        AND p."slug" = $1
   GROUP BY p."uuid"
      LIMIT 1;`
  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()
  project = new(model.Project)
  var tags *string
  err = r.db.QueryRowContext(ctx, getProjectBySlugQuery, slug).Scan(
    &project.UUID,
    &project.Name,
    &project.Slug,
    &project.Homepage,
    &project.Language,
    &project.Summary,
    &project.ReadTime,
    &project.Content,
    &project.FirstImageURL,
    &project.SecondImageURL,
    &project.GitHubURL,
    &project.CollectionURL,
    &project.PlaygroundURL,
    &project.Playable,
    &project.Archived,
    &project.Finished,
    &project.CreatedAt,
    &project.UpdatedAt,
    &tags)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      err = problem.NewSlugNotFound(slug, "project")
    } else {
      slog.Error(err.Error())
    }
    return nil, err
  }
  if nil != tags && "" != *tags {
    project.TechnologyTags = strings.Split(*tags, ",")
  }
  return project, nil
}

// Add creates a project record with the provided creation data.projectID
func (r *ProjectsRepository) Add(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }
  defer tx.Rollback()
  var addProjectQuery = `
  INSERT INTO "projects"."project"
                        ("name",
                         "slug",
                         "homepage",
                         "language",
                         "summary",
                         "read_time",
                         "content",
                         "first_image_url",
                         "second_image_url",
                         "github_url",
                         "collection_url")
                 VALUES ($1,
                         $2,
                         coalesce (nullif ($3, ''), 'about:blank'),
                         nullif ($4, ''),
                         coalesce (nullif ($5, ''), 'no summary'),
                         coalesce (nullif ($6, 0), 0),
                         coalesce (nullif ($7, ''), 'no content'),
                         coalesce (nullif ($8, ''), 'about:blank'),
                         coalesce (nullif ($9, ''), 'about:blank'),
                         coalesce (nullif ($10, ''), 'about:blank'),
                         coalesce (nullif ($11, ''), 'about:blank'))
              RETURNING "uuid";`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  var row = tx.QueryRowContext(ctx, addProjectQuery,
    creation.Name,
    creation.Slug,
    creation.Homepage,
    creation.Language,
    creation.Summary,
    creation.ReadTime,
    creation.Content,
    creation.FirstImageURL,
    creation.SecondImageURL,
    creation.GitHubURL,
    creation.CollectionURL)
  err = row.Scan(&id)
  if nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }
  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }
  return id, nil
}

// Exists checks whether a project exists in the database.
// If it does, it returns nil; otherwise a not found error.
func (r *ProjectsRepository) Exists(ctx context.Context, id string) (err error) {
  var query = `
  SELECT count (1)
    FROM projects."project"
   WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, query, id)
  var exists bool
  err = row.Scan(&exists)
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  if !exists {
    return problem.NewNotFound(id, "project")
  }
  return nil
}

func (r *ProjectsRepository) nothingToUpdate(current *model.Project, update *transfer.ProjectUpdate) bool {
  return ("" == update.Name || update.Name == current.Name) &&
    ("" == update.Slug || update.Slug == current.Slug) &&
    ("" == update.Homepage || update.Homepage == current.Homepage) &&
    ("" == update.Language || nil != current.Language && update.Language == *current.Language) &&
    ("" == update.Summary || update.Summary == current.Summary) &&
    (0 == update.ReadTime || update.ReadTime == current.ReadTime) &&
    ("" == update.Content || update.Content == current.Content) &&
    ("" == update.FirstImageURL || update.FirstImageURL == current.FirstImageURL) &&
    ("" == update.SecondImageURL || update.SecondImageURL == current.SecondImageURL) &&
    ("" == update.GitHubURL || update.GitHubURL == current.GitHubURL) &&
    ("" == update.CollectionURL || update.CollectionURL == current.CollectionURL) &&
    ("" == update.PlaygroundURL || update.PlaygroundURL == current.PlaygroundURL) &&
    (update.Archived == current.Archived) &&
    (update.Finished == current.Finished)
}

func (r *ProjectsRepository) doUpdate(ctx context.Context, id string, update *transfer.ProjectUpdate, ignoreArchived bool) (updated bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  current, err := r.doGetByID(ctx, id, ignoreArchived)
  if nil != err {
    return false, err
  }
  if r.nothingToUpdate(current, update) {
    return false, nil
  }
  var updateProjectQuery = `
  UPDATE "projects"."project"
     SET "name" = coalesce (nullif ($1, ''), $2),
         "slug" = coalesce (nullif ($3, ''), $4),
         "homepage" = coalesce (nullif ($5, ''), $6),
         "language" = coalesce (nullif ($7, ''), $8),
         "summary" = coalesce (nullif ($9, ''), $10),
         "read_time" = CASE WHEN $11::INTEGER = $12::INTEGER OR 0 = $11::INTEGER
                            THEN $12::INTEGER
                            ELSE $11::INTEGER
                            END,
         "content" = coalesce (nullif ($13, ''), $14),
         "first_image_url" = coalesce (nullif ($15, ''), $16),
         "second_image_url" = coalesce (nullif ($17, ''), $18),
         "github_url" = coalesce (nullif ($19, ''), $20),
         "collection_url" = coalesce (nullif ($21, ''), $22),
         "playground_url" = coalesce (nullif ($23, ''), $24),
         "playable" = $25,
         "archived" = $26,
         "finished" = $27,
         "updated_at" = current_timestamp
   WHERE "uuid" = $28;`
  var playable = current.Playable
  if "" != update.PlaygroundURL {
    var wantsToDefaultPlaygroundURL = "about:blank" == update.PlaygroundURL
    var notSamePlaygroundURLs = update.PlaygroundURL != current.PlaygroundURL
    if playable && wantsToDefaultPlaygroundURL && notSamePlaygroundURLs {
      playable = false
    } else if !playable && notSamePlaygroundURLs {
      playable = true
    }
  }
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, updateProjectQuery,
    update.Name, current.Name,
    update.Slug, current.Slug,
    update.Homepage, current.Homepage,
    update.Language, current.Language,
    update.Summary, current.Summary,
    update.ReadTime, current.ReadTime,
    update.Content, current.Content,
    update.FirstImageURL, current.FirstImageURL,
    update.SecondImageURL, current.SecondImageURL,
    update.GitHubURL, current.GitHubURL,
    update.CollectionURL, current.CollectionURL,
    update.PlaygroundURL, current.PlaygroundURL,
    playable,
    update.Archived,
    update.Finished,
    id)
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

// Update modifies an existing project record with the provided update data.
func (r *ProjectsRepository) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) (updated bool, err error) {
  return r.doUpdate(ctx, id, update, false)
}

// Unarchive makes a project not archived so that it can be normally listed.
func (r *ProjectsRepository) Unarchive(ctx context.Context, id string) (updated bool, err error) {
  return r.doUpdate(ctx, id, &transfer.ProjectUpdate{Archived: false}, true)
}

// Remove deletes an existing project type. If not found, returns a not found error.
func (r *ProjectsRepository) Remove(ctx context.Context, id string) (err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }
  defer tx.Rollback()

  // Remove technology tags associated with the project to remove.

  var query = `
  DELETE FROM "projects"."project_tag"
        WHERE "project_uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  _, err = tx.ExecContext(ctx, query, id)
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  // Remove the actual project.

  query = `
  DELETE FROM "projects"."project"
        WHERE "uuid" = $1;`
  ctx, cancel = context.WithTimeout(ctx, time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, query, id)
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "project")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
}

// ContainsTechnologyTag checks whether technologyTagID belongs to projectID.
func (r *ProjectsRepository) ContainsTechnologyTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  var hasTechTagQuery = `
  SELECT count (1)
    FROM projects.project_tag ptt
   WHERE ptt."project_uuid" = $1
     AND ptt."technology_tag_uuid" = $2;`
  err = r.db.QueryRowContext(ctx, hasTechTagQuery, projectID, technologyTagID).Scan(&success)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  return success, nil
}

// AddTechnologyTag adds an existing technology tag that will belong to the project represented by projectID .
func (r *ProjectsRepository) AddTechnologyTag(ctx context.Context, projectID, technologyTagID string) (added bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  var addTechTagQuery = `
  INSERT INTO "projects"."project_tag" ("project_uuid",
                                        "technology_tag_uuid")
                                VALUES ($1,
                                        $2);`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, addTechTagQuery, projectID, technologyTagID)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return false, nil
  }
  if err = tx.Commit(); nil != err {
    return false, err
  }
  return true, nil
}

// RemoveTechnologyTag removes a technology tag that belongs to the project represented by projectID.
func (r *ProjectsRepository) RemoveTechnologyTag(ctx context.Context, projectID, technologyTagID string) (removed bool, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  defer tx.Rollback()
  var removeTechTagQuery = `
  DELETE FROM "projects"."project_tag"
        WHERE "project_uuid" = $1
          AND "technology_tag_uuid" = $2;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, removeTechTagQuery, projectID, technologyTagID)
  if nil != err {
    slog.Error(err.Error())
    return false, err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return false, nil
  }
  if err = tx.Commit(); nil != err {
    return false, err
  }
  return true, nil
}
