package symbols

import "math"

// implementation of the discrete Fréchet distance,
// an approximation of the Fréchet metric for polygonal curves, defined by Eiter and Mannila in
// Eiter, Thomas; Mannila, Heikki (1994), Computing discrete Fréchet distance
// http://www.kr.tuwien.ac.at/staff/eiter/et-archive/cdtr9464.pdf

// Discrete Fréchet distance (positive, lower is better).
// See Eiter and Mannila (1994), Table 1: Algorithm computing the coupling measure
func frechetDistanceShapes(u, v Shape) fl {
	p, q := len(u), len(v)
	// couplings, initialized to -1
	ca := make([][]fl, p)
	for i := range ca {
		ca[i] = make([]fl, q)
		for j := range ca[i] {
			ca[i][j] = -1
		}
	}

	var computeCoupling func(i, j int) fl
	computeCoupling = func(i, j int) fl {
		if ca[i][j] > -1 {
			return ca[i][j]
		} else if i == 0 && j == 0 {
			ca[i][j] = distancePoints(u[0], v[0])
		} else if i > 0 && j == 0 {
			ca[i][j] = max(computeCoupling(i-1, 0), distancePoints(u[i], v[0]))
		} else if i == 0 && j > 0 {
			ca[i][j] = max(computeCoupling(0, j-1), distancePoints(u[0], v[j]))
		} else if i > 0 && j > 0 {
			d1 := computeCoupling(i-1, j)
			d2 := computeCoupling(i-1, j-1)
			d3 := computeCoupling(i, j-1)
			d4 := distancePoints(u[i], v[j])
			ca[i][j] = max(min(min(d1, d2), d3), d4)
		}
		return ca[i][j]
	}
	return computeCoupling(p-1, q-1)
}

// return the distance between the closest point between [u] and [v]
// This NOT a measure of similarity
func closestPointDistance(u, v Shape) fl {
	best := fl(math.Inf(1))
	for _, pu := range u {
		for _, pv := range v {
			if d := distancePoints(pu, pv); d < best {
				best = d
			}
		}
	}
	return best
}
