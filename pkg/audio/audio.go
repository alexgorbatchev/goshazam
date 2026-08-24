package audio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/alexgorbatchev/goshazam/pkg/signature"
)

var (
	// ErrUnsupportedAudioFormat indicates unrecognized or invalid audio format.
	ErrUnsupportedAudioFormat = errors.New("unsupported audio format")
	// ErrInvalidPCMData indicates corrupted or improperly sized PCM data.
	ErrInvalidPCMData = errors.New("invalid PCM data")
)

// Segment represents in-memory decoded audio samples.
type Segment struct {
	Samples    []int16
	SampleRate int
	Channels   int
}

// NewSegment creates a new Segment.
func NewSegment(samples []int16, sampleRate, channels int) *Segment {
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	if channels <= 0 {
		channels = 1
	}
	return &Segment{
		Samples:    samples,
		SampleRate: sampleRate,
		Channels:   channels,
	}
}

// FromPCM16LE creates a Segment from raw signed 16-bit little-endian PCM bytes.
func FromPCM16LE(data []byte, sampleRate, channels int) (*Segment, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("%w: byte length %d is not a multiple of 2", ErrInvalidPCMData, len(data))
	}

	numSamples := len(data) / 2
	samples := make([]int16, numSamples)
	for i := range numSamples {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}

	return NewSegment(samples, sampleRate, channels), nil
}

// DurationSeconds returns the audio length in seconds.
func (s *Segment) DurationSeconds() float64 {
	if s.SampleRate <= 0 || s.Channels <= 0 {
		return 0
	}
	return float64(len(s.Samples)) / float64(s.SampleRate*s.Channels)
}

// ToMono converts multi-channel audio to mono by averaging channels.
func (s *Segment) ToMono() *Segment {
	if s.Channels == 1 {
		return s
	}

	numFrames := len(s.Samples) / s.Channels
	mono := make([]int16, numFrames)

	for i := range numFrames {
		var sum int64
		for ch := range s.Channels {
			sum += int64(s.Samples[i*s.Channels+ch])
		}
		mono[i] = int16(sum / int64(s.Channels))
	}

	return NewSegment(mono, s.SampleRate, 1)
}

// Resample linear-interpolates samples to targetSampleRate.
func (s *Segment) Resample(targetSampleRate int) *Segment {
	if s.SampleRate == targetSampleRate || targetSampleRate <= 0 {
		return s
	}

	ratio := float64(s.SampleRate) / float64(targetSampleRate)
	numOutSamples := int(math.Floor(float64(len(s.Samples)) / ratio))
	out := make([]int16, numOutSamples)

	for i := range numOutSamples {
		srcPos := float64(i) * ratio
		idx0 := int(srcPos)
		idx1 := idx0 + 1
		if idx1 >= len(s.Samples) {
			idx1 = len(s.Samples) - 1
		}
		frac := srcPos - float64(idx0)

		v0 := float64(s.Samples[idx0])
		v1 := float64(s.Samples[idx1])
		out[i] = int16(v0 + frac*(v1-v0))
	}

	return NewSegment(out, targetSampleRate, s.Channels)
}

// Normalize prepares audio for Shazam fingerprinting (16000 Hz, mono, 16-bit).
func (s *Segment) Normalize() *Segment {
	seg := s
	if seg.Channels > 1 {
		seg = seg.ToMono()
	}
	if seg.SampleRate != 16000 {
		seg = seg.Resample(16000)
	}
	return seg
}

// CreateSignatureGenerator creates and configures a SignatureGenerator for this audio segment.
// If the audio is longer than 36s (12 * 3), it centers around the middle of the recording.
func CreateSignatureGenerator(audio *Segment) *signature.SignatureGenerator {
	normalized := audio.Normalize()
	sg := signature.NewSignatureGenerator()
	sg.FeedInput(normalized.Samples)
	sg.MaxTimeSeconds = 12.0

	dur := normalized.DurationSeconds()
	if dur > 12.0*3.0 {
		skipSec := int(dur/2.0) - 6
		if skipSec > 0 {
			sg.SamplesProcessed += 16000 * skipSec
		}
	}

	return sg
}

// PCMReader wraps an io.Reader returning raw 16-bit 16kHz mono PCM.
func ReadAllPCM16(r io.Reader) ([]int16, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	seg, err := FromPCM16LE(data, 16000, 1)
	if err != nil {
		return nil, err
	}
	return seg.Samples, nil
}
