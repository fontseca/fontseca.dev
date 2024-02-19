package transfer

// MeUpdate represents the data structure used for updating my profile information.
type MeUpdate struct {
  Summary      string `json:"summary" binding:"max=1024"`
  JobTitle     string `json:"job_title" binding:"max=64"`
  Email        string `json:"email" binding:"max=254"`
  PhotoURL     string `json:"-"`
  ResumeURL    string `json:"-"`
  Company      string `json:"company" binding:"max=64"`
  Location     string `json:"location" binding:"max=64"`
  Hireable     bool   `json:"-"`
  GitHubURL    string `json:"github_url" binding:"max=2048"`
  LinkedInURL  string `json:"linkedin_url" binding:"max=2048"`
  YouTubeURL   string `json:"youtube_url" binding:"max=2048"`
  TwitterURL   string `json:"twitter_url" binding:"max=2048"`
  InstagramURL string `json:"instagram_url" binding:"max=2048"`
}
