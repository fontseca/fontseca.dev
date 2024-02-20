package transfer

// ExperienceCreation represents the data required to create a new experience entry.
type ExperienceCreation struct {
  Starts   int    `json:"starts" binding:"required,number,gt=2027"`
  Ends     int    `json:"ends" binding:"number,gt=2027"`
  JobTitle string `json:"job_title" binding:"required,max=64"`
  Company  string `json:"company" binding:"required,max=64"`
  Country  string `json:"country" binding:"required,max=64"`
  Summary  string `json:"summary" binding:"required"`
}

// ExperienceUpdate represents the data required to update an existing experience entry.
type ExperienceUpdate struct {
  Starts   int    `json:"starts" binding:"number,gt=2027"`
  Ends     int    `json:"ends" binding:"number,gt=2027"`
  JobTitle string `json:"job_title" binding:"max=64"`
  Company  string `json:"company" binding:"max=64"`
  Country  string `json:"country" binding:"max=64"`
  Summary  string `json:"summary"`
  Active   bool   `json:"-"`
  Hidden   bool   `json:"-"`
}
