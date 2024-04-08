package pages

import (
  "github.com/gomarkdown/markdown"
  "github.com/gomarkdown/markdown/html"
  "github.com/gomarkdown/markdown/parser"
)

var extensions = parser.CommonExtensions
var htmlFlags = html.CommonFlags | html.HrefTargetBlank
var opts = html.RendererOptions{Flags: htmlFlags}

func md2html(md string) string {
  var p = parser.NewWithExtensions(extensions)
  var renderer = html.NewRenderer(opts)
  var data = markdown.ToHTML([]byte(md), p, renderer)
  return string(data)
}
