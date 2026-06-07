package tests

import (
	"bytes"
	"slices"
	"testing"

	"github.com/commander-spaceman/me2lzo/compress"
	"github.com/commander-spaceman/me2lzo/decompress"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"zeros_4k", make([]byte, 4096)},
		{"zeros_64k", make([]byte, 65536)},
		{"incremental_1k", incrementalData(1024)},
		{"incremental_8k", incrementalData(8192)},
		{"repeated_pattern", bytes.Repeat([]byte("ME2LZO "), 512)},
		{"text_short", []byte("Mass Effect 2 dialogue extraction")},
		{"text_long", bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 100)},
		{"single_byte", []byte{0x42}},
		{"two_bytes", []byte{0x42, 0x43}},
		{"all_same", bytes.Repeat([]byte{0xAA}, 2048)},
		{"mixed", mixedData(5000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := compress.Compress(tt.data)
			if err != nil {
				t.Fatalf("compress failed: %v", err)
			}

			dst := make([]byte, len(tt.data)*2)
			n, err := decompress.Decompress(compressed, dst)
			if err != nil {
				t.Fatalf("decompress failed: %v", err)
			}

			if n != len(tt.data) {
				t.Fatalf("size mismatch: got %d want %d", n, len(tt.data))
			}
			if !slices.Equal(dst[:n], tt.data) {
				t.Fatalf("data mismatch (len=%d)", len(tt.data))
			}
		})
	}
}

func TestCompressProducesValidLZO(t *testing.T) {
	original := bytes.Repeat([]byte("LZO1X compression test pattern "), 128)
	original = append(original, incrementalData(2048)...)

	compressed, err := compress.Compress(original)
	if err != nil {
		t.Fatalf("compress failed: %v", err)
	}

	if len(compressed) == 0 {
		t.Fatal("compressed output is empty")
	}

	ratio := float64(len(compressed)) / float64(len(original))
	if ratio >= 1.0 && len(original) > 100 {
		t.Errorf("compression ratio %.2f >= 1.0 for non-trivial input", ratio)
	}

	t.Logf("compressed %d -> %d bytes (%.1f%%)", len(original), len(compressed), ratio*100)
}

func TestCompressEmpty(t *testing.T) {
	compressed, err := compress.Compress(nil)
	if err != nil {
		t.Fatalf("compress failed: %v", err)
	}
	n, err := decompress.Decompress(compressed, make([]byte, 1024))
	if err != nil {
		t.Fatalf("decompress failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 bytes, got %d", n)
	}
}

func TestCompressDeterministic(t *testing.T) {
	data := bytes.Repeat([]byte{0x42}, 4096)

	first, err := compress.Compress(data)
	if err != nil {
		t.Fatalf("first compress failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		again, err := compress.Compress(data)
		if err != nil {
			t.Fatalf("compress %d failed: %v", i, err)
		}
		if !bytes.Equal(first, again) {
			t.Fatalf("non-deterministic output at iteration %d", i)
		}
	}
}

func incrementalData(n int) []byte {
	d := make([]byte, n)
	for i := range d {
		d[i] = byte(i)
	}
	return d
}

func mixedData(n int) []byte {
	d := make([]byte, n)
	for i := range d {
		switch i % 7 {
		case 0:
			d[i] = byte(i)
		case 1:
			d[i] = 0xFF
		case 2:
			d[i] = 0x00
		case 3, 4:
			d[i] = byte(i >> 2)
		default:
			d[i] = 0xAA
		}
	}
	return d
}
