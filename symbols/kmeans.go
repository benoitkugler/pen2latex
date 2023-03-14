package symbols

import (
	"math"
	"sort"
)

func meanAndStd(values []fl) (mean, std fl) {
	var esp2 fl
	for _, v := range values {
		mean += v
		esp2 += v * v
	}
	mean /= fl(len(values))
	esp2 /= fl(len(values))
	std = fl(math.Sqrt(float64(esp2 - mean*mean)))
	return mean, std
}

// adjust the scale and build Pos array
func anglesToPos(angles []fl) []Pos {
	N := fl(len(angles))
	min, _ := minMax(angles)
	toSegment := make([]Pos, len(angles))
	for i, a := range angles {
		toSegment[i] = Pos{X: float32(i), Y: N * (a - min) / 360}
	}
	return toSegment
}

type clusterRange [2]int // start, end (excluded, in slice syntax)

func segmentByAngleBreak(angles []fl, outliers map[int]bool) []clusterRange {
	const angleBreak = 75

	var currentRangeStart int
	var out []clusterRange

	push := func(cr clusterRange) {
		clusterSize := cr[1] - cr[0] + 1
		if clusterSize <= 3 { // artefacts
			return
		}
		out = append(out, cr)
	}

	var (
		// ignoring outliers
		previous      fl
		previousIndex int
	)
	for i, current := range angles {
		if outliers[i] {
			continue
		}
		if i == 0 {
			previous = current
			continue
		}
		if abs(current-previous) >= angleBreak { // new cluster, push the previous
			push(clusterRange{currentRangeStart, previousIndex + 1})
			currentRangeStart = i
		}
		previous = current
		previousIndex = i
	}
	if currentRangeStart < len(angles)-1 {
		push(clusterRange{currentRangeStart, len(angles)})
	}

	return out
}

// kmean to fit lines

type clusters []int // cluster for each point

// return the number of clusters
func (cl clusters) k() (K int) {
	if len(cl) == 0 {
		return 0
	}

	for _, k := range cl {
		if k > K {
			K = k
		}
	}
	return K + 1
}

func (cl clusters) segment(points []Pos) ([][]Pos, [][]int) {
	out := make([][]Pos, cl.k())
	mappingToIndex := make([][]int, cl.k())
	for i := range cl {
		k, p := cl[i], points[i]
		out[k] = append(out[k], p)
		mappingToIndex[k] = append(mappingToIndex[k], i)
	}
	return out, mappingToIndex
}

type line struct {
	a, b   fl
	center Pos
}

func (l line) distance(p Pos) fl {
	distLine := p.Y - (l.a*p.X + l.b)
	distLine = distLine * distLine
	distCenter := p.Sub(l.center).NormSquared()

	return distLine + 0.001*distCenter
}

// weights must sum to 1
func fitLineWeighted(points []Pos, weights []fl) (line, fl) {
	if len(points) == 1 {
		return line{1, points[0].Y, points[0]}, 0
	}

	var sx, sy, sxy, sx2 fl // 1// N Sum(x), ...
	var center Pos
	for i, p := range points {
		x, y := p.X, p.Y
		w := weights[i]

		sx += x * w
		sy += y * w
		sxy += x * y * w
		sx2 += x * x * w

		center.X += x * w
		center.Y += y * w
	}
	// y = ax + b
	a := (sxy - sx*sy) / (sx2 - sx*sx)
	b := sy - a*sx

	li := line{a, b, center}
	// fit error
	N := fl(len(points))
	var err fl
	for _, p := range points {
		err += li.distance(p)
	}
	err /= N

	return li, err
}

func assignClusters(points []Pos, lines []line, out clusters) (hasChanged bool) {
	for i, p := range points {
		// find the best line
		// enforcing sequentiallity using the fact that points have sorted X values

		if i == 0 {
			out[i] = 0
			continue
		}
		previousK := out[i-1]
		bestK := previousK
		if previousK != len(lines)-1 {
			// choose between the cluster of the previous point or the next
			d1, d2 := lines[previousK].distance(p), lines[previousK+1].distance(p)
			if d2 < d1 {
				bestK = previousK + 1
			}
		}

		if out[i] != bestK {
			hasChanged = true
		}

		out[i] = bestK
	}

	return
}

func computeLines(points []Pos, cls clusters, inOut []line) fl {
	byClusters, _ := cls.segment(points)
	var err fl
	for k, clusterPoints := range byClusters {

		// compute the weights from the distance to the center
		weights := make([]fl, len(clusterPoints))
		var totalWeigth fl
		for i, p := range clusterPoints {
			dx := abs(inOut[k].center.X - p.X)
			w := 1 / (10 + dx*sqrt(dx))
			weights[i] = w
			totalWeigth += w
		}
		for i := range weights { // normalize
			weights[i] /= totalWeigth
		}

		var clusterErr fl
		inOut[k], clusterErr = fitLineWeighted(clusterPoints, weights)
		err += clusterErr
	}
	return err
}

type kmResult struct {
	points   []Pos
	lines    []line
	clusters clusters
	fitError fl
}

// average cluster error
func (km kmResult) error() fl {
	var totalErr fl
	clusters, _ := km.clusters.segment(km.points)
	for k, points := range clusters {
		cl := km.lines[k]
		var clusterErr fl
		for _, p := range points {
			clusterErr += cl.distance(p)
		}
		Nk := fl(len(points))
		clusterErr /= Nk

		// penalize small clusters
		clusterErr *= 0.5 * (1 + 2/Nk)

		totalErr += clusterErr
	}
	return totalErr
}

func kmeans(points []Pos, K int) kmResult {
	// start with uniform repartition of centers
	L := len(points)
	step := L / K
	lines := make([]line, K)
	for i := 0; i < K; i++ {
		start, end := i*step, (i+1)*step
		segment := points[start:end]
		weights := make([]fl, len(segment))
		for j := range weights {
			weights[j] = 1. / fl(len(segment))
		}
		lines[i], _ = fitLineWeighted(segment, weights)
	}
	cls := make(clusters, L)

	var fitError fl
	const iterMax = 100
	for i, hasChanged := 0, true; i < iterMax && hasChanged; i++ {
		hasChanged = assignClusters(points, lines, cls)
		fitError = computeLines(points, cls, lines)
	}

	return kmResult{points: points, lines: lines, clusters: cls, fitError: fitError}
}

func bestKmeans(angles []fl) kmResult {
	const KMax = 5

	toSegment := anglesToPos(angles)

	// run kmeans for each K and pick the best error to detect outliers
	bestWCSSK, bestWCSS := kmResult{}, inf
	for K := 1; K <= KMax; K++ {
		kmeansOut := kmeans(toSegment, K)

		if e := kmeansOut.error(); e < bestWCSS {
			bestWCSS = e
			bestWCSSK = kmeansOut
		}
	}

	return bestWCSSK
}

type indexedDistances struct {
	distances []fl
	indices   []int
}

func newIndexedDistances(distances []fl) indexedDistances {
	indices := make([]int, len(distances))
	for i := range indices {
		indices[i] = i
	}
	return indexedDistances{distances, indices}
}

func (a indexedDistances) Len() int { return len(a.distances) }
func (a indexedDistances) Swap(i, j int) {
	a.distances[i], a.distances[j] = a.distances[j], a.distances[i]
	a.indices[i], a.indices[j] = a.indices[j], a.indices[i]
}
func (a indexedDistances) Less(i, j int) bool { return a.distances[i] < a.distances[j] }

func outliers(distances []fl) []int {
	sorted := newIndexedDistances(distances)
	sort.Sort(sorted) // distances is mutated

	// starting at the median, incrementaly
	// test each value for outliers
	i := len(distances) / 2
	mean, std := meanAndStd(sorted.distances[:i])
	mean2 := std*std + mean*mean // used to compute the std by recursion
	for ; i < len(distances); i++ {
		d := distances[i]

		thresholdMax := mean + 4*std // 4 is tuned experimentally

		if d > thresholdMax { // found outlier, we can break here since the slice is sorted
			break
		} else { // add the new value to the accepted serie
			N := fl(i + 1) // number of points before adding the new
			mean = (N*mean + d) / (N + 1)
			mean2 = (N*mean2 + d*d) / (N + 1)
			std = sqrt(mean2 - mean*mean)
		}
	}
	// reject indices starting at i (maybe empty)
	return sorted.indices[i:]
}

func (km kmResult) outliers() map[int]bool {
	out := make(map[int]bool)
	clusters, mappingToIndex := km.clusters.segment(km.points)
	for k, points := range clusters {
		cl := km.lines[k]
		pointIndices := mappingToIndex[k]

		if len(points) <= 2 { // consider a degenerated cluster as outliers
			for _, i := range pointIndices {
				out[i] = true
			}
			continue
		}

		distances := make([]fl, len(points))
		for i, p := range points {
			distances[i] = cl.distance(p)
		}

		outls := outliers(distances)
		for _, i := range outls { // map back to the input indices
			out[pointIndices[i]] = true
		}
	}

	return out
}

// segmentation filters outliers and segments the resulting
// points
// outlier are not present in the returned shapes
func (sh Shape) segment() (out []Shape) {
	if len(sh) < 2 {
		return Symbol{sh}
	}

	angles := sh.smooth().directions()

	// compute the sub shapes
	km := bestKmeans(angles)
	outls := km.outliers()

	clusters := segmentByAngleBreak(angles, outls)

	// return the subslices
	for _, cl := range clusters {
		subShape := sh[cl[0]:cl[1]]
		out = append(out, subShape)
	}
	return out
}
