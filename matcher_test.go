package sourceafis

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/menesesghz/go-sourceafis/config"
	"github.com/menesesghz/go-sourceafis/matcher"
	"github.com/menesesghz/go-sourceafis/templates"
)

var samplePaths = []string{
	"sample/gera/gera1.png", "sample/gera/gera2.png", "sample/gera/gera3.png", "sample/gera/gera4.png",
	"sample/richi/richi1.png", "sample/richi/richi2.png", "sample/richi/richi3.png", "sample/richi/richi4.png",
}

// goldenScores[i][j] is the score of probe i against a gallery holding only
// template j, recorded with 24d9195 — before the pair pool reused pairs,
// duplicate roots were skipped and each worker kept one MatcherThread. Those
// changes must be invisible in the scores.
var goldenScores = [8][8]float64{
	{711.6160707312994, 170.77824396331837, 0.13595108643080378, 74.08925776548662, 2.172672961723243, 0, 0.5162676674380642, 1.9114433137064641},
	{170.8005848507298, 653.3238299771295, 51.5055034957609, 30.878835037456774, 1.0749861717509552, 0, 0, 4.088597744115526},
	{23.929385040537714, 41.487940773561085, 583.3731410721259, 14.031842469055913, 0.5697570183759324, 2.2479084476416777, 0, 0.2822061396356587},
	{72.45245858839473, 30.428673869372638, 0.9972011421738762, 1084.6864115579854, 3.0422250774800923, 2.1737058103734777, 0, 1.9373274579662112},
	{0.7058667603906128, 1.0749861717509575, 0, 0, 420.15486696045076, 148.8120123925938, 175.00768900625314, 207.5701317613438},
	{0.15413831664960018, 1.4326853400619877, 2.2387024403107016, 2.5364989405735314, 148.8120123925938, 361.8626262062811, 106.8614817727551, 143.09289283323557},
	{0, 0, 0.03762635827682307, 0.8191779570084852, 175.00768900625314, 107.03550530488114, 385.17952250794895, 137.1990798957508},
	{1.9114433137064641, 4.088597744115526, 0.2822061396356587, 1.9373274579662112, 207.53565746985805, 138.3683960693626, 137.1990798957508, 385.17952250794895},
}

var (
	sampleOnce      sync.Once
	sampleTemplates []*templates.SearchTemplate
	sampleErr       error
)

func loadSamples(t testing.TB) []*templates.SearchTemplate {
	t.Helper()
	sampleOnce.Do(func() {
		config.LoadDefaultConfig()
		tc := NewTemplateCreator(NoopTransparencyLogger{})
		for _, p := range samplePaths {
			img, err := LoadImage(p)
			if err != nil {
				sampleErr = err
				return
			}
			st, err := tc.Template(img)
			if err != nil {
				sampleErr = err
				return
			}
			sampleTemplates = append(sampleTemplates, st)
		}
	})
	if sampleErr != nil {
		t.Fatal(sampleErr)
	}
	return sampleTemplates
}

func TestFindMatchScoresUnchanged(t *testing.T) {
	tm := loadSamples(t)
	loggers := map[string]interface {
		matcher.MatcherLogger
		matcher.HashTableLogger
	}{
		"noop":    NoopTransparencyLogger{},
		"default": NewTransparencyLogger(new(DefaultTransparency)),
	}
	for name, l := range loggers {
		for i := range tm {
			for j := range tm {
				m := NewMatcher(l, 30)
				m.Update(1, "x", tm[j])
				_, score, _ := m.FindMatch(context.Background(), "x", tm[i])
				if score != goldenScores[i][j] {
					t.Errorf("%s: probe %d vs template %d: score %v, want %v", name, i, j, score, goldenScores[i][j])
				}
			}
		}
	}
}

func TestFindMatchPicksBestMemberInGallery(t *testing.T) {
	tm := loadSamples(t)
	m := NewMatcher(NoopTransparencyLogger{}, 30)
	for id := 1; id <= 200; id++ {
		m.Update(id, "t", tm[4], tm[5], tm[6]) // richi
	}
	m.Update(137, "t", tm[0], tm[1], tm[2]) // gera
	m.Update(1, "other", tm[0], tm[1], tm[2])

	id, score, err := m.FindMatch(context.Background(), "t", tm[3])
	if err != nil || id != 137 {
		t.Fatalf("FindMatch = (%d, %v, %v), want member 137", id, score, err)
	}
	// Best of gera1..3 against gera4.
	if want := goldenScores[3][0]; score != want {
		t.Fatalf("score %v, want %v", score, want)
	}

	if _, _, err := m.FindMatch(context.Background(), "empty", tm[3]); !errors.Is(err, ErrNoMatchFound) {
		t.Fatalf("empty namespace: err %v, want ErrNoMatchFound", err)
	}
	m.Update(137, "t") // delete
	if _, _, err := m.FindMatch(context.Background(), "t", tm[3]); !errors.Is(err, ErrNoMatchFound) {
		t.Fatalf("after delete: err %v, want ErrNoMatchFound", err)
	}
}

// Run with -race: enrolling while a search runs used to be an unguarded map
// read/write, which Go aborts the whole process on.
func TestFindMatchConcurrentWithUpdate(t *testing.T) {
	tm := loadSamples(t)
	m := NewMatcher(NoopTransparencyLogger{}, 30)
	for id := 1; id <= 50; id++ {
		m.Update(id, "t", tm[4], tm[5], tm[6])
	}
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for id := 0; ; id++ {
			select {
			case <-stop:
				return
			default:
			}
			m.Update(1000+id%20, "t", tm[0], tm[1])
			m.Update(1000+(id+10)%20, "t")
		}
	}()
	for i := 0; i < 20; i++ {
		if _, _, err := m.FindMatch(context.Background(), "t", tm[3]); err != nil && !errors.Is(err, ErrNoMatchFound) {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
}

func TestFindMatchCancelledLeaksNothing(t *testing.T) {
	tm := loadSamples(t)
	m := NewMatcher(NoopTransparencyLogger{}, 30)
	for id := 1; id <= 100; id++ {
		m.Update(id, "t", tm[4], tm[5], tm[6])
	}
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := m.FindMatch(ctx, "t", tm[3]); !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v, want context.Canceled", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	m.FindMatch(ctx, "t", tm[3])

	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("%d goroutines after FindMatch returned, %d before", n, before)
	}
}

func BenchmarkFindMatch600Members(b *testing.B) {
	tm := loadSamples(b)
	m := NewMatcher(NoopTransparencyLogger{}, 30)
	for id := 1; id <= 600; id++ {
		m.Update(id, "t", tm[4], tm[5], tm[6])
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.FindMatch(context.Background(), "t", tm[3])
	}
}

func BenchmarkTemplate(b *testing.B) {
	loadSamples(b)
	img, err := LoadImage(samplePaths[0])
	if err != nil {
		b.Fatal(err)
	}
	tc := NewTemplateCreator(NoopTransparencyLogger{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tc.Template(img); err != nil {
			b.Fatal(err)
		}
	}
}
