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
			switch bitsPerSample {
			case 16:
				binary.Write(&buf, binary.LittleEndian, val)
			case 8:
				// Unsigned 8-bit
				u8 := byte((int(val) >> 8) + 128)
				buf.WriteByte(u8)
			case 24:
				// 24-bit
				v24 := int32(val) << 8
				buf.WriteByte(byte(v24))
				buf.WriteByte(byte(v24 >> 8))
				buf.WriteByte(byte(v24 >> 16))
			case 32:
				// 32-bit
				v32 := int32(val) << 16
				binary.Write(&buf, binary.LittleEndian, v32)
			}
		}
	}

	return buf.Bytes()
}

func createTestFloat32WAV(sampleRate, channels, numSamples int) []byte {
	var buf bytes.Buffer
	dataSize := numSamples * channels * 4
	chunkSize := 36 + dataSize

	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(3)) // IEEE float
	binary.Write(&buf, binary.LittleEndian, uint16(channels))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*channels*4))
	binary.Write(&buf, binary.LittleEndian, uint16(channels*4))
	binary.Write(&buf, binary.LittleEndian, uint16(32)) // 32-bit float

	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))
	for i := range numSamples {
		f := float32(0.5 * math.Sin(2.0*math.Pi*440.0*float64(i)/float64(sampleRate)))
		for range channels {
			binary.Write(&buf, binary.LittleEndian, f)
		}
	}

	return buf.Bytes()
}

func TestDecodeWAVBitDepths(t *testing.T) {
	for _, bitDepth := range []int{8, 16, 24, 32} {
		wavData := createTestWAV(16000, 1, bitDepth, 1600)
		seg, err := DecodeWAVBytes(wavData)
		if err != nil {
			t.Fatalf("failed decoding %d-bit WAV: %v", bitDepth, err)
		}
		if len(seg.Samples) != 1600 {
			t.Errorf("%d-bit: expected 1600 samples, got %d", bitDepth, len(seg.Samples))
		}
	}

	// 32-bit float WAV
	floatWAV := createTestFloat32WAV(16000, 1, 1600)
	segFloat, err := DecodeWAVBytes(floatWAV)
	if err != nil {
		t.Fatalf("failed decoding 32-bit float WAV: %v", err)
	}
	if len(segFloat.Samples) != 1600 {
		t.Errorf("expected 1600 float samples, got %d", len(segFloat.Samples))
	}
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
	dataSize := 100 * 2
	chunkSize := 36 + (8 + 15 + 1) + dataSize

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

func TestAudioHelpers(t *testing.T) {
	ctx := context.Background()

	// NewSegment zero values
	segZero := NewSegment([]int16{1, 2}, 0, 0)
	if segZero.SampleRate != 16000 || segZero.Channels != 1 {
		t.Errorf("expected fallback defaults in NewSegment")
	}

	// DurationSeconds zero values
	segZero.SampleRate = 0
	if segZero.DurationSeconds() != 0 {
		t.Errorf("expected 0 duration")
	}

	// FromPCM16LE odd byte length
	if _, err := FromPCM16LE([]byte{1, 2, 3}, 16000, 1); err == nil {
		t.Errorf("expected error for odd byte length in FromPCM16LE")
	}

	// ReadAllPCM16
	pcmData := []byte{0x01, 0x00, 0x02, 0x00}
	samples, err := ReadAllPCM16(bytes.NewReader(pcmData))
	if err != nil {
		t.Fatalf("ReadAllPCM16 failed: %v", err)
	}
	if len(samples) != 2 || samples[0] != 1 || samples[1] != 2 {
		t.Errorf("ReadAllPCM16 mismatch: %v", samples)
	}

	// DecodeAudioBytes & DecodeAudioReader on WAV
	wavData := createTestWAV(16000, 1, 16, 500)
	seg1, err := DecodeAudioBytes(ctx, wavData)
	if err != nil {
		t.Fatalf("DecodeAudioBytes failed on WAV: %v", err)
	}
	if len(seg1.Samples) != 500 {
		t.Errorf("expected 500 samples, got %d", len(seg1.Samples))
	}

	seg2, err := DecodeAudioReader(ctx, bytes.NewReader(wavData))
	if err != nil {
		t.Fatalf("DecodeAudioReader failed on WAV: %v", err)
	}
	if len(seg2.Samples) != 500 {
		t.Errorf("expected 500 samples, got %d", len(seg2.Samples))
	}

	// Error paths
	if _, err := DecodeAudioFile(ctx, "non_existent_file_123.mp3"); err == nil {
		t.Errorf("expected error for non-existent audio file")
	}

	// Invalid WAV data in DecodeWAV
	if _, err := DecodeWAV(bytes.NewReader([]byte("not a wav file header"))); err == nil {
		t.Errorf("expected error for invalid WAV data")
	}
	if _, err := DecodeWAV(bytes.NewReader([]byte("RIFF1234WAVXfmt 123456"))); err == nil {
		t.Errorf("expected error for missing fmt/data chunks")
	}

	// Non-WAV bytes with ffmpeg if available
	if HasFFmpeg() {
		oggPath := filepath.Join("..", "..", "ShazamIO", "examples", "data", "Gloria.ogg")
		if data, err := os.ReadFile(oggPath); err == nil {
			segOgg, err := DecodeAudioBytes(ctx, data)
			if err != nil || len(segOgg.Samples) == 0 {
				t.Errorf("DecodeAudioBytes on OGG bytes failed: %v", err)
			}
			segReader, err := DecodeAudioReader(ctx, bytes.NewReader(data))
			if err != nil || len(segReader.Samples) == 0 {
				t.Errorf("DecodeAudioReader on OGG stream failed: %v", err)
			}
		}
	}
	if HasFFmpeg() {
		oggPath := filepath.Join("..", "..", "ShazamIO", "examples", "data", "Gloria.ogg")
		if data, err := os.ReadFile(oggPath); err == nil {
			segOgg, err := DecodeAudioBytes(ctx, data)
			if err != nil || len(segOgg.Samples) == 0 {
				t.Errorf("DecodeAudioBytes on OGG bytes failed: %v", err)
			}
			segReader, err := DecodeAudioReader(ctx, bytes.NewReader(data))
			if err != nil || len(segReader.Samples) == 0 {
				t.Errorf("DecodeAudioReader on OGG stream failed: %v", err)
			}
		}
	}
}

func TestDecodeWAVUnsupported(t *testing.T) {
	// Format tag 2 (ADPCM)
	var buf bytes.Buffer
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(40))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(2)) // unsupported
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint32(16000))
	binary.Write(&buf, binary.LittleEndian, uint32(32000))
	binary.Write(&buf, binary.LittleEndian, uint16(2))
	binary.Write(&buf, binary.LittleEndian, uint16(16))
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(4))
	buf.Write([]byte{0, 0, 0, 0})

	if _, err := DecodeWAVBytes(buf.Bytes()); err == nil {
		t.Errorf("expected error on unsupported format tag")
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
