package symbols

// this file handles symbols with variable width,
// for which we impose one look :
// sqrt : √ U+221A
// fraction bar

func (fp Stroke) IsLine() bool {
	return len(fp.Curves) == 1 && fp.Curves[0].IsRoughlyLinear()
}

func (fp Stroke) IsSqrt() bool {
	// we accept either a V or a U, followed by a line
	if L := len(fp.Curves); !(L == 2 || L == 3) {
		return false
	}

	last := fp.Curves[len(fp.Curves)-1]
	if !last.IsRoughlyLinear() {
		return false
	}

	s1, s2 := fp.Curves[0], fp.Curves[1]
	// match a V
	if len(fp.Curves) == 3 {
		if !(s1.IsRoughlyLinear() && s2.IsRoughlyLinear()) {
			return false
		}
		// check the angle
		angle := angle(s1.P0.Sub(s1.P3), s2.P3.Sub(s2.P0))
		return 0 <= angle && angle <= 45
	}
	// .. or a U (here L == 2)
	startX, endX := s1.P0.X, s1.P3.X
	return startX < endX && startX <= s1.P1.X && s1.P2.X <= endX &&
		s1.P1.Y > s1.P0.Y && s1.P2.Y > s1.P3.Y // +Y is downward
}
