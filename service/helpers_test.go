package service

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_sanitizeURL(t *testing.T) {
  t.Run("errors on wrong urls", func(t *testing.T) {
    var urls = []string{
      "  \t\n . \n\t  ",
      "picsum.photos/200/300",
      "foo/bar/",
      "gotlim.com",
    }

    for _, url := range urls {
      err := sanitizeURL(&url)
      assert.Error(t, err, "URL was: %q", url)
    }
  })
}
