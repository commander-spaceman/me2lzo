package decompress

import "errors"

var (
	ErrLookbehindOverrun   = errors.New("me2lzo: lookbehind overrun")
	ErrOutputOverrun       = errors.New("me2lzo: output overrun")
	ErrInputOverrun        = errors.New("me2lzo: input overrun")
	ErrDecompressionFailed = errors.New("me2lzo: decompression failed")
	ErrInputNotConsumed    = errors.New("me2lzo: input not fully consumed")
)

const (
	m3Marker = 0x20
	m4Marker = 0x10
)

type decoder struct {
	src, dst         buf
	state, nextState int
	lbIdx, lbLen     int
}

func newDecoder(srcBytes, dstBytes []byte) *decoder {
	return &decoder{
		src: newBuf(srcBytes),
		dst: newBuf(dstBytes),
	}
}

func Decompress(src, dst []byte) (int, error) {
	d := newDecoder(src, dst)

	if d.src.remaining() < 3 {
		return 0, ErrInputOverrun
	}

	if err := d.handleFirstByte(); err != nil {
		return d.dst.idx, err
	}

	for {
		if d.src.idx >= d.src.end {
			break
		}
		inst := d.src.readByte()
		switch {
		case inst&0xC0 != 0:
			if err := d.handleM2(inst); err != nil {
				return d.dst.idx, err
			}
		case inst&m3Marker != 0:
			if err := d.handleM3(inst); err != nil {
				return d.dst.idx, err
			}
		case inst&m4Marker != 0:
			finished, err := d.handleM4(inst)
			if err != nil {
				return d.dst.idx, err
			}
			if finished {
				goto done
			}
		default:
			switch d.state {
			case 0:
				if err := d.handleM1LongLiteral(inst); err != nil {
					return d.dst.idx, err
				}
				d.state = 4
				continue
			default:
				if err := d.handleM1ShortCopy(inst); err != nil {
					return d.dst.idx, err
				}
			}
		}

		if d.lbIdx < 0 {
			return d.dst.idx, ErrLookbehindOverrun
		}
		if d.src.idx+d.nextState > d.src.end {
			return d.dst.idx, ErrInputOverrun
		}
		if d.dst.idx+d.lbLen+d.nextState > d.dst.end {
			return d.dst.idx, ErrOutputOverrun
		}

		d.copyLookbehind()
		d.copyLiterals(d.nextState)
	}

done:
	if d.lbLen != 3 {
		return d.dst.idx, ErrDecompressionFailed
	}
	if d.src.idx == d.src.end {
		return d.dst.idx, nil
	}
	if d.src.idx < d.src.end {
		return d.dst.idx, ErrInputNotConsumed
	}
	return d.dst.idx, ErrInputOverrun
}
