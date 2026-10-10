package matcher

type MinutiaPairPool struct {
	pool   []*MinutiaPair
	pooled int
}

func NewMinutiaPairPool() *MinutiaPairPool {
	return &MinutiaPairPool{
		pool: make([]*MinutiaPair, 1),
	}
}

func (p *MinutiaPairPool) Allocate() *MinutiaPair {
	if p.pooled > 0 {
		p.pooled--
		pair := p.pool[p.pooled]
		p.pool[p.pooled] = nil
		return pair
	} else {
		return new(MinutiaPair)
	}
}

// Release returns a pair to the pool. Every allocated pair has exactly one owner
// (the root list, the pairing tree or its support list) that releases it once.
func (p *MinutiaPairPool) Release(pair *MinutiaPair) {
	if p.pooled >= len(p.pool) {
		grown := make([]*MinutiaPair, 2*len(p.pool))
		copy(grown, p.pool)
		p.pool = grown
	}
	pair.Probe = 0
	pair.Candidate = 0
	pair.ProbeRef = 0
	pair.CandidateRef = 0
	pair.distance = 0
	pair.supportingEdges = 0
	p.pool[p.pooled] = pair
	p.pooled++
}
