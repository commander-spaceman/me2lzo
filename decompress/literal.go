package decompress

func (d *decoder) handleFirstByte() error {
	switch {
	case d.src.data[d.src.idx] >= 22:
		n := int(d.src.readByte()) - 17
		if d.src.idx+n > d.src.end {
			return ErrInputOverrun
		}
		if d.dst.idx+n > d.dst.end {
			return ErrOutputOverrun
		}
		d.copyLiterals(n)
		d.state = 4

	case d.src.data[d.src.idx] >= 18:
		n := int(d.src.readByte()) - 17
		if d.src.idx+n > d.src.end {
			return ErrInputOverrun
		}
		if d.dst.idx+n > d.dst.end {
			return ErrOutputOverrun
		}
		d.copyLiterals(n)
		d.state = n
	}
	return nil
}

func (d *decoder) handleM1LongLiteral(inst byte) error {
	n := int(inst) + 3
	if n == 3 {
		offset, err := d.countZeroBytes(15)
		if err != nil {
			return err
		}
		n += offset
	}
	if d.src.idx+n > d.src.end {
		return ErrInputOverrun
	}
	if d.dst.idx+n > d.dst.end {
		return ErrOutputOverrun
	}
	d.copyLiterals(n)
	return nil
}

func (d *decoder) handleM1ShortCopy(inst byte) error {
	if d.src.idx+1 > d.src.end {
		return ErrInputOverrun
	}
	d.nextState = int(inst & 0x3)
	extra := d.src.readByte()
	if d.state != 4 {
		d.lbIdx = d.dst.idx - (int(inst>>2) + (int(extra) << 2) + 1)
		d.lbLen = 2
	} else {
		d.lbIdx = d.dst.idx - (int(inst>>2) + (int(extra) << 2) + 2049)
		d.lbLen = 3
	}
	return nil
}
