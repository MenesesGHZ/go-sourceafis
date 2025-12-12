package transparency

import "github.com/menesesghz/go-sourceafis/primitives"

type ConsistentSkeletonRidge struct {
	Start, End int
	Points     primitives.List[primitives.IntPoint]
}
