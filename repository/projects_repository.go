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

// List retrieves a slice of project types.
func (r *ProjectsRepository) List(ctx context.Context, archived bool) (projects []*model.Project, err error) {
  var getProjectsQuery = `
     SELECT p."uuid",
            p."name",
            p."slug",
            p."homepage",
            p."company",
            p."company_homepage",
            p."date_start",
            p."date_end",
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
   ORDER BY p."date_start" DESC NULLS FIRST, p."created_at" DESC;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  rows, err := r.db.QueryContext(ctx, getProjectsQuery, archived)
  if nil != err {
    slog.Error(getErrMsg(err))
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
      &project.Company,
      &project.CompanyHomepage,
      &project.Starts,
      &project.Ends,
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
      slog.Error(getErrMsg(err))
      return nil, err
    }
    if nil != tags && "" != *tags {
      project.TechnologyTags = strings.Split(*tags, ",")
    }
    projects = append(projects, project)
  }
  return projects, nil
}

// Get retrieves a project type by its UUID.
func (r *ProjectsRepository) Get(ctx context.Context, id string) (project *model.Project, err error) {
  var getProjectByIDQuery = `
     SELECT p."uuid",
            p."name",
            p."slug",
            p."homepage",
            p."company",
            p."company_homepage",
            p."date_start",
            p."date_end",
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
        AND p."archived" IS FALSE
   GROUP BY p."uuid";`

  ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
  defer cancel()

  project = new(model.Project)
  var tags *string
  err = r.db.QueryRowContext(ctx, getProjectByIDQuery, id).
    Scan(
      &project.UUID,
      &project.Name,
      &project.Slug,
      &project.Homepage,
      &project.Company,
      &project.CompanyHomepage,
      &project.Starts,
      &project.Ends,
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
      slog.Error(getErrMsg(err))
    }

    return nil, err
  }

  if nil != tags && "" != *tags {
    project.TechnologyTags = strings.Split(*tags, ",")
  }

  return project, nil
}

// GetBySlug retrieves a project type by its slug.
func (r *ProjectsRepository) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  var getProjectBySlugQuery = `
     SELECT p."uuid",
            p."name",
            p."slug",
            p."homepage",
            p."company",
            p."company_homepage",
            p."date_start",
            p."date_end",
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
    &project.Company,
    &project.CompanyHomepage,
    &project.Starts,
    &project.Ends,
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
      slog.Error(getErrMsg(err))
    }
    return nil, err
  }
  if nil != tags && "" != *tags {
    project.TechnologyTags = strings.Split(*tags, ",")
  }
  return project, nil
}

// Create creates a project record with the provided creation data.projectID
func (r *ProjectsRepository) Create(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return uuid.Nil.String(), err
  }
  defer tx.Rollback()
  var addProjectQuery = `
  INSERT INTO "projects"."project"
                        ("name",
                         "slug",
                         "homepage",
                         "company",
                         "company_homepage",
                         "date_start",
                         "date_end",
                         "language",
                         "summary",
                         "read_time",
                         "content",
                         "first_image_url",
                         "second_image_url",
                         "github_url",
                         "collection_url",
                         "archived")
                 VALUES ($1,
                         $2,
                         coalesce (nullif ($3, ''), 'about:blank'),
                         nullif ($4, ''),
                         nullif ($5, ''),
                         nullif ($6, '')::DATE,
                         nullif ($7, '')::DATE,
                         nullif ($8, ''),
                         coalesce (nullif ($9, ''), 'no summary'),
                         coalesce (nullif ($10, 0), 0),
                         coalesce (nullif ($11, ''), 'no content'),
                         coalesce (nullif ($12, ''), 'about:blank'),
                         coalesce (nullif ($13, ''), 'about:blank'),
                         coalesce (nullif ($14, ''), 'about:blank'),
                         coalesce (nullif ($15, ''), 'about:blank'),
                         TRUE)
              RETURNING "uuid";`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  var row = tx.QueryRowContext(ctx, addProjectQuery,
    creation.Name,
    creation.Slug,
    creation.Homepage,
    creation.Company,
    creation.CompanyHomepage,
    creation.Starts,
    creation.Ends,
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
    slog.Error(getErrMsg(err))
    return uuid.Nil.String(), err
  }
  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return uuid.Nil.String(), err
  }
  return id, nil
}

// Exists checks whether a project exists in the database.
// If it does, it returns nil; otherwise a not found error.
func (r *ProjectsRepository) Exists(ctx context.Context, id string) error {
  var query = `
  SELECT count (1)
    FROM projects."project"
   WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var row = r.db.QueryRowContext(ctx, query, id)
  var exists bool
  err := row.Scan(&exists)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  if !exists {
    return problem.NewNotFound(id, "project")
  }
  return nil
}

// Update modifies an existing project record with the provided update data.
func (r *ProjectsRepository) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  defer tx.Rollback()
  var updateProjectQuery = `
  UPDATE "projects"."project"
     SET "name" = coalesce (nullif ($2, ''), "name"),
         "slug" = coalesce (nullif ($3, ''), "slug"),
         "homepage" = coalesce (nullif ($4, ''), "homepage"),
         "language" = coalesce (nullif ($5, ''), "language"),
         "summary" = coalesce (nullif ($6, ''), "summary"),
         "read_time" = CASE WHEN $7 = "read_time" OR 0 >= $7 THEN "read_time" ELSE $7 END,
         "content" = coalesce (nullif ($8, ''), "content"),
         "first_image_url" = coalesce (nullif ($9, ''), "first_image_url"),
         "second_image_url" = coalesce (nullif ($10, ''), "second_image_url"),
         "github_url" = coalesce (nullif ($11, ''), "github_url"),
         "collection_url" = coalesce (nullif ($12, ''), "collection_url"),
         "playground_url" = coalesce (nullif ($13, ''), "playground_url"),
         "playable" = CASE WHEN $13 <> '' AND $13 = 'about:blank' AND "playable" THEN FALSE
                           WHEN $13 <> '' AND $13 <> 'about:blank' THEN TRUE
                           ELSE "playable" END,
         "company" = coalesce (nullif ($14, ''), "company"),
         "company_homepage" = coalesce (nullif ($15, ''), "company_homepage"),
         "date_start" = coalesce (nullif ($16, '')::DATE, "date_start"),
         "date_end" = coalesce (nullif ($17, '')::DATE, "date_end"),
         "updated_at" = current_timestamp
   WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, updateProjectQuery,
    id,
    update.Name,
    update.Slug,
    update.Homepage,
    update.Language,
    update.Summary,
    update.ReadTime,
    update.Content,
    update.FirstImageURL,
    update.SecondImageURL,
    update.GitHubURL,
    update.CollectionURL,
    update.PlaygroundURL,
    update.Company,
    update.CompanyHomepage,
    update.Starts,
    update.Ends)
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  var affected, _ = result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "project")
  }
  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  return nil
}

// SetArchived makes a project archived or unarchived. If the project is archived it cannot be normally listed.
func (r *ProjectsRepository) SetArchived(ctx context.Context, id string, archive bool) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  defer tx.Rollback()
  var query = `
  UPDATE "projects"."project"
     SET "archived" = $2,
         "updated_at" = current_timestamp
   WHERE "uuid" = $1;`
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  result, err := tx.ExecContext(ctx, query, id, archive)
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
  return nil
}

// Remove deletes an existing project type. If not found, returns a not found error.
func (r *ProjectsRepository) Remove(ctx context.Context, id string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
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
    slog.Error(getErrMsg(err))
    return err
  }

  affected, _ := result.RowsAffected()
  if 1 != affected {
    return problem.NewNotFound(id, "project")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }

  return nil
}

// HasTag checks whether technologyTagID belongs to projectID.
func (r *ProjectsRepository) HasTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  var hasTechTagQuery = `
  SELECT count (1)
    FROM projects.project_tag ptt
   WHERE ptt."project_uuid" = $1
     AND ptt."technology_tag_uuid" = $2;`
  err = r.db.QueryRowContext(ctx, hasTechTagQuery, projectID, technologyTagID).Scan(&success)
  if nil != err {
    slog.Error(getErrMsg(err))
    return false, err
  }
  return success, nil
}

// AddTag adds an existing technology tag that will belong to the project represented by projectID .
func (r *ProjectsRepository) AddTag(ctx context.Context, projectID, technologyTagID string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
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
    slog.Error(getErrMsg(err))
    return err
  }
  affected, _ := result.RowsAffected()
  if 1 != affected {
    return nil
  }
  if err = tx.Commit(); nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  return nil
}

// RemoveTag removes a technology tag that belongs to the project represented by projectID.
func (r *ProjectsRepository) RemoveTag(ctx context.Context, projectID, technologyTagID string) error {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(getErrMsg(err))
    return err
  }
  defer tx.Rollback()
  var removeTechTagQuery = `
  DELETE FROM "projects"."project_tag"
        WHERE "project_uuid" = $1
          AND "technology_tag_uuid" = $2;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  _, err = tx.ExecContext(ctx, removeTechTagQuery, projectID, technologyTagID)
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
