package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "strconv"
  "strings"
  "time"
)

type experienceRepositoryAPI interface {
  List(context.Context, bool) ([]*model.Experience, error)
  Get(context.Context, string) (*model.Experience, error)
  Create(context.Context, *transfer.ExperienceCreation) (bool, error)
  Update(context.Context, string, *transfer.ExperienceUpdate) (bool, error)
  Remove(context.Context, string) error
}

// ExperienceService provides methods for interacting with
// experience data at a higher level.
type ExperienceService struct {
  r experienceRepositoryAPI
}

func NewExperienceService(r experienceRepositoryAPI) *ExperienceService {
  return &ExperienceService{r}
}

// List returns a slice of experience models and an error if
// the operation fails. If hidden is true it returns all the
// hidden experience records.
func (s *ExperienceService) List(ctx context.Context, hidden ...bool) (experience []*model.Experience, err error) {
  if 0 != len(hidden) && hidden[0] {
    return s.r.List(ctx, true)
  }
  return s.r.List(ctx, false)
}

// Get retrieves a single experience record by its UUID.
func (s *ExperienceService) Get(ctx context.Context, id string) (experience *model.Experience, err error) {
  if err = validateUUID(&id); nil != err {
    return nil, err
  }
  return s.r.Get(ctx, id)
}

// Create creates a new experience record with the provided creation data.
// It returns a boolean indicating whether the experience was successfully
// saved and an error if something went wrong.
func (s *ExperienceService) Create(ctx context.Context, creation *transfer.ExperienceCreation) (saved bool, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return false, err
  }
  creation.JobTitle = strings.TrimSpace(creation.JobTitle)
  creation.Company = strings.TrimSpace(creation.Company)
  creation.Country = strings.TrimSpace(creation.Country)
  creation.Summary = strings.TrimSpace(creation.Summary)

  var year = time.Now().Year()
  switch {
  case 0 != creation.Starts && 2017 >= creation.Starts:
    return false, problem.NewValidation([3]string{"starts", "gt", "2017"})
  case 0 != creation.Starts && year < creation.Starts:
    return false, problem.NewValidation([3]string{"starts", "lte", strconv.Itoa(year)})
  case 0 != creation.Ends && creation.Starts > creation.Ends:
    return false, problem.NewValidation([3]string{"ends", "gte", strconv.Itoa(creation.Starts)})
  case 0 != creation.Ends && year < creation.Ends:
    return false, problem.NewValidation([3]string{"ends", "lte", strconv.Itoa(year)})
  }

  return s.r.Create(ctx, creation)
}

// Update modifies an existing experience record with the provided update data.
// It returns a boolean indicating whether the experience was successfully updated
// and an error if something went wrong.
func (s *ExperienceService) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) (updated bool, err error) {
  if nil == update {
    err = errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return false, err
  }
  if err = validateUUID(&id); nil != err {
    return false, err
  }

  update.JobTitle = strings.TrimSpace(update.JobTitle)
  update.Company = strings.TrimSpace(update.Company)
  update.Country = strings.TrimSpace(update.Country)
  update.Summary = strings.TrimSpace(update.Summary)

  var year = time.Now().Year()
  switch {
  case 0 != update.Starts && 2017 >= update.Starts:
    return false, problem.NewValidation([3]string{"starts", "gt", "2017"})
  case 0 != update.Starts && year < update.Starts:
    return false, problem.NewValidation([3]string{"starts", "lte", strconv.Itoa(year)})
  case 0 != update.Ends:
    switch {
    case 0 != update.Starts && update.Starts > update.Ends:
      return false, problem.NewValidation([3]string{"ends", "gte", strconv.Itoa(update.Starts)})
    case 2017 >= update.Ends:
      return false, problem.NewValidation([3]string{"ends", "gt", "2017"})
    case year < update.Ends:
      return false, problem.NewValidation([3]string{"ends", "lte", strconv.Itoa(year)})
    }
  }

  return s.r.Update(ctx, id, update)
}

// Remove deletes an experience record by its UUID.
// It returns an error if the operation fails; for example,
// if the record does not exist.
func (s *ExperienceService) Remove(ctx context.Context, id string) error {
  if err := validateUUID(&id); err != nil {
    return err
  }
  return s.r.Remove(ctx, id)
}
