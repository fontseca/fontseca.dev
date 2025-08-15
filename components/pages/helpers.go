package pages

import (
  "fmt"
  "strings"

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

func fmtNumber(x int64) string {
  s := fmt.Sprintf("%d", x)
  var result strings.Builder

  length := len(s)
  for i, c := range s {
    if i != 0 && (length-i)%3 == 0 {
      result.WriteByte(',')
    }

    result.WriteByte(byte(c))
  }

  return result.String()
}
