package matcher

import (
	"container/heap"
	"context"
)

type PriorityQueue []*MinutiaPair

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].distance < pq[j].distance
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*MinutiaPair)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}

type matcherThreadContextKey struct{}

type MatcherThread struct {
	Roots   *RootList
	Pairing *PairingGraph
	Queue   PriorityQueue
	Score   *ScoringData
}

// CurrentThread returns the scratch state bound to ctx by WithThread, or a fresh
// one when there is none. A fresh one is discarded after the call, so a caller
// matching many templates should bind one per goroutine with WithThread.
func CurrentThread(ctx context.Context) *MatcherThread {
	if thread, ok := ctx.Value(matcherThreadContextKey{}).(*MatcherThread); ok {
		return thread
	}
	return NewMatcherThread()
}

// WithThread binds a reusable MatcherThread to ctx. The thread is not safe for
// concurrent use: bind one per goroutine.
func WithThread(ctx context.Context) context.Context {
	return context.WithValue(ctx, matcherThreadContextKey{}, NewMatcherThread())
}

func NewMatcherThread() *MatcherThread {
	pool := NewMinutiaPairPool()
	thread := &MatcherThread{
		Roots:   NewRootList(pool),
		Pairing: NewPairingGraph(pool),
		Queue:   make(PriorityQueue, 0),
		Score:   new(ScoringData),
	}
	heap.Init(&thread.Queue)
	return thread
}
