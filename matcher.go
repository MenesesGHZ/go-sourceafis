package sourceafis

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/menesesghz/go-sourceafis/matcher"
	"github.com/menesesghz/go-sourceafis/templates"
)

var (
	ErrNoMatchFound = errors.New("no match found")
)

const (
	// DefaultMatchThreshold is the default threshold for fingerprint matching, any score above this value is considered a match.
	DefaultMatchThreshold = 25.0
)

// Matcher is safe for concurrent use: FindMatch may run while Update changes
// the gallery.
type Matcher struct {
	mu             sync.RWMutex
	fingerprintMap map[string]map[int][]*templates.SearchTemplate
	logger         matcher.MatcherLogger
	matcher        *matcher.Matcher
	scoreThreshold float64
}

func NewMatcher(logger matcher.MatcherLogger, scoreThreshold float64) *Matcher {
	if scoreThreshold <= 0 {
		scoreThreshold = DefaultMatchThreshold
	}
	return &Matcher{
		matcher:        matcher.NewMatcher(logger),
		fingerprintMap: make(map[string]map[int][]*templates.SearchTemplate),
		logger:         logger,
		scoreThreshold: scoreThreshold,
	}
}

type galleryEntry struct {
	id        int
	templates []*templates.SearchTemplate
}

// snapshot copies the namespace's entries under the read lock, so a search
// never holds the lock while it matches and an Update never waits for one.
// Templates are never mutated after Update, so sharing them is safe.
func (m *Matcher) snapshot(namespace string) []galleryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entries := make([]galleryEntry, 0, len(m.fingerprintMap[namespace]))
	for id, list := range m.fingerprintMap[namespace] {
		entries = append(entries, galleryEntry{id: id, templates: list})
	}
	return entries
}

func (m *Matcher) FindMatch(ctx context.Context, namespace string, candidate *templates.SearchTemplate) (int, float64, error) {
	entries := m.snapshot(namespace)
	if len(entries) == 0 {
		return 0, 0, ErrNoMatchFound
	}

	// Build hash
	hashBuilder := matcher.NewEdgeHashBuilder(m.logger.(matcher.HashTableLogger))
	hash, err := hashBuilder.Build(candidate)
	if err != nil {
		return 0, 0, err
	}
	candidateProbe := matcher.NewProbe(candidate, hash)

	type result struct {
		id    int
		score float64
	}

	workerCount := min(runtime.NumCPU(), len(entries))
	// Buffered so a worker never blocks on a caller that already returned.
	results := make(chan result, workerCount)
	var next atomic.Int64

	// Each worker claims entries by index and reports only its own best, so the
	// per-member cost is one atomic add instead of two channel handoffs.
	worker := func() {
		ctx := matcher.WithThread(ctx)
		best := result{score: -1}
		for {
			i := int(next.Add(1)) - 1
			if i >= len(entries) || ctx.Err() != nil {
				break
			}
			e := entries[i]
			for _, tmpl := range e.templates {
				if score := m.matcher.Match(ctx, candidateProbe, tmpl); score > best.score {
					best = result{id: e.id, score: score}
				}
			}
		}
		results <- best
	}

	for i := 0; i < workerCount; i++ {
		go worker()
	}

	maxID := 0
	maxScore := -1.0
	for i := 0; i < workerCount; i++ {
		r := <-results
		// Ties go to the lower id, so the outcome does not depend on scheduling.
		if r.score > maxScore || (r.score == maxScore && r.id < maxID) {
			maxScore = r.score
			maxID = r.id
		}
	}
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}

	if maxScore < m.scoreThreshold {
		return 0, maxScore, ErrNoMatchFound
	}

	return maxID, maxScore, nil
}

func (m *Matcher) Update(id int, namespace string, ts ...*templates.SearchTemplate) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(ts) == 0 {
		delete(m.fingerprintMap[namespace], id)
	} else {
		if _, exists := m.fingerprintMap[namespace]; !exists {
			m.fingerprintMap[namespace] = make(map[int][]*templates.SearchTemplate)
		}
		m.fingerprintMap[namespace][id] = ts
	}
}
