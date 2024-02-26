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

func Test_validateUUID(t *testing.T) {
  t.Run("success", func(t *testing.T) {
    var expected = "d0c97bc8-ae21-4f12-8e5f-7c1d97c4538a"
    var uuids = []string{
      "{d0c97bc8-ae21-4f12-8e5f-7c1d97c4538a}",
      "urn:uuid:d0c97bc8-ae21-4f12-8e5f-7c1d97c4538a",
      "d0c97bc8ae214f128e5f7c1d97c4538a",
      "d0c97bc8-ae21-4f12-8e5f-7c1d97c4538a",
      "  \n\n\t\td0c97bc8-ae21-4f12-8e5f-7c1d97c4538a\t\t\n\n  ",
    }
    for _, id := range uuids {
      err := validateUUID(&id)
      assert.NoError(t, err)
      assert.Equal(t, expected, id)
    }
  })

  t.Run("error", func(t *testing.T) {
    var uuids = []string{
      "d0c97bc8-ae21-4f12-8e5f",              // invalid length
      "d0c97bc8-ae21-4f12-8e5f-7c1d97c4538z", // invalid format
    }
    for _, id := range uuids {
      err := validateUUID(&id)
      assert.ErrorContains(t, err, "An error occurred while attempting to parse the provided UUID string.")
      assert.Empty(t, id)
    }
  })
}
