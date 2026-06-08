# me2lzo

Pure Go implementation of LZO1X compression and decompression for Mass Effect 2
Original Trilogy `.pcc` package files.

## Overview

`me2lzo` provides a zero-dependency, MIT-licensed LZO1X library designed for
ME2 OT's specific LZO block format. It handles both decompression of
LZO-compressed PCC chunks and LZO1X-1 compression for writing compressed
packages.

The decompressor is byte-identical to the Oberhumer reference implementation
on all tested ME2 OT blocks.

## Installation

```
go get github.com/commander-spaceman/me2lzo
```

## Usage

```go
import (
    "github.com/commander-spaceman/me2lzo/compress"
    "github.com/commander-spaceman/me2lzo/decompress"
)

// Decompress an LZO1X block
n, err := decompress.Decompress(compressedBlock, outputBuffer)

// Compress data with LZO1X-1
compressed, err := compress.Compress(originalData)
```

## API

### Decompress

```go
func Decompress(src, dst []byte) (int, error)
```

Decompresses LZO1X data from `src` into `dst`. Returns the number of bytes
written to `dst`. The caller must ensure `dst` is large enough.

### Compress

```go
func Compress(src []byte) ([]byte, error)
```

Compresses `src` using LZO1X-1 (fast mode). Returns the compressed data.
The output includes a valid LZO1X stream terminator.

## Testing

```bash
go test ./tests/ -v
```

Tests include:

- Round-trip compression/decompression across multiple data patterns
- Cross-library validation against `anchore/go-lzo`
- Real-file validation against 428 LZO blocks from 6 ME2 OT PCC files
- Edge cases: empty input, truncated streams, output overrun

## License

MIT
