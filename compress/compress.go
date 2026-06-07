package compress

func Compress(src []byte) ([]byte, error) {
	srcLen := len(src)
	if srcLen <= maxMatchInputPad {
		out := make([]byte, 0, srcLen+32)
		if srcLen > 0 {
			out = encodeLiteral(out, src, true)
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
						out = encodeLiteral(out, src[literalStart:pos], firstBlock)
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
					pos = pos + matchLen

					switch {
					case matchOffset <= maxOffsetM2:
						out = encodeM2(out, matchOffset, matchLen)
					case matchOffset <= maxOffsetM3:
						out = encodeM3(out, matchOffset, matchLen)
					default:
						out = encodeM4(out, matchOffset, matchLen)
					}

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

	tail := srcLen - literalStart
	if tail > 0 {
		out = encodeLiteral(out, src[literalStart:literalStart+tail], firstBlock)
	}

	out = append(out, markerM4|1, 0, 0)
	return out, nil
}
