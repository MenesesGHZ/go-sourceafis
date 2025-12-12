package templates

import (
	"bytes"
	"encoding/gob"
	"os"

	"github.com/menesesghz/go-sourceafis/features"
	"github.com/menesesghz/go-sourceafis/primitives"
)

func init() {
	// Register types used by SearchTemplate so gob can encode/decode them reliably
	gob.Register(&features.SearchMinutia{})
	gob.Register(&features.NeighborEdge{})
	gob.Register(&features.EdgeShape{})
	gob.Register(features.MinutiaType(0))
	gob.Register(primitives.FloatAngle(0))
}

// MarshalBinary implements encoding.BinaryMarshaler for SearchTemplate
func (t *SearchTemplate) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	// encode fields into a plain value (not the method-bearing type) to avoid
	// re-entering MarshalBinary via gob's use of encoding interfaces.
	payload := struct {
		Width    int
		Height   int
		Minutiae []*features.SearchMinutia
		Edges    [][]*features.NeighborEdge
	}{
		Width:    t.Width,
		Height:   t.Height,
		Minutiae: t.Minutiae,
		Edges:    t.Edges,
	}
	if err := enc.Encode(&payload); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for SearchTemplate
func (t *SearchTemplate) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	payload := struct {
		Width    int
		Height   int
		Minutiae []*features.SearchMinutia
		Edges    [][]*features.NeighborEdge
	}{}
	if err := dec.Decode(&payload); err != nil {
		return err
	}
	t.Width = payload.Width
	t.Height = payload.Height
	t.Minutiae = payload.Minutiae
	t.Edges = payload.Edges
	return nil
}

// SaveSearchTemplateFile saves a SearchTemplate to a file using gob encoding.
func SaveSearchTemplateFile(path string, tmpl *SearchTemplate) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(tmpl)
}

// LoadSearchTemplateFile loads a SearchTemplate from a file previously saved with SaveSearchTemplateFile.
func LoadSearchTemplateFile(path string) (*SearchTemplate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	var t SearchTemplate
	if err := dec.Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}
