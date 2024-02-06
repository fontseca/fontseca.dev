package transfer

// MeUpdate represents the data structure used for updating my profile information.
type MeUpdate struct {
  Summary      string `json:"summary"`
  JobTitle     string `json:"job_title"`
  Email        string `json:"email"`
  Company      string `json:"company"`
  Location     string `json:"location"`
  Hireable     bool   `json:"hireable"`
  GitHubURL    string `json:"github_url"`
  LinkedInURL  string `json:"linkedin_url"`
  YouTubeURL   string `json:"youtube_url"`
  TwitterURL   string `json:"twitter_url"`
  InstagramURL string `json:"instagram_url"`
}
