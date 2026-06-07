package compress

const (
	dictBits = 14
	dictMask = (1 << dictBits) - 1
	dictHigh = (dictMask >> 1) + 1

	maxOffsetM1 = 0x0400
	maxOffsetM2 = 0x0800
	maxOffsetM3 = 0x4000
	maxOffsetM4 = 0xbfff

	minLenM2 = 3
	maxLenM2 = 8
	maxLenM3 = 33
	maxLenM4 = 9

	maxMatchInputPad = maxLenM2 + 5
)

type dictionary []int32

func newDictionary() dictionary {
	return make([]int32, 1<<dictBits)
}

func hash4(in []byte, pos int) int {
	key := int(in[pos+3])
	key = (key << 6) ^ int(in[pos+2])
	key = (key << 5) ^ int(in[pos+1])
	key = (key << 5) ^ int(in[pos+0])
	return ((0x21 * key) >> 5) & dictMask
}

type match struct {
	pos    int
	offset int
	len    int
}

func findMatch(dict dictionary, in []byte, pos int) (match, bool) {
	dictIdx := hash4(in, pos)

	for attempt := 0; attempt < 2; attempt++ {
		matchPos := int(dict[dictIdx]) - 1
		if matchPos < 0 || pos == matchPos || pos-matchPos > maxOffsetM4 {
			if attempt == 0 {
				dictIdx = (dictIdx & (dictMask & 0x7ff)) ^ (dictHigh | 0x1f)
			}
			continue
		}

		matchOffset := pos - matchPos
		if matchOffset > maxOffsetM2 && in[matchPos+3] != in[pos+3] {
			if attempt == 0 {
				dictIdx = (dictIdx & (dictMask & 0x7ff)) ^ (dictHigh | 0x1f)
			}
			continue
		}

		if in[matchPos] != in[pos] || in[matchPos+1] != in[pos+1] || in[matchPos+2] != in[pos+2] {
			if attempt == 0 {
				dictIdx = (dictIdx & (dictMask & 0x7ff)) ^ (dictHigh | 0x1f)
			}
			continue
		}

		matchLen := 3
		end := len(in) - 1
		for i := pos + 3; i < end; i++ {
			if in[matchPos+matchLen] != in[i] {
				break
			}
			matchLen++
		}

		return match{pos: matchPos, offset: matchOffset, len: matchLen}, true
	}

	return match{}, false
}
