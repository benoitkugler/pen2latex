package symbols

// distinguish using size

type Context struct {
	Box      Rect
	Baseline Fl // Y coordinate
}

func distinguishByContext(fp Footprint, context Context, r rune) rune {
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
func isOverBaseline(fp Footprint, context Context) bool {
	bottom := fp.BoundingBox().LR.Y
	contextHeight := context.Box.Height()

	return bottom <= context.Baseline+contextHeight*0.1
}

// return true if the symbol is more than half the height (between top and baseline)
func isUpperSize(fp Footprint, context Context) bool {
	h := abs(context.Box.UL.Y - context.Baseline)
	bbox := fp.BoundingBox()
	ratio := bbox.Height() / h
	return ratio >= 0.5
}
