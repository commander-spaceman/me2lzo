package tests

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	anchore "github.com/anchore/go-lzo"
	"github.com/commander-spaceman/me2lzo/decompress"
)

const (
	compressionLZO   = 0x2
	chunkHeaderMagic = 0x9E2A83C1
	chunkHeaderSize  = 16
	blockHeaderSize  = 8
	maxBlockSizeOT   = 0x20000
)

func TestDecompressRealPCCBlocks(t *testing.T) {
	pccDir := filepath.Join("..", "..", "pcc-toolkit", "output")
	entries, err := os.ReadDir(pccDir)
	if err != nil {
		t.Skipf("pcc-toolkit/output not found: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pcc" {
			continue
		}

		pccPath := filepath.Join(pccDir, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			testPCCFile(t, pccPath)
		})
	}
}

func testPCCFile(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	blocks, err := extractLZOBlocks(data)
	if err != nil {
		t.Fatalf("extract blocks: %v", err)
	}

	t.Logf("found %d LZO blocks", len(blocks))
	if len(blocks) == 0 {
		t.Fatal("no LZO blocks found")
	}

	for i, block := range blocks {
		anchoreOut := make([]byte, block.uncompressedSize*2)
		anchoreN, anchoreErr := anchore.Decompress(block.data, anchoreOut)

		me2lzoOut := make([]byte, block.uncompressedSize*2)
		me2lzoN, me2lzoErr := decompress.Decompress(block.data, me2lzoOut)

		if (anchoreErr == nil) != (me2lzoErr == nil) {
			t.Errorf("block %d: error mismatch: anchore=%v me2lzo=%v", i, anchoreErr, me2lzoErr)
			continue
		}
		if anchoreErr != nil {
			t.Logf("block %d: both failed as expected: %v", i, anchoreErr)
			continue
		}
		if anchoreN != me2lzoN {
			t.Errorf("block %d: size mismatch: anchore=%d me2lzo=%d", i, anchoreN, me2lzoN)
			continue
		}
		for j := 0; j < anchoreN; j++ {
			if anchoreOut[j] != me2lzoOut[j] {
				t.Errorf("block %d: byte mismatch at offset %d: anchore=%02x me2lzo=%02x",
					i, j, anchoreOut[j], me2lzoOut[j])
				break
			}
		}
	}
}

type lzoBlock struct {
	data             []byte
	uncompressedSize int
}

func extractLZOBlocks(data []byte) ([]lzoBlock, error) {
	cursor, err := locateCompressionInfo(data)
	if err != nil {
		return nil, err
	}

	if cursor+8 > len(data) {
		return nil, fmt.Errorf("compression header out of range")
	}

	compressionType := readI32(data, cursor)
	numChunks := readI32(data, cursor+4)
	cursor += 8

	if compressionType != compressionLZO {
		return nil, fmt.Errorf("not LZO compressed (type=%d)", compressionType)
	}
	if numChunks <= 0 {
		return nil, fmt.Errorf("invalid chunk count: %d", numChunks)
	}

	type chunkInfo struct {
		uncompressedSize int
		compressedOffset int
		compressedSize   int
	}
	chunks := make([]chunkInfo, 0, numChunks)
	for i := 0; i < numChunks; i++ {
		if cursor+16 > len(data) {
			return nil, fmt.Errorf("chunk table out of range")
		}
		chunks = append(chunks, chunkInfo{
			uncompressedSize: readI32(data, cursor+4),
			compressedOffset: readI32(data, cursor+8),
			compressedSize:   readI32(data, cursor+12),
		})
		cursor += 16
	}

	var blocks []lzoBlock
	for _, c := range chunks {
		if c.compressedOffset+c.compressedSize > len(data) {
			return nil, fmt.Errorf("compressed chunk out of range")
		}
		chunkBlob := data[c.compressedOffset : c.compressedOffset+c.compressedSize]
		if len(chunkBlob) < chunkHeaderSize {
			return nil, fmt.Errorf("truncated chunk header")
		}

		magic := readU32(chunkBlob, 0)
		blockSize := readI32(chunkBlob, 4)
		compressedSizeHeader := readI32(chunkBlob, 8)
		uncompressedSizeHeader := readI32(chunkBlob, 12)

		if magic != chunkHeaderMagic {
			return nil, fmt.Errorf("invalid chunk magic")
		}
		if uncompressedSizeHeader != c.uncompressedSize {
			return nil, fmt.Errorf("chunk size mismatch")
		}
		if compressedSizeHeader+chunkHeaderSize > c.compressedSize {
			return nil, fmt.Errorf("truncated chunk payload")
		}
		if blockSize <= 0 || blockSize > maxBlockSizeOT {
			return nil, fmt.Errorf("invalid block size: %d", blockSize)
		}

		blockCount := uncompressedSizeHeader / blockSize
		if uncompressedSizeHeader%blockSize != 0 {
			blockCount++
		}
		blockTableOffset := chunkHeaderSize
		blockDataOffset := blockTableOffset + blockCount*blockHeaderSize
		if blockDataOffset > len(chunkBlob) {
			return nil, fmt.Errorf("invalid block table")
		}

		for i := 0; i < blockCount; i++ {
			hdrOff := blockTableOffset + i*blockHeaderSize
			if hdrOff+8 > len(chunkBlob) {
				return nil, fmt.Errorf("block header out of range")
			}
			compSz := readI32(chunkBlob, hdrOff)
			uncompSz := readI32(chunkBlob, hdrOff+4)
			if compSz < 0 || uncompSz < 0 || uncompSz > blockSize {
				return nil, fmt.Errorf("invalid block sizes")
			}
			if compSz > 0 {
				blocks = append(blocks, lzoBlock{
					data:             chunkBlob[blockDataOffset : blockDataOffset+compSz],
					uncompressedSize: uncompSz,
				})
			}
			blockDataOffset += compSz
		}
	}

	return blocks, nil
}

func locateCompressionInfo(data []byte) (int, error) {
	cursor := 8
	if cursor+4 > len(data) {
		return 0, fmt.Errorf("truncated header")
	}
	cursor += 4
	folderLen := readI32(data, cursor)
	cursor += 4
	if folderLen > 0 {
		cursor += folderLen
	} else if folderLen < 0 {
		cursor += (-folderLen) * 2
	}
	cursor += 4
	cursor += 24
	cursor += 4
	cursor += 16

	if cursor+4 > len(data) {
		return 0, fmt.Errorf("truncated generations")
	}
	generations := int(readU32(data, cursor))
	cursor += 4
	if generations > 0 {
		cursor += 12
		cursor += (generations - 1) * 12
	}
	cursor += 8
	cursor += 16
	cursor += 8
	if cursor < 0 || cursor > len(data) {
		return 0, fmt.Errorf("compression info out of range")
	}
	return cursor, nil
}

func readU32(b []byte, off int) uint32 {
	return binary.LittleEndian.Uint32(b[off : off+4])
}

func readI32(b []byte, off int) int {
	return int(int32(binary.LittleEndian.Uint32(b[off : off+4])))
}
