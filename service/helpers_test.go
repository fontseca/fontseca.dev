package service

import (
  "bytes"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "os"
  "strings"
  "testing"
  "time"
)

func Test_sanitizeTextWordIntersections(t *testing.T) {
  text := "a             b c\n\t  d \t\n e"
  sanitizeTextWordIntersections(&text)
  assert.Equal(t, "a b c d e", text)
}

func Test_generateSlug(t *testing.T) {
  sources := [][2]string{
    {"Quisque egestas cursus.", "quisque-egestas-cursus"},
    {"Nisi est sit_amet facilisis!", "nisi-est-sit-amet-facilisis"},
    {"Dolor magna eget?", "dolor-magna-eget"},
    {"20 VIVERRA adipiscing 60 at", "20-viverra-adipiscing-60-at"},
  }

  for _, source := range sources {
    out := generateSlug(source[0])
    assert.Equal(t, source[1], out)
  }
}

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

type textAndWords struct {
  content    string
  wordsCount int
}

func Test_approximatePostWordsCount(t *testing.T) {
  var texts = []textAndWords{
    {
      "" +
        "# Title " +
        "--- " +
        "## Subtitle 1 " +
        "### Subtitle 2 " +
        "#### Subtitle 3 " +
        "##### Subtitle 4 " +
        "=== " +
        "**Aliquet eget sit amet** tellus cras *adipiscing enim eu*. Ligula ullamcorper " +
        "malesuada _proin libero_ nunc consequat interdum. `Accumsan in nisl nisi " +
        "scelerisque eu`.",
      32,
    },
    {
      `>
       > Quibusdam soluta ratione.
       >> Condimentum vitae sapien pellentesque habitant morbi.
       >>> Vestibulum mattis ullamcorper velit.
       >>>> Magna fermentum iaculis eu non diam.
       >>>>> At varius vel pharetra vel turpis
       > Praesentium consectetur.
       >`,
      27,
    },
    { // Ignores everything inside a <figure> tag.
      `Reprehenderit quia ut.
    
       <figure>
         <video
           controls
           loop
           src="video.mp4"
           width="100%"
           height="auto"
         ></video>
         <figcaption>
           <div class="caption">
             <p>Quis commodo odio aenean sed adipiscing.</p>
           </div>
         </figcaption>
       </figure>`,
      3,
    },
    {
      `
       <div class="class name">
         <div>
           <div id="div_id">
             <div>
               <p>This text will be ignored!</p>
             </div>
           </div>
         </div>
       </div>`,
      0,
    },
    {
      `Quisquam maiores.
    
       <figure>
         <video
           controls
           loop
           src="video.mp4"
           width="100%"
           height="auto"
         ></video>
         <figcaption>
           <div class="caption">
             <p>Quis commodo odio aenean sed adipiscing.</p>
           </div>
         </figcaption>
       </figure>
    
       <p>Repellendus nobis!</p>`,
      4,
    },
    {
      " \n\t\n ",
      0,
    },
    {
      "<figure></figure>",
      0,
    },
    {
      "<figure> a b c </figure>",
      0,
    },
    {
      "a b c </figure>",
      3,
    },
    {
      "a b c </div> </div> </div>",
      3,
    },
  }

  for _, text := range texts {
    var count, err = approximatePostWordsCount(strings.NewReader(text.content))
    assert.NoError(t, err)
    assert.Equal(t, text.wordsCount, count)
  }
}

func Test_computePostReadingTime(t *testing.T) {
  var text = []byte(`
# Vulputate enim nulla aliquet porttitor lacus.

---

Deleniti culpa  aspernatur nihil similique dicta, esse non veniam itaque illo
consectetur, recusandae hic _repellat veritatis commodi_ quam possimus tenetur
doloribus. Iusto id, consequuntur necessitatibus explicabo alias voluptates ad
voluptatem tempora maiores doloremque rem nobis illum dicta saepe nihil! Nulla
mollitia possimus in repellat maxime, praesentium consectetur dignissimos
minus exercitationem ad voluptate!

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Concepts](#concepts)
- [Alternatives](#alternatives)
- [Contributing](#contributing)

## Features

- Sunt quas  unde deleniti laborum obcaecati officiis nesciunt repellendus quis placeat.
- Similique doloribus saepe corrupti autem dolore, vero eligendi minus.
- Totam laborum, ipsa culpa quas ad quod debitis incidunt nostrum.
- Reprehenderit quia ut id provident rerum, blanditiis minus.
- Tempore, atque quisquam maiores pariatur architecto.

<figure>
  <video
    controls
    loop
    src="video.mp4"
    width="100%"
    height="auto"
  ></video>
  <figcaption>
    <div class="caption">
      <p>Quis commodo odio aenean sed adipiscing.</p>
    </div>
  </figcaption>
</figure>

## Installation

Aliquet eget sit amet tellus cras adipiscing enim eu. Ligula ullamcorper
malesuada proin libero nunc consequat interdum. Accumsan in nisl nisi
scelerisque eu. Tortor at auctor urna nunc id cursus metus. Dui nunc mattis enim
ut tellus elementum sagittis vitae. Amet mattis vulputate enim nulla. Elit sed
vulputate mi sit. Consequat mauris nunc congue nisi vitae suscipit tellus mauris
a. Duis convallis convallis tellus id interdum velit laoreet id. Viverra orci
sagittis eu volutpat odio facilisis mauris. Nibh mauris cursus mattis molestie.
Vitae justo eget magna fermentum. Turpis egestas pretium aenean pharetra. Ac
orci phasellus egestas tellus. Vestibulum mattis ullamcorper velit sed
ullamcorper.

## Usage

Habitant morbi tristique senectus et netus. Cursus eget nunc scelerisque
viverra. Suspendisse interdum consectetur libero id faucibus nisl tincidunt
eget. A diam maecenas sed enim ut sem. Enim blandit volutpat maecenas
volutpat. **Arcu non sodales neque sodales ut etiam sit amet.** Ut tellus
elementum sagittis vitae et. Vel eros donec ac odio tempor orci. Sed
euismod nisi porta lorem mollis aliquam ut. Nisl condimentum id venenatis
a. Id diam maecenas ultricies mi eget mauris pharetra et ultrices.
Id diam maecenas ultricies mi eget. Vel turpis nunc eget lorem dolor sed
viverra ipsum nunc. Vitae turpis massa sed elementum tempus egestas sed
sed risus. Egestas pretium aenean pharetra magna. Enim nunc faucibus a
pellentesque.

Duis convallis convallis tellus id interdum velit laoreet id. _Viverra orci
sagittis eu volutpat odio facilisis mauris._ Nibh mauris cursus mattis molestie.
Vitae justo eget magna fermentum. Turpis egestas pretium aenean pharetra. Ac
orci phasellus egestas tellus. Vestibulum mattis ullamcorper velit sed
ullamcorper.

<div class="snippet">
  <div class="options">
    <div>
      <p>Go</p>
    </div>
  </div>
    <pre><code>package time

// A Ticker holds a channel that delivers “ticks” of a clock
// at intervals.
type Ticker struct {
	C <-chan Time // The channel on which the ticks are delivered.
	r runtimeTimer
}

// NewTicker returns a new Ticker containing a channel that will send
// the current time on the channel after each tick. The period of the
// ticks is specified by the duration argument. The ticker will adjust
// the time interval or drop ticks to make up for slow receivers.
// The duration d must be greater than zero; if not, NewTicker will
// panic. Stop the ticker to release associated resources.
func NewTicker(d Duration) *Ticker {
	if d <= 0 {
		panic("non-positive interval for NewTicker")
	}
	// Give the channel a 1-element time buffer.
	// If the client falls behind while reading, we drop ticks
	// on the floor until the client catches up.
	c := make(chan Time, 1)
	t := &Ticker{
		C: c,
		r: runtimeTimer{
			when:   when(d),
			period: int64(d),
			f:      sendTime,
			arg:    c,
		},
	}
	startTimer(&t.r)
	return t
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
// Stop does not close the channel, to prevent a concurrent goroutine
// reading from the channel from seeing an erroneous "tick".
func (t *Ticker) Stop() {
	stopTimer(&t.r)
}

// Reset stops a ticker and resets its period to the specified duration.
// The next tick will arrive after the new period elapses. The duration d
// must be greater than zero; if not, Reset will panic.
func (t *Ticker) Reset(d Duration) {
	if d <= 0 {
		panic("non-positive interval for Ticker.Reset")
	}
	if t.r.f == nil {
		panic("time: Reset called on uninitialized Ticker")
	}
	modTimer(&t.r, when(d), int64(d), t.r.f, t.r.arg, t.r.seq)
}

// Tick is a convenience wrapper for NewTicker providing access to the ticking
// channel only. While Tick is useful for clients that have no need to shut down
// the Ticker, be aware that without a way to shut it down the underlying
// Ticker cannot be recovered by the garbage collector; it "leaks".
// Unlike NewTicker, Tick will return nil if d <= 0.
func Tick(d Duration) <-chan Time {
	if d <= 0 {
		return nil
	}
	return NewTicker(d).C
}</code></pre>
</div>

Est pellentesque elit ullamcorper dignissim. Nisi est sit amet facilisis magna
etiam tempor. ~~Elit duis tristique sollicitudin nibh. Felis donec et odio
pellentesque diam volutpat commodo.~~ Dolor morbi non arcu risus quis varius.
Condimentum vitae sapien pellentesque habitant morbi tristique senectus et
netus. Vestibulum mattis ullamcorper velit sed ullamcorper morbi tincidunt.
Magna fermentum iaculis eu non diam. **Neque sodales ut etiam sit amet nisl
purus. At varius vel pharetra vel turpis.** Eget nunc lobortis mattis aliquam
faucibus purus in massa tempor. Lectus magna fringilla urna porttitor rhoncus
dolor.

<figure>
  <img
    src="image.jpg"
    alt="Eget nunc lobortis mattis aliquam faucibus purus in massa tempor."
  />
  <figcaption>
    <div class="caption">
      <p>
        Quis commodo odio aenean sed adipiscing. Ullamcorper sit amet risus
        nullam eget felis eget nunc lobortis.
      </p>
    </div>
  </figcaption>
</figure>

### Concepts

Vestibulum mattis ullamcorper velit sed ullamcorper morbi tincidunt.
Magna fermentum iaculis eu non diam. **Neque sodales ut etiam sit amet nisl
purus. At varius vel pharetra vel turpis.** Eget nunc lobortis mattis aliquam
faucibus purus in massa tempor. Lectus magna fringilla urna porttitor rhoncus
dolor.`)
  var duration time.Duration
  var err error
  var readTime = "2m47s"

  t.Run("reading time from a string", func(t *testing.T) {
    duration, err = computePostReadingTime(bytes.NewReader(text), 183.0)
    assert.NoError(t, err)
    assert.Equal(t, readTime, duration.String())
  })

  t.Run("reading time from a file", func(t *testing.T) {
    var handle *os.File
    handle, err = os.CreateTemp("", "sample.md")
    require.NoError(t, err)
    defer handle.Close()
    defer os.Remove(handle.Name())

    _, err = handle.Write(text)
    require.NoError(t, err)

    _, err = handle.Seek(0, 0)
    require.NoError(t, err)

    duration, err = computePostReadingTime(handle, 183.0)
    assert.NoError(t, err)

    assert.Equal(t, readTime, duration.String())
    require.NoError(t, handle.Close())
  })
}
