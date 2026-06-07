package tests

import (
	"bytes"
	"slices"
	"testing"

	anchore "github.com/anchore/go-lzo"
	"github.com/commander-spaceman/me2lzo/decompress"
	"github.com/woozymasta/lzo"
)

func TestDecompressZeroData(t *testing.T) {
	original := make([]byte, 4096)
	assertRoundTrip(t, original)
}

func TestDecompressIncremental(t *testing.T) {
	original := make([]byte, 1024)
	for i := range original {
		original[i] = byte(i)
	}
	assertRoundTrip(t, original)
}

func TestDecompressRepeatedPattern(t *testing.T) {
	pattern := []byte("ME2LZO ")
	original := make([]byte, 0, 2048)
	for len(original) < 2048 {
		original = append(original, pattern...)
	}
	assertRoundTrip(t, original)
}

func TestDecompressText(t *testing.T) {
	original := []byte("The quick brown fox jumps over the lazy dog. " +
		"This is a test of the LZO1X decompression algorithm for ME2 OT PCC files. " +
		"Mass Effect 2 uses LZO compression for package chunks.")
	assertRoundTrip(t, original)
}

func TestDecompressSingleByte(t *testing.T) {
	original := []byte{0x42}
	assertRoundTrip(t, original)
}

func TestDecompressEmpty(t *testing.T) {
	compressed := compressLZO1X(t, []byte{})
	n, err := decompress.Decompress(compressed, make([]byte, 1024))
	if err != nil {
		t.Fatalf("decompress failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 bytes, got %d", n)
	}
}

func TestDecompressCrossLibrary(t *testing.T) {
	patterns := [][]byte{
		make([]byte, 4096),
		[]byte("aaaaaabbbbbbcccccc"),
		[]byte("Mass Effect 2 Original Trilogy dialogue extraction"),
	}

	for i, original := range patterns {
		compressed := compressLZO1X(t, original)

		refOut := make([]byte, len(original)*2)
		refN, refErr := anchore.Decompress(compressed, refOut)
		if refErr != nil {
			t.Fatalf("pattern %d: anchore/go-lzo decompress failed: %v", i, refErr)
		}

		ourOut := make([]byte, len(original)*2)
		ourN, ourErr := decompress.Decompress(compressed, ourOut)
		if ourErr != nil {
			t.Fatalf("pattern %d: me2lzo decompress failed: %v", i, ourErr)
		}

		if refN != ourN {
			t.Errorf("pattern %d: output size mismatch: anchore=%d me2lzo=%d", i, refN, ourN)
		}
		if !bytes.Equal(refOut[:refN], ourOut[:ourN]) {
			t.Errorf("pattern %d: output mismatch", i)
		}
	}
}

func TestDecompressOutputSizeMismatch(t *testing.T) {
	original := bytes.Repeat([]byte{0xAA}, 1024)
	compressed := compressLZO1X(t, original)

	smallDst := make([]byte, 100)
	_, err := decompress.Decompress(compressed, smallDst)
	if err == nil {
		t.Error("expected output overrun error")
	}
}

func TestDecompressTruncatedInput(t *testing.T) {
	original := bytes.Repeat([]byte{0xBB}, 2048)
	compressed := compressLZO1X(t, original)

	truncated := compressed[:len(compressed)-10]
	dst := make([]byte, len(original)*2)
	_, err := decompress.Decompress(truncated, dst)

	switch {
	case err == nil:
		t.Error("expected error for truncated input")
	case err == decompress.ErrInputOverrun:
	case err == decompress.ErrInputNotConsumed:
	default:
		t.Logf("got error: %v", err)
	}
}

func assertRoundTrip(t *testing.T, original []byte) {
	t.Helper()
	compressed := compressLZO1X(t, original)
	decompressed := make([]byte, len(original)*2)
	n, err := decompress.Decompress(compressed, decompressed)
	if err != nil {
		t.Fatalf("decompress failed: %v", err)
	}
	if n != len(original) {
		t.Fatalf("size mismatch: got %d want %d", n, len(original))
	}
	if !slices.Equal(decompressed[:n], original) {
		t.Fatalf("data mismatch")
	}
}

func compressLZO1X(t *testing.T, data []byte) []byte {
	t.Helper()
	compressed, err := lzo.Compress(data, lzo.DefaultCompressOptions())
	if err != nil {
		t.Fatalf("compress failed: %v", err)
	}
	return compressed
}
