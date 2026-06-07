package decompress

import "encoding/binary"

const max255Count = (^uint(0))/255 - 2

type buf struct {
	data     []byte
	idx, end int
}

func newBuf(data []byte) buf {
	return buf{data: data, idx: 0, end: len(data)}
}

func (b *buf) remaining() int {
	return b.end - b.idx
}

func (b *buf) readByte() byte {
	v := b.data[b.idx]
	b.idx++
	return v
}

func (b *buf) readU16LE() uint16 {
	v := binary.LittleEndian.Uint16(b.data[b.idx : b.idx+2])
	b.idx += 2
	return v
}

func (d *decoder) copyLiterals(n int) {
	for i := 0; i < n; i++ {
		d.dst.data[d.dst.idx+i] = d.src.data[d.src.idx+i]
	}
	d.dst.idx += n
	d.src.idx += n
}

func (d *decoder) copyLookbehind() {
	for i := 0; i < d.lbLen; i++ {
		d.dst.data[d.dst.idx+i] = d.dst.data[d.lbIdx+i]
	}
	d.dst.idx += d.lbLen
	d.state = d.nextState
}

func (d *decoder) countZeroBytes(factor int) (int, error) {
	start := d.src.idx
	for d.src.idx < d.src.end && d.src.data[d.src.idx] == 0 {
		d.src.idx++
	}
	count := d.src.idx - start
	if count > int(max255Count) {
		return -1, ErrDecompressionFailed
	}
	if d.src.idx+1 > d.src.end {
		return -1, ErrInputOverrun
	}
	offset := count*255 + factor + int(d.src.data[d.src.idx])
	d.src.idx++
	return offset, nil
}
