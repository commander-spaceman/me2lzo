# Project Map

**Purpose:** Pure Go implementation of LZO1X compression and decompression for Mass Effect 2 Original Trilogy `.pcc` package files. Built as the LZO dependency for [pcc-toolkit](https://github.com/commander-spaceman/pcc-toolkit).

## Notes for AI Agents

- **Entry points:** `compress.Compress()` (`compress/compress.go:3`), `decompress.Decompress()` (`decompress/decompress.go:31`). The top-level module exposes nothing; callers import `compress` or `decompress` directly.
- **Main patterns:** Two independent packages with a single exported function each. No shared internal package. Tests live under `tests/` and import both packages plus external reference libraries for cross-validation.
- **General rule:** Read this file before proposing structural changes or modifying multiple modules. The compress and decompress packages are designed to be independent — changes in one should not need changes in the other.

---

## 1. Decompress

LZO1X-1 decompression. Byte-identical to the Oberhumer reference implementation on all tested ME2 OT blocks. No stream terminators or framing.

```text
decompress/
  buffer.go         # buf helper, copyLiterals, copyLookbehind, countZeroBytes
  decompress.go     # Entry point: Decompress(), decoder struct, main decode loop
  literal.go        # First-byte literal handling, M1 long/short copy handlers
  match.go          # M2, M3, M4 match instruction handlers
```

**Main responsibilities:**

- Decode raw LZO1X-1 compressed blocks into decompressed output
- Handle all four match types (M1–M4) and variable-length literals
- Validate bounds to prevent overruns and detect truncated input

**Key files:**

- `decompress/decompress.go`: Exported `Decompress(src, dst []byte) (int, error)` — main entry point and main decode loop
- `decompress/match.go`: M2/M3/M4 instruction decoding (back-references with varying offset/length ranges)
- `decompress/literal.go`: First-byte literal decoding, M1 long literal and short copy
- `decompress/buffer.go`: Byte buffer abstraction with bounds-checked reading and byte-copy operations

**Relationships:**

- Standalone — depends only on `encoding/binary` (stdlib)
- Referenced by tests for round-trip validation and cross-library comparison against `anchore/go-lzo`

---

## 2. Compress

LZO1X-1 compression (fast mode). Produces valid LZO1X blocks with a stream terminator.

```text
compress/
  compress.go       # Entry point: Compress(), main compression loop, encodeLiteral, encodeMatch
  encode.go         # LZO instruction encoding: M2, M3, M4, variable-length
  hash.go           # 4-byte rolling hash, dictionary, findMatch helper
```

**Main responsibilities:**

- Compress arbitrary data using LZO1X-1 algorithm
- Emit LZO instructions (M2/M3/M4 matches, M1 literals) with valid stream terminator
- Maintain a 14-bit hash dictionary for match finding

**Key files:**

- `compress/compress.go`: Exported `Compress(src []byte) ([]byte, error)` — compression entry point and main loop
- `compress/encode.go`: Instruction encoding for M2, M3, M4 match types and variable-length integer encoding
- `compress/hash.go`: 4-byte rolling hash, dictionary type (`[]int32`), and `findMatch` with two-attempt probing

**Relationships:**

- Standalone — depends only on stdlib
- Referenced by tests for round-trip compression/decompression validation

---

## 3. Tests

Integration and cross-library validation tests. Uses external reference libraries to verify correctness.

```text
tests/
  compress_test.go      # Round-trip, empty input, deterministic output
  decompress_test.go    # Round-trip via woozymasta/lzo, cross-library vs anchore/go-lzo, edge cases
  pcc_test.go           # Real-file validation against ME2 OT .pcc files
```

**Main responsibilities:**

- Validate round-trip compress → decompress for multiple data patterns
- Cross-validate decompress output against `anchore/go-lzo` (bit-identical)
- Test real ME2 OT `.pcc` LZO blocks extracted from `pcc-toolkit/output`
- Edge case coverage: empty input, truncated streams, output buffer overrun

**Key files:**

- `tests/compress_test.go`: Compress-then-decompress round-trip tests with multiple data patterns; deterministic output verification
- `tests/decompress_test.go`: Decompress tests using `woozymasta/lzo` as reference compressor and `anchore/go-lzo` for output comparison
- `tests/pcc_test.go`: Parses real ME2 OT `.pcc` chunk/block structure and validates every LZO block against `anchore/go-lzo`

**Relationships:**

- Depends on `compress`, `decompress`, `anchore/go-lzo`, `woozymasta/lzo`
- Reads external `.pcc` test data from `../../pcc-toolkit/output` (skipped if unavailable, not committed to this repo)

---

## 4. Root Configuration

```text
go.mod              # Module: github.com/commander-spaceman/me2lzo, Go 1.25.5
go.sum              # Dependency checksums
LICENSE             # MIT
README.md           # Usage, API docs, testing instructions
.gitignore          # Go binary/test artifacts, IDE files
```

**Main responsibilities:**

- Module definition and dependency management
- Project documentation and licensing

**Key files:**

- `go.mod`: Declares the module path and indirect test dependencies (`anchore/go-lzo`, `woozymasta/lzo`)
- `README.md`: Installation, usage examples, API signatures, and testing instructions
