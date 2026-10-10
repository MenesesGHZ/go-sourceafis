package matcher

import (
	"github.com/menesesghz/go-sourceafis/config"
)

type RootList struct {
	pool       *MinutiaPairPool
	count      int
	pairs      []*MinutiaPair
	duplicates map[int]struct{}
}

func NewRootList(pool *MinutiaPairPool) *RootList {
	return &RootList{
		pool:       pool,
		pairs:      make([]*MinutiaPair, config.Config.MaxTriedRoots),
		duplicates: make(map[int]struct{}),
	}
}

func (r *RootList) Add(pair *MinutiaPair) {
	r.pairs[r.count] = pair
	r.count++
}

// MarkSeen records a root by key and reports whether it was new.
func (r *RootList) MarkSeen(key int) bool {
	if _, seen := r.duplicates[key]; seen {
		return false
	}
	r.duplicates[key] = struct{}{}
	return true
}

func (r *RootList) Discard() {
	for i := 0; i < r.count; i++ {
		r.pool.Release(r.pairs[i])
		r.pairs[i] = nil
	}
	r.count = 0
	clear(r.duplicates)
}
