package symbols

import (
	"math"
)

type kmOut struct {
	points  []Pos // the input of the algorithm
	cls     clusters
	centers []Pos
}

func kmeans(points []Pos, K int) kmOut {
	if K > 255 {
		panic("K > 255")
	}

	// start with uniform repartition of centers
	L := len(points)
	step := L / K
	centers := make([]Pos, K)
	for i := 0; i < K; i++ {
		centers[i] = points[step/2+i*step]
	}
	cls := clusters{
		pointClusters: make([]uint8, L), // will be erased by the first assignClusters call
		clusterSize:   make([]int, K),   // idem
	}

	for hasChanged := true; hasChanged; {
		hasChanged = assignClusters(points, centers, cls)
		computeCenters(points, cls, centers)
	}

	return kmOut{points, cls, centers}
}

type clusters struct {
	pointClusters []uint8 // index of the cluster for point i
	clusterSize   []int   // with len K
}

// update [out], selecting for each point the closest center
//
// len(points) == len(out)
func assignClusters(points []Pos, centers []Pos, out clusters) (hasChanged bool) {
	// reset cluster sizes
	for i := range out.clusterSize {
		out.clusterSize[i] = 0
	}

	for i, p := range points {
		bestD := inf
		bestCluster := -1
		for j, center := range centers {
			if dist := distP(p, center); dist < bestD {
				bestD = dist
				bestCluster = j
			}
		}
		cl := uint8(bestCluster)

		if out.pointClusters[i] != cl {
			hasChanged = true
		}

		out.pointClusters[i] = cl
		out.clusterSize[cl]++
	}

	return hasChanged
}

// update [out] by averaring the points in cluster
// empty clusters are ignored
//
// len(clusters) == len(out)
func computeCenters(points []Pos, clusters clusters, out []Pos) {
	// reset out
	for i := range out {
		out[i] = Pos{}
	}

	for i, p := range points {
		cl := clusters.pointClusters[i]
		out[cl].X += p.X
		out[cl].Y += p.Y
	}
	// divide by the number of cluster
	for i := range out {
		if s := fl(clusters.clusterSize[i]); s != 0 {
			out[i].X /= s
			out[i].Y /= s
		}
	}
}

// Within-Cluster-Sum-of-Squares (WCSS)
func (km kmOut) wcss() fl {
	var s fl
	for i, p := range km.points {
		cl := km.cls.pointClusters[i]
		s += distP(p, km.centers[cl])
	}
	s /= fl(len(km.points)) // to ease comparison
	return s
}

// return true if at least one class has less than 3 points
func (cls clusters) isDegenerated() bool {
	for _, size := range cls.clusterSize {
		if size <= 3 {
			return true
		}
	}
	return false
}

func detectOutlierInCluster(points []Pos, cls clusters, center Pos, cl uint8) []int {
	// average distance to center in cluster
	Nk := cls.clusterSize[cl]
	if Nk == 0 {
		return nil
	}

	inCluster := make([]int, 0, Nk)
	distances := make([]fl, 0, Nk)

	for i, p := range points {
		if cls.pointClusters[i] == cl {
			inCluster = append(inCluster, i)
			distances = append(distances, distP(p, center))
		}
	}
	mean, std := meanAndStd(distances)
	thresholdMax := mean + 1*std

	var out []int
	for j, pointIndex := range inCluster {
		d := distances[j]
		if d > thresholdMax { // found outlier
			out = append(out, pointIndex)
		}
	}
	return out
}

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

func (clustered kmOut) detectOutliers() map[int]bool {
	out := map[int]bool{}
	for k := range clustered.cls.clusterSize {
		for _, outlier := range detectOutlierInCluster(clustered.points, clustered.cls, clustered.centers[k], uint8(k)) {
			out[outlier] = true
		}
	}
	return out
}

// complete algorithm

type clusterRange [2]int // start, end (excluded, in slice syntax)

// segmentation filters outliers and segments the resulting
// points
// outlier are not present in the returned ranges
func segmentation(angles []Pos) []clusterRange {
	const KMax = 5

	// run kmeans for each K and pick the best WCSS to detect outliers
	bestWCSSK, bestWCSS := kmOut{}, inf
	for K := 1; K <= KMax; K++ {
		kmeansOut := kmeans(angles, K)

		if kmeansOut.cls.isDegenerated() {
			continue
		}

		if w := kmeansOut.wcss(); w < bestWCSS {
			bestWCSS = w
			bestWCSSK = kmeansOut
		}
	}

	outliers := bestWCSSK.detectOutliers()
	clusters := segmentByAngleBreak(angles, outliers)
	return clusters
}

func segmentByAngleBreak(angles []Pos, outliers map[int]bool) []clusterRange {
	const angleBreak = 80

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
			previous = current.Y
			continue
		}
		if abs(current.Y-previous) >= angleBreak { // new cluster, push the previous
			push(clusterRange{currentRangeStart, previousIndex + 1})
			currentRangeStart = i
		}
		previous = current.Y
		previousIndex = i
	}
	if currentRangeStart < len(angles)-1 {
		push(clusterRange{currentRangeStart, len(angles)})
	}

	return out
}
