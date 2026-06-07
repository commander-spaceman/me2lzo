package decompress

func (d *decoder) handleM2(inst byte) error {
	if d.src.idx+1 > d.src.end {
		return ErrInputOverrun
	}
	extra := d.src.readByte()
	d.lbIdx = d.dst.idx - ((int(extra) << 3) + int((inst>>2)&0x7) + 1)
	d.lbLen = int(inst>>5) + 1
	d.nextState = int(inst & 0x3)
	return nil
}

func (d *decoder) handleM3(inst byte) error {
	d.lbLen = int(inst&0x1f) + 2
	if d.lbLen == 2 {
		offset, err := d.countZeroBytes(31)
		if err != nil {
			return err
		}
		d.lbLen += offset
	}
	if d.src.idx+2 > d.src.end {
		return ErrInputOverrun
	}
	val := int(d.src.readU16LE())
	d.lbIdx = d.dst.idx - (val>>2 + 1)
	d.nextState = val & 0x3
	return nil
}

func (d *decoder) handleM4(inst byte) (finished bool, err error) {
	d.lbLen = int(inst&0x7) + 2
	if d.lbLen == 2 {
		offset, err := d.countZeroBytes(7)
		if err != nil {
			return false, err
		}
		d.lbLen += offset
	}
	if d.src.idx+2 > d.src.end {
		return false, ErrInputOverrun
	}
	val := int(d.src.readU16LE())
	d.lbIdx = d.dst.idx - (int(inst&0x8)<<11 + val>>2)
	d.nextState = val & 0x3

	if d.lbIdx == d.dst.idx {
		return true, nil
	}
	d.lbIdx -= 16384
	return false, nil
}
