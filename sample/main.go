package main

import (
	"context"
	"runtime"
	"time"

	sourceafis "github.com/miqdadyyy/go-sourceafis"
	"github.com/miqdadyyy/go-sourceafis/config"
	"github.com/miqdadyyy/go-sourceafis/templates"

	"log"
)

func main() {
	config.LoadDefaultConfig()
	config.Config.Workers = runtime.NumCPU()
	l := sourceafis.NewTransparencyLogger(new(sourceafis.DefaultTransparency))
	tc := sourceafis.NewTemplateCreator(l)

	geras, err := LoadImageTemplates(tc, "gera/gera1.png", "gera/gera2.png", "gera/gera3.png", "gera/gera4.png", "gera/gera5.pgm")
	if err != nil {
		log.Fatal(err.Error())
	}
	richis, err := LoadImageTemplates(tc, "richi/richi1.png", "richi/richi2.png", "richi/richi3.png", "richi/richi4.png")
	if err != nil {
		log.Fatal(err.Error())
	}

	matcher := sourceafis.NewMatcher(l)

	matcher.Update(1, geras[:3]...) // 1, 2, 3
	matcher.Update(2, richis[:3]...)

	ctx := context.Background()
	now := time.Now()
	id, err := matcher.FindMatch(ctx, geras[4]) // 4
	elapsed := time.Since(now)

	if err != nil {
		log.Fatalf("No match found: %s", err.Error())
	} else {
		log.Printf("Match found: ID=%d (in %s)", id, elapsed)
	}
}

func LoadImageTemplates(tc *sourceafis.TemplateCreator, paths ...string) ([]*templates.SearchTemplate, error) {
	tmpls := make([]*templates.SearchTemplate, 0, len(paths))
	for _, path := range paths {
		img, err := sourceafis.LoadImage(path)
		if err != nil {
			return nil, err
		}
		tmpl, err := tc.Template(img)
		if err != nil {
			return nil, err
		}
		tmpls = append(tmpls, tmpl)
	}
	return tmpls, nil
}
