package signature

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestSignatureRoundtrip(t *testing.T) {
	msg := NewDecodedMessage()
	msg.SampleRateHz = 16000
	msg.NumberSamples = 192000

	msg.FrequencyBandToSoundPeaks[FrequencyBand250_520] = []FrequencyPeak{
		{FFTPassNumber: 10, PeakMagnitude: 7000, CorrectedPeakFrequencyBin: 300, SampleRateHz: 16000},
		{FFTPassNumber: 280, PeakMagnitude: 8500, CorrectedPeakFrequencyBin: 450, SampleRateHz: 16000}, // delta > 255
	}
	msg.FrequencyBandToSoundPeaks[FrequencyBand520_1450] = []FrequencyPeak{
		{FFTPassNumber: 5, PeakMagnitude: 6500, CorrectedPeakFrequencyBin: 600, SampleRateHz: 16000},
		{FFTPassNumber: 50, PeakMagnitude: 9000, CorrectedPeakFrequencyBin: 1000, SampleRateHz: 16000},
	}

	bin := msg.EncodeToBinary()
	if len(bin) <= HeaderSize {
		t.Fatalf("encoded binary too small: %d", len(bin))
	}

	decoded, err := DecodeFromBinary(bin)
	if err != nil {
		t.Fatalf("DecodeFromBinary failed: %v", err)
	}

	if decoded.SampleRateHz != msg.SampleRateHz {
		t.Errorf("expected sample rate %d, got %d", msg.SampleRateHz, decoded.SampleRateHz)
	}
	if decoded.NumberSamples != msg.NumberSamples {
		t.Errorf("expected number samples %d, got %d", msg.NumberSamples, decoded.NumberSamples)
	}

	if len(decoded.FrequencyBandToSoundPeaks[FrequencyBand250_520]) != 2 {
		t.Fatalf("expected 2 peaks in band 0, got %d", len(decoded.FrequencyBandToSoundPeaks[FrequencyBand250_520]))
	}

	p0 := decoded.FrequencyBandToSoundPeaks[FrequencyBand250_520][0]
	if p0.FFTPassNumber != 10 || p0.PeakMagnitude != 7000 || p0.CorrectedPeakFrequencyBin != 300 {
		t.Errorf("peak 0 mismatch: %+v", p0)
	}

	p1 := decoded.FrequencyBandToSoundPeaks[FrequencyBand250_520][1]
	if p1.FFTPassNumber != 280 || p1.PeakMagnitude != 8500 || p1.CorrectedPeakFrequencyBin != 450 {
		t.Errorf("peak 1 (with delta > 255) mismatch: %+v", p1)
	}

	uri := msg.EncodeToURI()
	decodedFromURI, err := DecodeFromURI(uri)
	if err != nil {
		t.Fatalf("DecodeFromURI failed: %v", err)
	}

	if decodedFromURI.TotalPeaks() != msg.TotalPeaks() {
		t.Errorf("expected %d total peaks, got %d", msg.TotalPeaks(), decodedFromURI.TotalPeaks())
	}
}

func TestSignatureExactWireFormat(t *testing.T) {
	msg := NewDecodedMessage()
	msg.SampleRateHz = 16000
	msg.NumberSamples = 16000
	bin := msg.EncodeToBinary()

	magic1 := binary.LittleEndian.Uint32(bin[0:4])
	if magic1 != 0xCAFE2580 {
		t.Fatalf("expected magic1 0xCAFE2580, got 0x%08X", magic1)
	}

	magic2 := binary.LittleEndian.Uint32(bin[12:16])
	if magic2 != 0x94119C00 {
		t.Fatalf("expected magic2 0x94119C00, got 0x%08X", magic2)
	}

	fixedVal := binary.LittleEndian.Uint32(bin[44:48])
	if fixedVal != 0x007C0000 {
		t.Fatalf("expected fixed value 0x007C0000, got 0x%08X", fixedVal)
	}
}

func TestSignatureCorruptedCRC(t *testing.T) {
	msg := NewDecodedMessage()
	bin := msg.EncodeToBinary()
	// Corrupt one byte in payload
	bin[len(bin)-1] ^= 0xFF

	_, err := DecodeFromBinary(bin)
	if err == nil {
		t.Fatalf("expected CRC error on corrupted data, got nil")
	}
}

func TestFrequencyPeakMethods(t *testing.T) {
	p := FrequencyPeak{
		FFTPassNumber:             100,
		PeakMagnitude:             8000,
		CorrectedPeakFrequencyBin: 640,
		SampleRateHz:              16000,
	}

	freq := p.FrequencyHz()
	// freq = 640 * (16000 / 2 / 1024 / 64) = 640 * 0.1220703125 = 78.125 Hz
	if math.Abs(freq-78.125) > 0.001 {
		t.Errorf("expected frequency 78.125, got %f", freq)
	}

	amp := p.AmplitudePCM()
	if amp <= 0 {
		t.Errorf("expected positive amplitude, got %f", amp)
	}

	secs := p.Seconds()
	// secs = 100 * 128 / 16000 = 0.8s
	if math.Abs(secs-0.8) > 0.001 {
		t.Errorf("expected seconds 0.8, got %f", secs)
	}

	// Test zero sample rate defaults
	pZero := FrequencyPeak{FFTPassNumber: 100, CorrectedPeakFrequencyBin: 640}
	if pZero.FrequencyHz() <= 0 || pZero.Seconds() <= 0 {
		t.Errorf("expected valid non-zero results with zero sample rate fallback")
	}
}

func TestSampleRateEnums(t *testing.T) {
	rates := []int{8000, 11025, 16000, 32000, 44100, 48000}
	for _, r := range rates {
		enum, ok := SampleRateFromHz(r)
		if !ok {
			t.Errorf("expected ok for %d Hz", r)
		}
		if enum.Hz() != r {
			t.Errorf("expected %d Hz from enum, got %d", r, enum.Hz())
		}
	}

	// Invalid rate fallback
	fallback, ok := SampleRateFromHz(99999)
	if ok || fallback != SampleRate16000 {
		t.Errorf("expected fallback to 16000 for invalid rate")
	}

	invalidEnum := SampleRate(999)
	if invalidEnum.Hz() != 16000 {
		t.Errorf("expected invalid enum Hz() to fallback to 16000")
	}
}

func TestDecodedMessageDuration(t *testing.T) {
	msg := NewDecodedMessage()
	msg.SampleRateHz = 16000
	msg.NumberSamples = 32000
	if msg.DurationSeconds() != 2.0 {
		t.Errorf("expected duration 2.0s, got %f", msg.DurationSeconds())
	}

	msgZero := &DecodedMessage{}
	if msgZero.DurationSeconds() != 0 {
		t.Errorf("expected 0 duration for zero sample rate")
	}
}

func TestDecodeErrors(t *testing.T) {
	// Too short
	if _, err := DecodeFromBinary([]byte{1, 2, 3}); err == nil {
		t.Errorf("expected error for short data")
	}

	// Invalid base64 URI
	if _, err := DecodeFromURI("data:audio/vnd.shazam.sig;base64,!!!invalid!!!"); err == nil {
		t.Errorf("expected error for invalid base64 URI")
	}
}
