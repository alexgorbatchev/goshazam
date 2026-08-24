package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

var (
	// ErrInvalidWAV indicates corrupted or unsupported WAV container format.
	ErrInvalidWAV = errors.New("invalid WAV file")
)

// DecodeWAV decodes a standard WAV / RIFF PCM file into an Audio Segment in pure Go.
func DecodeWAV(r io.Reader) (*Segment, error) {
	var header [12]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, fmt.Errorf("%w: failed reading RIFF header: %w", ErrInvalidWAV, err)
	}

	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, fmt.Errorf("%w: missing RIFF/WAVE header", ErrInvalidWAV)
	}

	var (
		audioFormat   uint16
		numChannels   uint16
		sampleRate    uint32
		bitsPerSample uint16
		dataBytes     []byte
	)

	for {
		var chunkHeader [8]byte
		if _, err := io.ReadFull(r, chunkHeader[:]); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return nil, fmt.Errorf("%w: reading chunk header: %w", ErrInvalidWAV, err)
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		switch chunkID {
		case "fmt ":
			if chunkSize < 16 {
				return nil, fmt.Errorf("%w: fmt chunk too small", ErrInvalidWAV)
			}
			fmtData := make([]byte, chunkSize)
			if _, err := io.ReadFull(r, fmtData); err != nil {
				return nil, fmt.Errorf("%w: reading fmt chunk: %w", ErrInvalidWAV, err)
			}

			audioFormat = binary.LittleEndian.Uint16(fmtData[0:2])
			numChannels = binary.LittleEndian.Uint16(fmtData[2:4])
			sampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			bitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])

		case "data":
			dataBytes = make([]byte, chunkSize)
			if _, err := io.ReadFull(r, dataBytes); err != nil {
				return nil, fmt.Errorf("%w: reading data chunk: %w", ErrInvalidWAV, err)
			}
			// Keep scanning or stop after data chunk
		default:
			// Skip unrecognized chunks
			if _, err := io.CopyN(io.Discard, r, int64(chunkSize)); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, fmt.Errorf("%w: skipping chunk %q: %w", ErrInvalidWAV, chunkID, err)
			}
		}

		// RIFF chunks are 2-byte aligned; discard padding byte if chunkSize is odd
		if chunkSize%2 != 0 {
			if _, err := io.CopyN(io.Discard, r, 1); err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("%w: discarding padding byte: %w", ErrInvalidWAV, err)
			}
		}
	}

	if dataBytes == nil {
		return nil, fmt.Errorf("%w: no data chunk found", ErrInvalidWAV)
	}

	if numChannels == 0 || sampleRate == 0 {
		return nil, fmt.Errorf("%w: invalid format parameters (channels=%d, sampleRate=%d)", ErrInvalidWAV, numChannels, sampleRate)
	}

	var samples []int16

	switch audioFormat {
	case 1: // Standard Integer PCM
		switch bitsPerSample {
		case 16:
			numSamples := len(dataBytes) / 2
			samples = make([]int16, numSamples)
			for i := range numSamples {
				samples[i] = int16(binary.LittleEndian.Uint16(dataBytes[i*2 : i*2+2]))
			}
		case 8:
			// 8-bit PCM is unsigned: 0..255, midpoint 128
			samples = make([]int16, len(dataBytes))
			for i, b := range dataBytes {
				samples[i] = int16((int(b) - 128) << 8)
			}
		case 24:
			numSamples := len(dataBytes) / 3
			samples = make([]int16, numSamples)
			for i := range numSamples {
				b0 := dataBytes[i*3]
				b1 := dataBytes[i*3+1]
				b2 := dataBytes[i*3+2]
				// Sign extend 24-bit
				val := int32(b0) | (int32(b1) << 8) | (int32(int8(b2)) << 16)
				samples[i] = int16(val >> 8)
			}
		case 32:
			numSamples := len(dataBytes) / 4
			samples = make([]int16, numSamples)
			for i := range numSamples {
				val := int32(binary.LittleEndian.Uint32(dataBytes[i*4 : i*4+4]))
				samples[i] = int16(val >> 16)
			}
		default:
			return nil, fmt.Errorf("%w: unsupported PCM bit depth %d", ErrInvalidWAV, bitsPerSample)
		}

	case 3: // IEEE 754 Float PCM (32-bit)
		if bitsPerSample != 32 {
			return nil, fmt.Errorf("%w: unsupported float bit depth %d", ErrInvalidWAV, bitsPerSample)
		}
		numSamples := len(dataBytes) / 4
		samples = make([]int16, numSamples)
		for i := range numSamples {
			bits := binary.LittleEndian.Uint32(dataBytes[i*4 : i*4+4])
			f := math.Float32frombits(bits)
			if f > 1.0 {
				f = 1.0
			} else if f < -1.0 {
				f = -1.0
			}
			samples[i] = int16(f * 32767.0)
		}

	default:
		return nil, fmt.Errorf("%w: unsupported audio format tag %d", ErrInvalidWAV, audioFormat)
	}

	return NewSegment(samples, int(sampleRate), int(numChannels)), nil
}

// DecodeWAVBytes decodes WAV data from in-memory byte slice.
func DecodeWAVBytes(data []byte) (*Segment, error) {
	return DecodeWAV(bytes.NewReader(data))
}
