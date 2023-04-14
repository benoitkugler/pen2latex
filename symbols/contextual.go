package symbols

// distinguish using size

// HeightGrid defines the reference for the Y coordinate.
type HeightGrid struct {
	Ymin, Ymax Fl
	Baseline   Fl // Y coordinate, with Ymin <= Baseline <= Ymax
}

func (ct HeightGrid) height() Fl { return ct.Ymax - ct.Ymin }

func distinguishByContext(fp Footprint, context HeightGrid, r rune) rune {
	switch r {
	case 'j', 'J':
		if isOverBaseline(fp, context) {
			return 'J'
		}
		return 'j'
	case 'o', 'O':
		if isUpperSize(fp, context) {
			return 'O'
		}
		return 'o'
	case 'p', 'P':
		if isUpperSize(fp, context) {
			return 'P'
		}
		return 'p'
	case 's', 'S':
		if isUpperSize(fp, context) {
			return 'S'
		}
		return 's'
	case 'u', 'U':
		if isUpperSize(fp, context) {
			return 'U'
		}
		return 'u'
	case 'v', 'V':
		if isUpperSize(fp, context) {
			return 'V'
		}
		return 'v'
	case 'w', 'W':
		if isUpperSize(fp, context) {
			return 'W'
		}
		return 'w'
	case 'z', 'Z':
		if isUpperSize(fp, context) {
			return 'Z'
		}
		return 'z'
	case 'π', 'Π':
		if isUpperSize(fp, context) {
			return 'Π'
		}
		return 'π'
	}

	return r
}

// return true if the symbol is over the baseline
func isOverBaseline(fp Footprint, context HeightGrid) bool {
	bottom := fp.BoundingBox().LR.Y
	contextHeight := context.height()

	return bottom <= context.Baseline+contextHeight*0.1
}

// return true if the symbol is more than half the height (between top and baseline)
func isUpperSize(fp Footprint, context HeightGrid) bool {
	h := abs(context.Ymin - context.Baseline)
	bbox := fp.BoundingBox()
	ratio := bbox.Height() / h
	return ratio >= 0.5
}
