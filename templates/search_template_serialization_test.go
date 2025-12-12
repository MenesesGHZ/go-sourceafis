package templates

import (
	"os"
	"testing"

	"github.com/menesesghz/go-sourceafis/features"
)

func TestSearchTemplateSerializationRoundtrip(t *testing.T) {
	// build a minimal SearchTemplate with two minutiae and a neighbor edge
	sm1 := &features.SearchMinutia{X: 10, Y: 20, Direction: 0.5, T: features.ENDING}
	sm2 := &features.SearchMinutia{X: 30, Y: 45, Direction: 1.2, T: features.BIFURCATION}
	minutiae := []*features.SearchMinutia{sm1, sm2}

	// Build edges manually: neighbor edges referencing indices
	ne := features.NewNeighborEdge(minutiae, 0, 1)
	edges := make([][]*features.NeighborEdge, 2)
	edges[0] = []*features.NeighborEdge{ne}
	edges[1] = []*features.NeighborEdge{}

	orig := &SearchTemplate{
		Width:    100,
		Height:   120,
		Minutiae: minutiae,
		Edges:    edges,
	}

	tmpfile := "test_search_template.gob"
	if err := SaveSearchTemplateFile(tmpfile, orig); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	defer os.Remove(tmpfile)

	loaded, err := LoadSearchTemplateFile(tmpfile)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Width != orig.Width || loaded.Height != orig.Height {
		t.Fatalf("dimensions mismatch: got %dx%d want %dx%d", loaded.Width, loaded.Height, orig.Width, orig.Height)
	}

	if len(loaded.Minutiae) != len(orig.Minutiae) {
		t.Fatalf("minutiae count mismatch: got %d want %d", len(loaded.Minutiae), len(orig.Minutiae))
	}

	if len(loaded.Edges) != len(orig.Edges) {
		t.Fatalf("edges length mismatch: got %d want %d", len(loaded.Edges), len(orig.Edges))
	}

	// basic content checks
	if loaded.Minutiae[0].X != orig.Minutiae[0].X || loaded.Minutiae[1].Y != orig.Minutiae[1].Y {
		t.Fatalf("minutiae content mismatch")
	}

	if len(loaded.Edges[0]) != 1 || loaded.Edges[0][0].Neighbor != orig.Edges[0][0].Neighbor {
		t.Fatalf("edge content mismatch")
	}
}
