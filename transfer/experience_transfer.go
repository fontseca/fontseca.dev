package transfer

// ExperienceCreation represents the data required to create a new experience entry.
type ExperienceCreation struct {
  Starts          string `json:"starts" binding:"required"`
  Ends            string `json:"ends"`
  JobTitle        string `json:"job_title" binding:"required,max=64"`
  Company         string `json:"company" binding:"required,max=64"`
  CompanyHomepage string `json:"company_homepage" binding:"max=2048"`
  Country         string `json:"country" binding:"required,max=64"`
  Summary         string `json:"summary" binding:"required"`
}

// ExperienceUpdate represents the data required to update an existing experience entry.
type ExperienceUpdate struct {
  Starts          string `json:"starts"`
  Ends            string `json:"ends"`
  JobTitle        string `json:"job_title" binding:"max=64"`
  Company         string `json:"company" binding:"max=64"`
  CompanyHomepage string `json:"company_homepage" binding:"max=2048"`
  Country         string `json:"country" binding:"max=64"`
  Summary         string `json:"summary"`
  Active          bool   `json:"-"`
  Hidden          bool   `json:"-"`
}
