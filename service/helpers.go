package service

import (
  "bufio"
  "bytes"
  "fontseca/problem"
  "github.com/google/uuid"
  "io"
  "log/slog"
  "math"
  "net/http"
  "net/url"
  "strings"
  "time"
  "unicode"
  "unicode/utf8"
)

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

func computePostReadingTimeInMinutes(r io.Reader) int {
  duration, err := computePostReadingTime(r, 183.0)
  if err != nil {
    return 0
  }
  return int(math.Ceil(duration.Minutes()))
}
