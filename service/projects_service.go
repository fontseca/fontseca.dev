package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "net/http"
  "strings"
  "time"
)

type projectsRepositoryAPI interface {
  List(ctx context.Context, archived bool) ([]*model.Project, error)
  Get(ctx context.Context, projectID string) (*model.Project, error)
  GetBySlug(ctx context.Context, projectID string) (*model.Project, error)
  Create(ctx context.Context, creation *transfer.ProjectCreation) (string, error)
  Exists(ctx context.Context, projectID string) error
  Update(ctx context.Context, projectID string, update *transfer.ProjectUpdate) error
  SetArchived(ctx context.Context, id string, archive bool) error
  Remove(ctx context.Context, projectID string) error
  HasTag(ctx context.Context, projectID, tagID string) (bool, error)
  AddTag(ctx context.Context, projectID, tagID string) error
  RemoveTag(ctx context.Context, projectID, tagID string) error
}

type technologyTagsServiceAPI interface {
  Exists(ctx context.Context, id string) error
}

// ProjectsService provides methods for interacting with projects
// data at a higher level and allows extra validation.
type ProjectsService struct {
  r    projectsRepositoryAPI
  tags technologyTagsServiceAPI
}

func NewProjectsService(repository projectsRepositoryAPI, tags technologyTagsServiceAPI) *ProjectsService {
  return &ProjectsService{
    r:    repository,
    tags: tags,
  }
}

// List retrieves a slice of projects.
func (s *ProjectsService) List(ctx context.Context, archived ...bool) (projects []*model.Project, err error) {
  var a = false
  if 0 != len(archived) && archived[0] {
    a = true
  }
  return s.r.List(ctx, a)
}

// Get retrieves a single project by its UUID.
func (s *ProjectsService) Get(ctx context.Context, id string) (project *model.Project, err error) {
  if err = validateUUID(&id); err != nil {
    return nil, err
  }
  return s.r.Get(ctx, id)
}

// GetBySlug retrieves a project type by its slug.
func (s *ProjectsService) GetBySlug(ctx context.Context, slug string) (project *model.Project, err error) {
  return s.r.GetBySlug(ctx, slug)
}

// Create creates a project record with the provided creation data.
func (s *ProjectsService) Create(ctx context.Context, creation *transfer.ProjectCreation) (id string, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return "", err
  }

  creation.Name = strings.TrimSpace(creation.Name)
  sanitizeTextWordIntersections(&creation.Name)
  creation.Homepage = strings.TrimSpace(creation.Homepage)
  creation.Company = strings.TrimSpace(creation.Company)
  creation.CompanyHomepage = strings.TrimSpace(creation.CompanyHomepage)
  creation.Starts = strings.TrimSpace(creation.Starts)
  creation.Ends = strings.TrimSpace(creation.Ends)
  creation.Language = strings.TrimSpace(creation.Language)
  creation.Summary = strings.TrimSpace(creation.Summary)
  creation.Content = strings.TrimSpace(creation.Content)
  creation.FirstImageURL = strings.TrimSpace(creation.FirstImageURL)
  creation.SecondImageURL = strings.TrimSpace(creation.SecondImageURL)
  creation.GitHubURL = strings.TrimSpace(creation.GitHubURL)
  creation.CollectionURL = strings.TrimSpace(creation.CollectionURL)

  switch {
  case 0 != len(creation.Name) && 128 < len(creation.Name):
    return "", problem.NewValidation([3]string{"name", "max", "128"})
  case 0 != len(creation.Homepage) && 2048 < len(creation.Homepage):
    return "", problem.NewValidation([3]string{"homepage", "max", "2048"})
  case 0 != len(creation.Company) && 64 < len(creation.Company):
    return "", problem.NewValidation([3]string{"company", "max", "64"})
  case 0 != len(creation.CompanyHomepage) && 2048 < len(creation.CompanyHomepage):
    return "", problem.NewValidation([3]string{"company_homepage", "max", "2048"})
  case 0 != len(creation.Language) && 64 < len(creation.Language):
    return "", problem.NewValidation([3]string{"language", "max", "64"})
  case 0 != len(creation.Summary) && 1024 < len(creation.Summary):
    return "", problem.NewValidation([3]string{"summary", "max", "1024"})
  case 0 != len(creation.Content) && 3145728 < len(creation.Content):
    return "", problem.NewValidation([3]string{"content", "max", "3145728"})
  case 0 != len(creation.FirstImageURL) && 2048 < len(creation.FirstImageURL):
    return "", problem.NewValidation([3]string{"first_image_url", "max", "2048"})
  case 0 != len(creation.SecondImageURL) && 2048 < len(creation.SecondImageURL):
    return "", problem.NewValidation([3]string{"second_image_url", "max", "2048"})
  case 0 != len(creation.GitHubURL) && 2048 < len(creation.GitHubURL):
    return "", problem.NewValidation([3]string{"github_url", "max", "2048"})
  case 0 != len(creation.CollectionURL) && 2048 < len(creation.CollectionURL):
    return "", problem.NewValidation([3]string{"collection_url", "max", "2048"})
  }

  err = sanitizeURL(
    &creation.Homepage,
    &creation.FirstImageURL,
    &creation.SecondImageURL,
    &creation.GitHubURL,
    &creation.CollectionURL,
    &creation.CompanyHomepage)
  if nil != err {
    return "", err
  }

  if "" != creation.Starts {
    tm, err := time.Parse(time.DateOnly, creation.Starts)
    if nil != err {
      switch {
      default:
        return "", problem.NewInternal()
      case strings.Contains(err.Error(), "cannot parse"):
        return "", problem.NewValidation([3]string{"starts", "format", "YYYY-MM-DD"})
      case strings.Contains(err.Error(), "out of range"):
        return "", problem.NewValueOutOfRange("date", "starts", creation.Starts)
      }
    }

    creation.Starts = tm.Format(time.DateOnly)
  }

  if "" != creation.Ends {
    tm, err := time.Parse(time.DateOnly, creation.Ends)
    if nil != err {
      switch {
      default:
        return "", problem.NewInternal()
      case strings.Contains(err.Error(), "cannot parse"):
        return "", problem.NewValidation([3]string{"ends", "format", "YYYY-MM-DD"})
      case strings.Contains(err.Error(), "out of range"):
        return "", problem.NewValueOutOfRange("date", "ends", creation.Ends)
      }
    }

    creation.Ends = tm.Format(time.DateOnly)
  }

  var projectText strings.Builder

  if "" != creation.Name {
    projectText.WriteString(creation.Name)
  }

  if "" != creation.Summary {
    sanitizeTextWordIntersections(&creation.Summary)
    if 60 < wordsIn(creation.Summary) {
      return "", problem.NewValidation([3]string{"summary", "max", "60"})
    }

    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(creation.Summary)
  }

  if "" != creation.Content {
    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(creation.Content)
  }

  if 0 < projectText.Len() {
    var r = strings.NewReader(projectText.String())
    creation.ReadTime = computePostReadingTimeInMinutes(r)
  }

  creation.Slug = generateSlug(creation.Name)

  return s.r.Create(ctx, creation)
}

// Exists checks whether a given project exists in the database.
// If it does, it returns nil; otherwise a not found error.
func (s *ProjectsService) Exists(ctx context.Context, id string) error {
  if err := validateUUID(&id); err != nil {
    return err
  }
  return s.r.Exists(ctx, id)
}

// Update modifies an existing project record with the provided update data.
func (s *ProjectsService) Update(ctx context.Context, id string, update *transfer.ProjectUpdate) error {
  if nil == update {
    err := errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return err
  }

  if err := validateUUID(&id); err != nil {
    return err
  }

  if update.Archived {
    return s.r.SetArchived(ctx, id, true)
  }

  update.Name = strings.TrimSpace(update.Name)

  if "" != update.Name {
    sanitizeTextWordIntersections(&update.Name)
  }

  update.Homepage = strings.TrimSpace(update.Homepage)
  update.Company = strings.TrimSpace(update.Company)
  update.CompanyHomepage = strings.TrimSpace(update.CompanyHomepage)
  update.Starts = strings.TrimSpace(update.Starts)
  update.Ends = strings.TrimSpace(update.Ends)
  update.Language = strings.TrimSpace(update.Language)
  update.Summary = strings.TrimSpace(update.Summary)
  update.Content = strings.TrimSpace(update.Content)
  update.FirstImageURL = strings.TrimSpace(update.FirstImageURL)
  update.SecondImageURL = strings.TrimSpace(update.SecondImageURL)
  update.GitHubURL = strings.TrimSpace(update.GitHubURL)
  update.CollectionURL = strings.TrimSpace(update.CollectionURL)
  update.PlaygroundURL = strings.TrimSpace(update.PlaygroundURL)

  switch {
  case 0 != len(update.Name) && 36 < len(update.Name):
    return problem.NewValidation([3]string{"name", "max", "36"})
  case 0 != len(update.Homepage) && 2048 < len(update.Homepage):
    return problem.NewValidation([3]string{"homepage", "max", "2048"})
  case 0 != len(update.Company) && 64 < len(update.Company):
    return problem.NewValidation([3]string{"company", "max", "64"})
  case 0 != len(update.CompanyHomepage) && 2048 < len(update.CompanyHomepage):
    return problem.NewValidation([3]string{"company_homepage", "max", "2048"})
  case 0 != len(update.Language) && 64 < len(update.Language):
    return problem.NewValidation([3]string{"language", "max", "64"})
  case 0 != len(update.Summary) && 1024 < len(update.Summary):
    return problem.NewValidation([3]string{"summary", "max", "1024"})
  case 0 != len(update.Content) && 3145728 < len(update.Content):
    return problem.NewValidation([3]string{"content", "max", "3145728"})
  case 0 != len(update.FirstImageURL) && 2048 < len(update.FirstImageURL):
    return problem.NewValidation([3]string{"first_image_url", "max", "2048"})
  case 0 != len(update.SecondImageURL) && 2048 < len(update.SecondImageURL):
    return problem.NewValidation([3]string{"second_image_url", "max", "2048"})
  case 0 != len(update.GitHubURL) && 2048 < len(update.GitHubURL):
    return problem.NewValidation([3]string{"github_url", "max", "2048"})
  case 0 != len(update.CollectionURL) && 2048 < len(update.CollectionURL):
    return problem.NewValidation([3]string{"collection_url", "max", "2048"})
  case 0 != len(update.PlaygroundURL) && 2048 < len(update.PlaygroundURL):
    return problem.NewValidation([3]string{"playground_url", "max", "2048"})
  }

  err := sanitizeURL(
    &update.Homepage,
    &update.FirstImageURL,
    &update.SecondImageURL,
    &update.GitHubURL,
    &update.CollectionURL,
    &update.PlaygroundURL,
    &update.CompanyHomepage)
  if nil != err {
    return err
  }

  if "" != update.Starts {
    tm, err := time.Parse(time.DateOnly, update.Starts)
    if nil != err {
      switch {
      default:
        return problem.NewInternal()
      case strings.Contains(err.Error(), "cannot parse"):
        return problem.NewValidation([3]string{"starts", "format", "YYYY-MM-DD"})
      case strings.Contains(err.Error(), "out of range"):
        return problem.NewValueOutOfRange("date", "starts", update.Starts)
      }
    }

    update.Starts = tm.Format(time.DateOnly)
  }

  if "" != update.Ends {
    tm, err := time.Parse(time.DateOnly, update.Ends)
    if nil != err {
      switch {
      default:
        return problem.NewInternal()
      case strings.Contains(err.Error(), "cannot parse"):
        return problem.NewValidation([3]string{"ends", "format", "YYYY-MM-DD"})
      case strings.Contains(err.Error(), "out of range"):
        return problem.NewValueOutOfRange("date", "ends", update.Ends)
      }
    }

    update.Ends = tm.Format(time.DateOnly)
  }

  var projectText strings.Builder

  if "" != update.Name {
    projectText.WriteString(update.Name)
  }

  if "" != update.Summary {
    sanitizeTextWordIntersections(&update.Summary)
    if 60 < wordsIn(update.Summary) {
      return problem.NewValidation([3]string{"summary", "max", "60"})
    }

    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(update.Summary)
  }

  if "" != update.Content {
    if 0 < projectText.Len() {
      projectText.WriteRune('\n')
    }
    projectText.WriteString(update.Content)
  }

  if 0 < projectText.Len() {
    var r = strings.NewReader(projectText.String())
    update.ReadTime = computePostReadingTimeInMinutes(r)
  }

  if "" != update.Name {
    update.Slug = generateSlug(update.Name)
  }

  return s.r.Update(ctx, id, update)
}

// Unarchive makes a project not archived so that it can be normally listed.
func (s *ProjectsService) Unarchive(ctx context.Context, id string) error {
  if err := validateUUID(&id); err != nil {
    return err
  }
  return s.r.SetArchived(ctx, id, false)
}

// Remove deletes an existing project. If not found, returns a not found error.
func (s *ProjectsService) Remove(ctx context.Context, id string) error {
  if err := validateUUID(&id); err != nil {
    return err
  }
  return s.r.Remove(ctx, id)
}

// HasTag checks whether technologyTagID belongs to projectID.
func (s *ProjectsService) HasTag(ctx context.Context, projectID, technologyTagID string) (success bool, err error) {
  if err = validateUUID(&projectID); err != nil {
    return false, err
  }
  if err = validateUUID(&technologyTagID); err != nil {
    return false, err
  }
  return s.r.HasTag(ctx, projectID, technologyTagID)
}

// AddTag adds an existing technology tag that will belong to the project represented by projectID .
func (s *ProjectsService) AddTag(ctx context.Context, projectID, technologyTagID string) error {
  if err := validateUUID(&projectID); err != nil {
    return err
  }
  if err := validateUUID(&technologyTagID); err != nil {
    return err
  }
  if err := s.Exists(ctx, projectID); nil != err {
    return err
  }

  if err := s.tags.Exists(ctx, technologyTagID); nil != err {
    return err
  }

  conflict, err := s.HasTag(ctx, projectID, technologyTagID)
  if nil != err {
    return err
  }
  if conflict {
    var p problem.Problem
    p.Type(problem.TypeDuplicateKey)
    p.Status(http.StatusConflict)
    p.Title("Duplicate technology tag.")
    p.Detail("The specified technology tag is already associated with this project. Try using a different one.")
    p.With("project_id", projectID)
    p.With("technology_tag_id", projectID)
    return &p
  }
  return s.r.AddTag(ctx, projectID, technologyTagID)
}

// RemoveTag removes a technology tag that belongs to the project represented by projectID.
func (s *ProjectsService) RemoveTag(ctx context.Context, projectID, technologyTagID string) error {
  if err := validateUUID(&projectID); err != nil {
    return err
  }
  if err := validateUUID(&technologyTagID); err != nil {
    return err
  }
  if err := s.Exists(ctx, projectID); nil != err {
    return err
  }
  if err := s.tags.Exists(ctx, technologyTagID); nil != err {
    return err
  }
  return s.r.RemoveTag(ctx, projectID, technologyTagID)
}
