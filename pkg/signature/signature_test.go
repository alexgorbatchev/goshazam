package signature

import (
	"encoding/binary"
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
