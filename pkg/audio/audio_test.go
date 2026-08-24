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

	// Direct Resample on stereo segment
	resampled := seg.Resample(16000)
	if resampled.SampleRate != 16000 || resampled.Channels != 1 {
		t.Errorf("expected resampled stereo segment to convert to mono 16000Hz, got rate=%d ch=%d", resampled.SampleRate, resampled.Channels)
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

func TestDecodeWAVOddMetadataChunk(t *testing.T) {
	var buf bytes.Buffer
	// RIFF header with odd-sized metadata chunk before data
	dataSize := 100 * 2                       // 100 16-bit samples = 200 bytes
	chunkSize := 36 + (8 + 15 + 1) + dataSize // with 1 byte pad for 15 byte chunk

	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // mono
	binary.Write(&buf, binary.LittleEndian, uint32(16000))
	binary.Write(&buf, binary.LittleEndian, uint32(32000))
	binary.Write(&buf, binary.LittleEndian, uint16(2))
	binary.Write(&buf, binary.LittleEndian, uint16(16))

	// odd-sized metadata chunk
	buf.WriteString("INFO")
	binary.Write(&buf, binary.LittleEndian, uint32(15)) // odd size
	buf.Write([]byte("odd_metadata_12"))                // 15 bytes
	buf.WriteByte(0)                                    // 1 byte pad

	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))
	for range 100 {
		binary.Write(&buf, binary.LittleEndian, int16(1234))
	}

	seg, err := DecodeWAVBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("DecodeWAV with odd metadata chunk failed: %v", err)
	}

	if len(seg.Samples) != 100 {
		t.Errorf("expected 100 samples, got %d", len(seg.Samples))
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
