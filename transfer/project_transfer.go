package transfer

// ProjectCreation represents the data required to create a new project entry.
type ProjectCreation struct {
  Name           string `json:"name" binding:"required,max=64"`
  Slug           string `json:"slug"`
  Homepage       string `json:"homepage"`
  Language       string `json:"language"`
  Summary        string `json:"summary"`
  Content        string `json:"content"`
  EstimatedTime  int    `json:"estimated_time"`
  FirstImageURL  string `json:"first_image_url"`
  SecondImageURL string `json:"second_image_url"`
  GitHubURL      string `json:"github_url"`
  CollectionURL  string `json:"collection_url"`
}

// ProjectUpdate represents the data required to update an existing project entry.
type ProjectUpdate struct {
  Name           string `json:"name"`
  Homepage       string `json:"homepage"`
  Language       string `json:"language"`
  Summary        string `json:"summary"`
  Content        string `json:"content"`
  EstimatedTime  int    `json:"estimated_time"`
  FirstImageURL  string
  SecondImageURL string
  GitHubURL      string
  CollectionURL  string
  PlaygroundURL  string
  Archived       bool
  Finished       bool
}
