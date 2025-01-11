package transfer

// OG represents metadata for the Open Graph protocol, which is used to optimize web pages
// for social media sharing by embedding structured information about the content.
// More information on the protocol can be found at https://ogp.me/.
type OG struct {
  Description string
  URL         string
  ImageURL    string
  ImageAlt    string
  Type        string

  ArticlePublishedTime string
  ArticleAuthor        string
  ArticlePublisher     string
}
