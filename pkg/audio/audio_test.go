package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func createTestWAV(sampleRate, channels, bitsPerSample int, numSamples int) []byte {
	var buf bytes.Buffer

	byteRate := sampleRate * channels * (bitsPerSample / 8)
	blockAlign := channels * (bitsPerSample / 8)
	dataSize := numSamples * channels * (bitsPerSample / 8)
	chunkSize := 36 + dataSize

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16)) // subchunk1 size
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(&buf, binary.LittleEndian, uint16(channels))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(&buf, binary.LittleEndian, uint16(blockAlign))
	binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample))

	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))

	for i := range numSamples {
		val := int16(10000.0 * math.Sin(2.0*math.Pi*440.0*float64(i)/float64(sampleRate)))
		for range channels {
			binary.Write(&buf, binary.LittleEndian, val)
		}
	}

	return buf.Bytes()
}

func TestDecodeWAV16Bit(t *testing.T) {
	wavData := createTestWAV(44100, 2, 16, 44100) // 1 second of stereo 44.1kHz
	seg, err := DecodeWAVBytes(wavData)
	if err != nil {
		t.Fatalf("DecodeWAVBytes failed: %v", err)
	}

	if seg.SampleRate != 44100 {
		t.Errorf("expected sample rate 44100, got %d", seg.SampleRate)
	}
	if seg.Channels != 2 {
		t.Errorf("expected channels 2, got %d", seg.Channels)
	}

	norm := seg.Normalize()
	if norm.SampleRate != 16000 {
		t.Errorf("expected normalized sample rate 16000, got %d", norm.SampleRate)
	}
	if norm.Channels != 1 {
		t.Errorf("expected normalized channels 1, got %d", norm.Channels)
	}
	if math.Abs(norm.DurationSeconds()-1.0) > 0.01 {
		t.Errorf("expected duration ~1.0s, got %f", norm.DurationSeconds())
	}
}

func TestDecodeAudioWithFFmpeg(t *testing.T) {
	if !HasFFmpeg() {
		t.Skip("ffmpeg not installed, skipping ffmpeg tests")
	}

	oggPath := filepath.Join("..", "..", "ShazamIO", "examples", "data", "Gloria.ogg")
	if _, err := os.Stat(oggPath); os.IsNotExist(err) {
		t.Skipf("test file %s not found", oggPath)
	}

	ctx := context.Background()
	seg, err := DecodeAudioFile(ctx, oggPath)
	if err != nil {
		t.Fatalf("DecodeAudioFile failed: %v", err)
	}

	if seg.SampleRate != 16000 {
		t.Errorf("expected 16000 Hz, got %d", seg.SampleRate)
	}
	if seg.Channels != 1 {
		t.Errorf("expected 1 channel, got %d", seg.Channels)
	}
	if len(seg.Samples) == 0 {
		t.Errorf("expected non-empty samples")
	}

	sg := CreateSignatureGenerator(seg)
	sig := sg.GetNextSignature()
	if sig == nil {
		t.Fatalf("expected non-nil signature from Gloria.ogg")
	}

	if sig.TotalPeaks() == 0 {
		t.Errorf("expected signature peaks, got 0")
	}
}
