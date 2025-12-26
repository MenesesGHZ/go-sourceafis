package sourceafis

import (
	"context"
	"errors"
	"runtime"

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

type Matcher struct {
	fingerprintMap map[string]map[int][]*templates.SearchTemplate
	logger         matcher.MatcherLogger
	matcher        *matcher.Matcher
}

func NewMatcher(logger matcher.MatcherLogger) *Matcher {
	return &Matcher{
		matcher:        matcher.NewMatcher(logger),
		fingerprintMap: make(map[string]map[int][]*templates.SearchTemplate),
		logger:         logger,
	}
}

func (m *Matcher) FindMatch(ctx context.Context, namespace string, candidate *templates.SearchTemplate) (int, float64, error) {
	templatesMap, exists := m.fingerprintMap[namespace]
	if !exists || len(templatesMap) == 0 {
		return 0, 0, ErrNoMatchFound
	}

	// Build hash
	hashBuilder := matcher.NewEdgeHashBuilder(m.logger.(matcher.HashTableLogger))
	hash, err := hashBuilder.Build(candidate)
	if err != nil {
		return 0, 0, err
	}
	candidateProbe := matcher.NewProbe(candidate, hash)

	type job struct {
		id        int
		templates []*templates.SearchTemplate
	}

	type result struct {
		id    int
		score float64
	}

	workerCount := runtime.NumCPU()
	jobs := make(chan job)
	results := make(chan result)

	// Worker logic
	worker := func() {
		for j := range jobs {
			localMax := -1.0
			for _, tmpl := range j.templates {
				select {
				case <-ctx.Done():
					return
				default:
				}

				score := m.matcher.Match(ctx, candidateProbe, tmpl)
				if score > localMax {
					localMax = score
				}
			}
			results <- result{id: j.id, score: localMax}
		}
	}

	// Start workers
	for i := 0; i < workerCount; i++ {
		go worker()
	}

	// Feed jobs
	go func() {
		for id, list := range m.fingerprintMap[namespace] {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- job{id: id, templates: list}:
			}
		}
		close(jobs)
	}()

	// Collect results
	maxID := 0
	maxScore := -1.0
	for i := 0; i < len(m.fingerprintMap[namespace]); i++ {
		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		case r := <-results:
			if r.score > maxScore {
				maxScore = r.score
				maxID = r.id
			}
		}
	}

	if maxScore < DefaultMatchThreshold {
		return 0, maxScore, ErrNoMatchFound
	}

	return maxID, maxScore, nil
}

func (m *Matcher) Update(id int, namespace string, ts ...*templates.SearchTemplate) {
	if len(ts) == 0 {
		delete(m.fingerprintMap[namespace], id)
	} else {
		if _, exists := m.fingerprintMap[namespace]; !exists {
			m.fingerprintMap[namespace] = make(map[int][]*templates.SearchTemplate)
		}
		m.fingerprintMap[namespace][id] = ts
	}
}
