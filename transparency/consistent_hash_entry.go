package transparency

import "github.com/menesesghz/go-sourceafis/features"

type ConsistentHashEntry struct {
	Key   int
	Edges []*features.IndexedEdge
}
