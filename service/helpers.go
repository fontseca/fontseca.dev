package service

import (
  "bufio"
  "bytes"
  "fontseca.dev/problem"
  "github.com/google/uuid"
  "io"
  "log/slog"
  "math"
  "net/http"
  "net/url"
  "regexp"
  "strings"
  "time"
  "unicode"
  "unicode/utf8"
)

var (
  contiguousSpacesRegexp = regexp.MustCompile(`\s+`)
  wordsOnly              = regexp.MustCompile(`\w+`)
)

// sanitizeTextWordIntersections removes contiguous whitespace characters from the given text,
// reducing them to a single space.
func sanitizeTextWordIntersections(text *string) {
  if nil == text {
    return
  }
  *text = contiguousSpacesRegexp.ReplaceAllString(*text, " ")
}

// toKebabCase converts a given text to kebab-case by replacing spaces and underscores with hyphens,
// making all letters lowercase, and removing any non-word characters.
func toKebabCase(text string) string {
  if "" == text {
    return ""
  }

  text = strings.ToLower(text)

  if strings.Contains(text, "_") {
    text = strings.ReplaceAll(text, "_", "-")
  }

  words := wordsOnly.FindAllString(text, -1)

  return strings.Join(words, "-")
}

// generateSlug creates a URL-friendly slug by converting the source text to kebab-case.
func generateSlug(source string) string {
  return toKebabCase(source)
}

// wordsIn counts the number of words in a given text based on whitespace splitting.
func wordsIn(text string) int {
  fields := strings.Fields(text)
  return len(fields)
}

// sanitizeURL checks if the provided URLs are in a valid format. If valid, it
// trims whitespace and returns the sanitized URL(s). If invalid, it returns a problem
// detailing the URL error.
func sanitizeURL(urls ...*string) error {
  if 0 == len(urls) {
    return nil
  }

  for _, u := range urls {
    if nil == u {
      continue
    }

    *u = strings.TrimSpace(*u)
    if "" == *u {
      continue
    }

    uri, err := url.ParseRequestURI(*u)
    if nil != err {
      var p problem.Problem
      p.Type("unparseable_value")
      p.Title("Unprocessable URL format.")
      p.Status(http.StatusUnprocessableEntity)
      p.Detail("There was an error parsing the requested URL. Please try with a different URL or verify the current one for correctness.")
      p.With("wrong_url", *u)
      return &p
    }
    *u = uri.String()
  }

  return nil
}

// validateUUID checks if a string is a valid UUID format, trimming whitespace
// and standardizing it if valid. If invalid, it sets the UUID to an empty string
// and returns an error describing the problem.
func validateUUID(id *string) error {
  if nil == id {
    return nil
  }
  *id = strings.TrimSpace(*id)
  parsed, err := uuid.Parse(*id)
  if nil != err {
    var p problem.Problem
    p.Title("Could not parse UUID.")
    p.Status(http.StatusUnprocessableEntity)
    p.Detail("An error occurred while attempting to parse the provided UUID string.")
    p.With("uuid", *id)
    p.With("error", err.Error())
    *id = ""
    return &p
  }
  *id = parsed.String()
  return nil
}

// approximatePostWordsCount counts the approximate number of words in HTML or text content
// while ignoring specific HTML elements like <figure> and nested <div> tags.
func approximatePostWordsCount(r io.Reader) (words int, err error) {
  data, err := io.ReadAll(r)
  if nil != err {
    slog.Error(err.Error())
    return 0, err
  }

  var (
    insideFigure bool
    divDepth     int
  )

  for len(data) > 0 {
    var (
      advance int
      word    []byte
    )

    // Skip leading spaces.
    var start = 0
    for width := 0; start < len(data); start += width {
      var rr rune
      rr, width = utf8.DecodeRune(data[start:])
      if !unicode.IsSpace(rr) {
        break
      }
    }

    // If it's empty data.
    if bytes.Equal(data[start:], []byte("")) {
      break
    }

    // Scan until a space if found, signaling end of word or data is empty.
    for width, i := 0, start; i < len(data); i += width {
      var rr rune
      rr, width = utf8.DecodeRune(data[i:])
      if unicode.IsSpace(rr) {
        advance = i + width
        word = data[start:i]
        break
      }

      // If data length is 0, then we're in the last word.
      if len(data[i+width:]) == 0 {
        advance = len(data)
        word = data[start:]
      }
    }

    // Slice to the remaining of the data of interest.
    data = data[advance:]

    if bytes.HasPrefix(word, []byte("<figure>")) {
      insideFigure = true
      continue
    }

    if bytes.HasSuffix(word, []byte("</figure>")) {
      insideFigure = false
      continue
    }

    if insideFigure {
      continue
    }

    if bytes.Equal(word, []byte("<div>")) || bytes.Equal(word, []byte("<div")) {
      divDepth++
      continue
    }

    if bytes.Equal(word, []byte("</div>")) {
      divDepth--
      continue
    }

    if divDepth > 0 {
      continue
    }

    if bytes.HasPrefix(word, []byte("#")) ||
      bytes.HasPrefix(word, []byte("-")) ||
      bytes.HasPrefix(word, []byte("=")) ||
      bytes.HasPrefix(word, []byte(">")) ||
      bytes.Equal(word, []byte("*")) {
      continue
    }

    words++
  }

  return words, nil
}

// computePostReadingTime estimates the reading time of a post based on the word count
// and a specified words-per-minute rate.
func computePostReadingTime(r io.Reader, wordsPerMinute float64) (readingTime time.Duration, err error) {
  count, err := approximatePostWordsCount(bufio.NewReader(r))
  if nil != err {
    return 0, err
  }
  var totalWords = float64(count)
  var wordsPerSecond = wordsPerMinute / 60
  readingTime = time.Duration(math.Ceil(totalWords/wordsPerSecond)) * time.Second
  return readingTime, nil
}

// computePostReadingTimeInMinutes calculates the approximate reading time in whole minutes,
// assuming an average reading speed of 183 words per minute.
func computePostReadingTimeInMinutes(r io.Reader) int {
  duration, err := computePostReadingTime(r, 183.0)
  if err != nil {
    return 0
  }
  return int(math.Ceil(duration.Minutes()))
}
