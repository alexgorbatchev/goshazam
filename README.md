# goshazam

[![Go Reference](https://pkg.go.dev/badge/github.com/alexgorbatchev/goshazam.svg)](https://pkg.go.dev/github.com/alexgorbatchev/goshazam)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexgorbatchev/goshazam)](https://goreportcard.com/report/github.com/alexgorbatchev/goshazam)

A pure Go port of [ShazamIO](https://github.com/dotX12/ShazamIO), the framework for the reverse-engineered **Shazam API** and audio fingerprinting algorithm.

- **100% Pure Go DSP**: Zero CGO, zero external audio processing dependencies.
- **Audio Recognition**: Identify songs from files (`.mp3`, `.ogg`, `.wav`, `.flac`, `.m4a`, etc.), in-memory audio bytes, `io.Reader` streams, or raw PCM.
- **Pure Go Fingerprinting**: Built-in Cooley-Tukey Radix-2 Real FFT, Hanning windowing, frequency/time-domain peak spreading, parabolic interpolation, and binary/URI signature serialization with CRC-32 IEEE verification.
- **Full Shazam Catalog APIs**: Top charts (world, country, city, genre), artist profiles, album views, related song recommendations, and listening counters.
- **Resilient HTTP Client**: Automatic exponential backoff retries (429, 500, 502, 503, 504), User-Agent rotation, proxy support, and transparent decompression.
- **Includes `goshazam` CLI**: Full-featured command-line utility with self-upgrade capabilities.

> [!NOTE]
> This library is a Go port of the Python library [ShazamIO](https://github.com/dotX12/ShazamIO) by [dotX12](https://github.com/dotX12).

---

## Installation

### Go Library

```bash
go get github.com/alexgorbatchev/goshazam
```

### `goshazam` CLI

#### Pre-built Binaries (GitHub Releases)

Download pre-built binaries for macOS (Darwin), Linux, or Windows (amd64 / arm64) directly from the [GitHub Releases](https://github.com/alexgorbatchev/goshazam/releases) page:

```bash
# Using GitHub CLI
gh release download -R alexgorbatchev/goshazam --pattern "*darwin_arm64.tar.gz"

# Or install / update via Go toolchain
go install github.com/alexgorbatchev/goshazam/cmd/goshazam@latest
```

Once installed, `goshazam` can also self-upgrade to the latest release in-place:

```bash
goshazam upgrade
```

---

## Library Usage

### 1. Recognizing a Song from Audio

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexgorbatchev/goshazam"
)

func main() {
	client := goshazam.New()
	ctx := context.Background()

	// Recognize from an audio file (MP3, OGG, WAV, FLAC, M4A, etc.)
	result, err := client.Recognize(ctx, "path/to/song.mp3")
	if err != nil {
		log.Fatalf("recognition failed: %v", err)
	}

	if len(result.Matches) == 0 || result.Track == nil {
		fmt.Println("No matches found")
		return
	}

	fmt.Printf("Track:       %s\n", result.Track.Title)
	fmt.Printf("Artist:      %s\n", result.Track.Subtitle)
	fmt.Printf("Shazam Key:  %s\n", result.Track.Key)
	fmt.Printf("Cover Art:   %s\n", result.Track.PhotoURL)
	fmt.Printf("Apple Music: %s\n", result.Track.AppleMusicURL)
	fmt.Printf("Spotify URL: %s\n", result.Track.SpotifyURL)
}
```

### 2. Generating a Shazam Signature

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexgorbatchev/goshazam/pkg/audio"
)

func main() {
	ctx := context.Background()

	// Decode audio file (WAV, MP3, OGG, FLAC, etc.)
	seg, err := audio.DecodeAudioFile(ctx, "track.ogg")
	if err != nil {
		log.Fatal(err)
	}

	// Generate fingerprint
	sg := audio.CreateSignatureGenerator(seg)
	sig := sg.GetNextSignature()
	if sig == nil {
		log.Fatal("audio too short for signature")
	}

	// Output Data URI string: "data:audio/vnd.shazam.sig;base64,..."
	fmt.Println(sig.EncodeToURI())
}
```

### 3. Decoding an Existing Signature URI

```go
package main

import (
	"fmt"
	"log"

	"github.com/alexgorbatchev/goshazam/pkg/signature"
)

func main() {
	uri := "data:audio/vnd.shazam.sig;base64,gCX+ygg8jeyQCgAAA..."
	sig, err := signature.DecodeFromURI(uri)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sample Rate: %d Hz\n", sig.SampleRateHz)
	fmt.Printf("Samples:     %d (%.2f seconds)\n", sig.NumberSamples, sig.DurationSeconds())
	fmt.Printf("Total Peaks: %d\n", sig.TotalPeaks())
}
```

### 4. Fetching Top Charts & Related Tracks

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexgorbatchev/goshazam"
	"github.com/alexgorbatchev/goshazam/pkg/models"
)

func main() {
	client := goshazam.New(goshazam.WithEndpointCountry("US"))
	ctx := context.Background()

	// Top tracks in United States
	topUS, err := client.TopCountryTracks(ctx, "US", 10, 0)
	if err != nil {
		log.Fatal(err)
	}

	for i, song := range topUS.Data {
		fmt.Printf("%d. %s - %s\n", i+1, song.Attributes.Name, song.Attributes.ArtistName)
	}

	// Top Pop tracks worldwide
	topPop, err := client.TopWorldGenreTracks(ctx, models.GenrePop, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	_ = topPop

	// Similar / related songs based on track ID
	related, err := client.RelatedTracks(ctx, 53982678, 5, 0)
	if err != nil {
		log.Fatal(err)
	}
	_ = related
}
```

### 5. Client Configuration Options

```go
client := goshazam.New(
	goshazam.WithLanguage("fr-FR"),                // Accept-Language header
	goshazam.WithEndpointCountry("FR"),             // Apple Music / Shazam catalog region
	goshazam.WithTimeZone("Europe/Paris"),          // Timezone in search payload
	goshazam.WithTimeout(15 * time.Second),         // Request timeout
	goshazam.WithProxy("socks5://127.0.0.1:9050"),  // HTTP/HTTPS/SOCKS5 proxy URL
	goshazam.WithUserAgent("CustomUserAgent/1.0"),  // Custom User-Agent
)
```

---

## `goshazam` CLI

`goshazam` is a command-line tool supporting song recognition, signature extraction, track lookups, and automatic self-updates:

```bash
# Recognize music from a file
goshazam recognize song.mp3

# Output structured JSON
goshazam recognize --json song.ogg

# Generate signature data URI
goshazam signature song.mp3

# Find related tracks
goshazam related 53982678

# Self-upgrade to the latest GitHub release
goshazam upgrade

# Check version
goshazam --version
```

---

## Versioning Policy

`goshazam` version numbers follow SemVer with upstream synchronization:

- **`Major.Minor`** (`0.8.x`): Directly tracks the upstream [ShazamIO](https://github.com/dotX12/ShazamIO) `Major.Minor` release version to ensure algorithm and API parity.
- **`Patch`** (`0.8.0`, `0.8.1`, ...): Represents the `goshazam` Go library release version for bug fixes, performance optimizations, and pure Go tooling enhancements built against that upstream version baseline.

---

## Development

Use `just` to run recipes:

```bash
just build    # Build bin/goshazam
just test     # Run all unit tests with race detector
just coverage # Run tests with coverage summary
just lint     # Run static code analysis and vet
just tidy     # Check module hygiene
```

---

## License

MIT License (compatible with ShazamIO).
