package ws

import (
	"bytes"
	"compress/flate"
	"io"
	"testing"
)

// TestCompressFramePayload_Basic tests compressFramePayload with normal data
func TestCompressFramePayload_Basic(t *testing.T) {
	input := []byte("Hello, WebSocket compression!")
	compressed, err := compressFramePayload(input)
	if err != nil {
		t.Fatalf("compressFramePayload failed: %v", err)
	}

	// Append deflate tail back for proper decompression
	stream := append(compressed, 0x00, 0x00, 0xFF, 0xFF)
	r := flate.NewReader(bytes.NewReader(stream))
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if !bytes.Equal(decompressed, input) {
		t.Errorf("decompressed data mismatch: got %q, want %q", decompressed, input)
	}
}

// TestCompressFramePayload_Empty tests compressFramePayload with empty input
func TestCompressFramePayload_Empty(t *testing.T) {
	input := []byte{}
	compressed, err := compressFramePayload(input)
	if err != nil {
		t.Fatalf("compressFramePayload failed: %v", err)
	}

	// Append deflate tail back
	stream := append(compressed, 0x00, 0x00, 0xFF, 0xFF)
	r := flate.NewReader(bytes.NewReader(stream))
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if len(decompressed) != 0 {
		t.Errorf("expected empty decompressed data, got %v", decompressed)
	}
}

// TestCompressFramePayload_Large tests compressFramePayload with larger data
func TestCompressFramePayload_Large(t *testing.T) {
	input := bytes.Repeat([]byte("ABCD"), 1024) // 4KB of repetitive data
	compressed, err := compressFramePayload(input)
	if err != nil {
		t.Fatalf("compressFramePayload failed: %v", err)
	}

	// Append deflate tail back
	stream := append(compressed, 0x00, 0x00, 0xFF, 0xFF)
	r := flate.NewReader(bytes.NewReader(stream))
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if !bytes.Equal(decompressed, input) {
		t.Errorf("decompressed data mismatch")
	}
}
