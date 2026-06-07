package compress

const (
	markerM1 = 0
	markerM2 = 64
	markerM3 = 32
	markerM4 = 16
)

func encodeLiteral(out []byte, lit []byte, firstBlock bool) []byte {
	n := len(lit)
	if n == 0 {
		return out
	}

	switch {
	case firstBlock && n <= 238:
		out = append(out, byte(17+n))
	case n <= 3:
		out[len(out)-2] |= byte(n)
	case n <= 18:
		out = append(out, byte(n-3))
	default:
		out = append(out, 0)
		out = encodeVarLen(out, n-18)
	}

	out = append(out, lit...)
	return out
}

func encodeM2(out []byte, offset, length int) []byte {
	offset--
	out = append(out,
		byte(((length-1)<<5)|((offset&0x7)<<2)),
		byte(offset>>3),
	)
	return out
}

func encodeM3(out []byte, offset, length int) []byte {
	offset--
	if length <= maxLenM3 {
		out = append(out, byte(markerM3|(length-2)))
	} else {
		out = append(out, byte(markerM3))
		out = encodeVarLen(out, length-maxLenM3)
	}
	out = append(out, byte((offset&63)<<2), byte(offset>>6))
	return out
}

func encodeM4(out []byte, offset, length int) []byte {
	offset -= 0x4000
	if length <= maxLenM4 {
		out = append(out, byte(markerM4|((offset&0x4000)>>11)|(length-2)))
	} else {
		out = append(out, byte(markerM4|((offset&0x4000)>>11)))
		out = encodeVarLen(out, length-maxLenM4)
	}
	out = append(out, byte((offset&63)<<2), byte(offset>>6))
	return out
}

func encodeVarLen(out []byte, t int) []byte {
	for t > 255 {
		out = append(out, 0)
		t -= 255
	}
	out = append(out, byte(t))
	return out
}
