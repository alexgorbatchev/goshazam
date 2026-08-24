package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	// ErrFFmpegNotFound is returned when ffmpeg is required to decode an audio format but is not in PATH.
	ErrFFmpegNotFound = errors.New("ffmpeg executable not found in PATH; install ffmpeg to decode MP3, OGG, FLAC, AAC, and other formats")
)

// FFmpegPath returns the configured path to the ffmpeg executable, checking PATH.
func FFmpegPath() (string, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", ErrFFmpegNotFound
	}
	return path, nil
}

// HasFFmpeg returns true if ffmpeg is available on the system.
func HasFFmpeg() bool {
	_, err := FFmpegPath()
	return err == nil
}

// DecodeAudioFile decodes any audio file (WAV, MP3, OGG, FLAC, M4A, etc.) into a 16 kHz mono Segment.
func DecodeAudioFile(ctx context.Context, filePath string) (*Segment, error) {
	// Try reading file header first
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening audio file: %w", err)
	}
	defer f.Close()

	var head [12]byte
	n, _ := io.ReadFull(f, head[:])
	if n >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WAVE" {
		if _, err := f.Seek(0, io.SeekStart); err == nil {
			if seg, err := DecodeWAV(f); err == nil {
				return seg.Normalize(), nil
			}
		}
	}

	// Use ffmpeg to convert to 16kHz s16le mono PCM
	ffmpeg, err := FFmpegPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, ffmpeg,
		"-v", "error",
		"-i", filePath,
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", "1",
		"-ar", "16000",
		"-",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %w (stderr: %s)", err, stderr.String())
	}

	return FromPCM16LE(stdout.Bytes(), 16000, 1)
}

// DecodeAudioBytes decodes in-memory audio bytes (WAV, MP3, OGG, FLAC, M4A, etc.) into a 16 kHz mono Segment.
func DecodeAudioBytes(ctx context.Context, data []byte) (*Segment, error) {
	// Check for WAV format
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		if seg, err := DecodeWAVBytes(data); err == nil {
			return seg.Normalize(), nil
		}
	}

	ffmpeg, err := FFmpegPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, ffmpeg,
		"-v", "error",
		"-i", "pipe:0",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", "1",
		"-ar", "16000",
		"-",
	)

	cmd.Stdin = bytes.NewReader(data)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg decoding failed: %w (stderr: %s)", err, stderr.String())
	}

	return FromPCM16LE(stdout.Bytes(), 16000, 1)
}

// DecodeAudioReader decodes audio from an io.Reader stream into a 16 kHz mono Segment.
func DecodeAudioReader(ctx context.Context, r io.Reader) (*Segment, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading audio stream: %w", err)
	}
	return DecodeAudioBytes(ctx, data)
}
