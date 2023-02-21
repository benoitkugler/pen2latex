package symbols

type Bezier struct {
	P0, P1, P2, P3 Pos
}

func (knots Shape) ToBezierCurves() []Bezier {
	if len(knots) <= 1 {
		return nil
	}

	n := len(knots) - 1
	if n == 1 { // Special case: Bezier curve should be a straight line.
		var P1, P2 Pos
		// 3P1 = 2P0 + P3
		P1.X = (2*knots[0].X + knots[1].X) / 3
		P1.Y = (2*knots[0].Y + knots[1].Y) / 3

		// P2 = 2P1 – P0
		P2.X = 2*P1.X - knots[0].X
		P2.Y = 2*P1.Y - knots[0].Y
		return []Bezier{{P0: knots[0], P1: P1, P2: P2, P3: knots[1]}}
	}

	// Calculate first Bezier control points
	// Right hand side vector
	rhs := make([]fl, n)

	// Set right hand side X values
	for i := 1; i < n-1; i++ {
		rhs[i] = 4*knots[i].X + 2*knots[i+1].X
	}
	rhs[0] = knots[0].X + 2*knots[1].X
	rhs[n-1] = (8*knots[n-1].X + knots[n].X) / 2.0
	// Get first control points X-values
	x := getFirstControlPoints(rhs)

	// Set right hand side Y values
	for i := 1; i < n-1; i++ {
		rhs[i] = 4*knots[i].Y + 2*knots[i+1].Y
	}
	rhs[0] = knots[0].Y + 2*knots[1].Y
	rhs[n-1] = (8*knots[n-1].Y + knots[n].Y) / 2.0
	// Get first control points Y-values
	y := getFirstControlPoints(rhs)

	// Fill output arrays.
	out := make([]Bezier, n)
	for i := range out {
		// First control point
		P1 := Pos{x[i], y[i]}
		// Second control point
		var P2 Pos
		if i < n-1 {
			P2 = Pos{2*knots[i+1].X - x[i+1], 2*knots[i+1].Y - y[i+1]}
		} else {
			P2 = Pos{(knots[n].X + x[n-1]) / 2, (knots[n].Y + y[n-1]) / 2}
		}
		out[i] = Bezier{P0: knots[i], P1: P1, P2: P2, P3: knots[i+1]}
	}
	return out
}

// Solves a tridiagonal system for one of coordinates (x or y)
// of first Bezier control points.
func getFirstControlPoints(rhs []fl) []float32 {
	N := len(rhs)
	x := make([]fl, N)   // Solution vector.
	tmp := make([]fl, N) // Temp workspace.

	var b fl = 2.0
	x[0] = rhs[0] / b
	for i := 1; i < N; i++ { // Decomposition and forward substitution.
		tmp[i] = 1 / b
		if i < N-1 {
			b = 4 - tmp[i]
		} else {
			b = 3.5 - tmp[i]
		}
		x[i] = (rhs[i] - x[i-1]) / b
	}
	for i := 1; i < N; i++ { // Backsubstitution.
		x[N-i-1] -= tmp[N-i] * x[N-i]
	}

	return x
}
