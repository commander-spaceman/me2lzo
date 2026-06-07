package compress

func Compress(src []byte) ([]byte, error) {
	srcLen := len(src)
	if srcLen <= maxMatchInputPad {
		out := make([]byte, 0, srcLen+32)
		if srcLen > 0 {
			out = encodeLiteral(out, src, true, false)
		}
		out = append(out, markerM4|1, 0, 0)
		return out, nil
	}

	dict := newDictionary()
	inputLimit := srcLen - maxMatchInputPad

	out := make([]byte, 0, srcLen+srcLen/8+32)

	pos := 4
	literalStart := 0
	firstBlock := true

	for {
		dictIdx := hash4(src, pos)
		matchPos := int(dict[dictIdx]) - 1

		matched := false
		if matchPos >= 0 && pos != matchPos && pos-matchPos <= maxOffsetM4 {
			matchOffset := pos - matchPos
			if matchOffset <= maxOffsetM2 || src[matchPos+3] == src[pos+3] {
				if src[matchPos] == src[pos] &&
					src[matchPos+1] == src[pos+1] &&
					src[matchPos+2] == src[pos+2] {
					dict[dictIdx] = int32(pos + 1)

					if pos != literalStart {
						out = encodeLiteral(out, src[literalStart:pos], firstBlock, true)
						firstBlock = false
						literalStart = pos
					}

					matchLen := 3
					limit := srcLen - 1
					posBase := pos + 3
					for i := posBase; i < limit; i++ {
						if src[matchPos+matchLen] != src[i] {
							break
						}
						matchLen++
					}
					pos += matchLen

					out = encodeMatch(out, matchOffset, matchLen)
					literalStart = pos
					matched = true
				}
			}
		}

		if matched {
			if pos >= inputLimit {
				break
			}
			continue
		}

		dict[dictIdx] = int32(pos + 1)
		pos += 1 + (pos-literalStart)>>5
		if pos >= inputLimit {
			break
		}
	}

	tailLen := srcLen - literalStart
	if tailLen > 0 {
		out = encodeLiteral(out, src[literalStart:srcLen], firstBlock, !firstBlock)
	}

	out = append(out, markerM4|1, 0, 0)
	return out, nil
}

func encodeMatch(out []byte, offset, length int) []byte {
	switch {
	case offset <= maxOffsetM2 && length <= maxLenM2:
		return encodeM2(out, offset, length)
	case offset <= maxOffsetM3:
		return encodeM3(out, offset, length)
	default:
		return encodeM4(out, offset, length)
	}
}

func encodeLiteral(out []byte, lit []byte, firstBlock, hasPrevMatch bool) []byte {
	n := len(lit)
	if n == 0 {
		return out
	}

	switch {
	case firstBlock && n <= 238:
		out = append(out, byte(17+n))
	case !firstBlock && hasPrevMatch && n <= 3:
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
