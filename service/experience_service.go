package service

import (
  "context"
  "errors"
  "fontseca.dev/model"
  "fontseca.dev/problem"
  "fontseca.dev/transfer"
  "log/slog"
  "strings"
  "time"
)

type experienceRepositoryAPI interface {
  List(context.Context, bool) ([]*model.Experience, error)
  Get(context.Context, string, bool) (*model.Experience, error)
  Create(context.Context, *transfer.ExperienceCreation) (string, error)
  Update(context.Context, string, *transfer.ExperienceUpdate) error
  SetHidden(ctx context.Context, id string, hidden bool) error
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
  return s.r.Get(ctx, id, true)
}

// Create creates a new experience record with the provided creation data.
// It returns a boolean indicating whether the experience was successfully
// saved and an error if something went wrong.
func (s *ExperienceService) Create(ctx context.Context, creation *transfer.ExperienceCreation) (created string, err error) {
  if nil == creation {
    err = errors.New("nil value for parameter: creation")
    slog.Error(err.Error())
    return "", err
  }
  creation.JobTitle = strings.TrimSpace(creation.JobTitle)
  creation.Company = strings.TrimSpace(creation.Company)
  creation.CompanyHomepage = strings.TrimSpace(creation.CompanyHomepage)
  creation.Starts = strings.TrimSpace(creation.Starts)
  creation.Ends = strings.TrimSpace(creation.Ends)
  creation.Country = strings.TrimSpace(creation.Country)
  creation.Summary = strings.TrimSpace(creation.Summary)

  sanitizeTextWordIntersections(&creation.JobTitle)
  sanitizeTextWordIntersections(&creation.Company)

  if "" != creation.CompanyHomepage {
    if 2048 < len(creation.CompanyHomepage) {
      return "", problem.NewValidation([3]string{"company_homepage", "max", "2048"})
    }

    err := sanitizeURL(&creation.CompanyHomepage)
    if nil != err {
      return "", err
    }

  }

  var now = time.Now().UTC()
  if "" != creation.Starts {
    start, err := time.Parse(time.DateOnly, creation.Starts)
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

    switch {
    case 2017 >= start.Year():
      return "", problem.NewValidation([3]string{"starts", "lte", "year:2017"})
    case now.Before(start):
      return "", problem.NewValidation([3]string{"starts", "lt", now.Format(time.DateOnly)})
    }

    creation.Starts = start.Format(time.DateOnly)
  }

  if "" != creation.Ends {
    end, err := time.Parse(time.DateOnly, creation.Ends)
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

    start, _ := time.Parse(time.DateOnly, creation.Starts)
    switch {
    case start.After(end):
      return "", problem.NewValidation([3]string{"ends", "lt", creation.Starts})
    case now.Before(end):
      return "", problem.NewValidation([3]string{"ends", "lt", now.Format(time.DateOnly)})
    }

    creation.Ends = end.Format(time.DateOnly)
  }

  return s.r.Create(ctx, creation)
}

// Hide hides an experience.
func (s *ExperienceService) Hide(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }
  return s.r.SetHidden(ctx, id, true)
}

// Show shows a hidden experience.
func (s *ExperienceService) Show(ctx context.Context, id string) error {
  if err := validateUUID(&id); nil != err {
    return err
  }
  return s.r.SetHidden(ctx, id, false)
}

// Update modifies an existing experience record with the provided update data.
// It returns a boolean indicating whether the experience was successfully updated
// and an error if something went wrong.
func (s *ExperienceService) Update(ctx context.Context, id string, update *transfer.ExperienceUpdate) error {
  if nil == update {
    err := errors.New("nil value for parameter: update")
    slog.Error(err.Error())
    return err
  }
  if err := validateUUID(&id); nil != err {
    return err
  }

  update.JobTitle = strings.TrimSpace(update.JobTitle)
  update.Company = strings.TrimSpace(update.Company)
  update.Starts = strings.TrimSpace(update.Starts)
  update.Ends = strings.TrimSpace(update.Ends)
  update.CompanyHomepage = strings.TrimSpace(update.CompanyHomepage)
  update.Country = strings.TrimSpace(update.Country)
  update.Summary = strings.TrimSpace(update.Summary)

  sanitizeTextWordIntersections(&update.JobTitle)
  sanitizeTextWordIntersections(&update.Company)

  if "" != update.CompanyHomepage {
    if 2048 < len(update.CompanyHomepage) {
      return problem.NewValidation([3]string{"company_homepage", "max", "2048"})
    }

    err := sanitizeURL(&update.CompanyHomepage)
    if nil != err {
      return err
    }
  }

  var now = time.Now().UTC()
  if "" != update.Starts {
    start, err := time.Parse(time.DateOnly, update.Starts)
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

    switch {
    case 2017 >= start.Year():
      return problem.NewValidation([3]string{"starts", "lte", "year:2017"})
    case now.Before(start):
      return problem.NewValidation([3]string{"starts", "lt", now.Format(time.DateOnly)})
    }

    update.Starts = start.Format(time.DateOnly)
  }

  if "" != update.Ends && "2006-01-02" != update.Ends {
    end, err := time.Parse(time.DateOnly, update.Ends)
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

    switch {
    case 2017 >= end.Year():
      return problem.NewValidation([3]string{"starts", "lte", "year:2017"})
    case now.Before(end):
      return problem.NewValidation([3]string{"ends", "lt", now.Format(time.DateOnly)})
    }

    e, err := s.r.Get(ctx, id, false)
    if nil != err {
      return err
    }

    start := e.Starts
    switch {
    case start.After(end):
      return problem.NewValidation([3]string{"ends", "lt", start.Format(time.DateOnly)})
    }

    update.Ends = end.Format(time.DateOnly)
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
