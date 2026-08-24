# goshazam

A fast, idiomatic, asynchronous-friendly Go library and CLI for the reverse-engineered **Shazam API** and audio fingerprinting algorithm.

`goshazam` includes a **pure Go** implementation of Shazam's audio fingerprinting algorithm (Cooley-Tukey Radix-2 FFT, Hanning windowing, frequency/time-domain peak spreading, peak recognition, and binary/URI signature serialization) with **zero CGo dependencies**.

> [!NOTE]
> This library is a Go port of the Python library [ShazamIO](https://github.com/dotX12/ShazamIO) by [dotX12](https://github.com/dotX12).

---

## Features

- **Audio Recognition**: Identify songs from files (`.mp3`, `.ogg`, `.wav`, `.flac`, `.m4a`, etc.), in-memory audio bytes, `io.Reader` streams, or raw PCM.
- **Pure Go DSP & Fingerprinting**:
  - FFT power spectrum analysis (2048-sample window, 128 stride, Hanning window).
  - Peak spreading across frequency and time domains.
  - Parabolic peak interpolation and band grouping.
  - Binary signature encoding and decoding with CRC-32 IEEE verification.
  - Standard Shazam Data URI formatting (`data:audio/vnd.shazam.sig;base64,...`).
- **Shazam Discovery & Catalog APIs**:
  - Top world charts, country charts, city charts, and genre-specific charts.
  - Track metadata, artist profiles, albums, and related track recommendations.
  - Track listening counter statistics.
- **HTTP Client**:
  - Configurable exponential backoff retry policy (429, 500, 502, 503, 504).
  - User-Agent pool rotation.
  - Proxy support (HTTP, HTTPS, SOCKS5).
  - Full `context.Context` cancellation and timeout support.
- **CLI Utility**: Built with Cobra for quick command-line recognition and signature generation.

---

## Installation

```bash
go get github.com/alexgorbatchev/goshazam
```

---

## Quickstart

### 1. Song Recognition

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

	// Recognize from an audio file
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

### 2. Generate a Shazam Signature

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

	// Output Data URI string
	fmt.Println(sig.EncodeToURI())
}
```

### 3. Decode an Existing Signature URI

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

### 4. Fetching Top Charts

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
}
```

---

## Client Configuration Options

```go
client := goshazam.New(
	goshazam.WithLanguage("fr-FR"),                // Accept-Language header
	goshazam.WithEndpointCountry("FR"),             // Apple Music / Shazam catalog region
	goshazam.WithTimeZone("Europe/Paris"),          // Timezone in search payload
	goshazam.WithTimeout(15 * time.Second),         // Request timeout
	goshazam.WithProxy("socks5://127.0.0.1:9050"),  // Proxy URL
	goshazam.WithUserAgent("CustomUserAgent/1.0"),  // Custom User-Agent
)
```

---

## CLI Usage

Build the CLI binary:
```bash
just build
# Binary created at bin/goshazam
```

```bash
# Recognize music from a file
./bin/goshazam recognize song.mp3

# Output JSON format
./bin/goshazam recognize --json song.ogg

# Generate signature data URI
./bin/goshazam signature song.mp3

# Find related tracks
./bin/goshazam related 53982678

# Check version
./bin/goshazam --version
```

---

## Development

```bash
# Run tests
just test

# Run tests with coverage
just coverage

# Lint
just lint

# Tidy dependencies
just tidy
```

---

## License

MIT License
