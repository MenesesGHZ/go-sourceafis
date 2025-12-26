package main

import (
	"context"
	"log"
	"runtime"
	"time"

	sourceafis "github.com/menesesghz/go-sourceafis"
	"github.com/menesesghz/go-sourceafis/config"
	"github.com/menesesghz/go-sourceafis/templates"
)

func main() {
	config.LoadDefaultConfig()
	config.Config.Workers = runtime.NumCPU()
	l := sourceafis.NewTransparencyLogger(new(sourceafis.DefaultTransparency))
	tc := sourceafis.NewTemplateCreator(l)

	geras, err := LoadImageTemplates(tc, "gera/gera1.png", "gera/gera2.png", "gera/gera3.png", "gera/gera4.png")
	if err != nil {
		log.Fatal(err.Error())
	}
	richis, err := LoadImageTemplates(tc, "richi/richi1.png", "richi/richi2.png", "richi/richi3.png", "richi/richi4.png")
	if err != nil {
		log.Fatal(err.Error())
	}

	matcher := sourceafis.NewMatcher(l)

	const namespace = "group-1"
	matcher.Update(1, namespace, geras[:3]...) // 1, 2, 3
	matcher.Update(2, namespace, richis[:3]...)

	ctx := context.Background()
	now := time.Now()
	id, score, err := matcher.FindMatch(ctx, namespace, geras[3]) // 4
	elapsed := time.Since(now)
	if err != nil {
		log.Fatalf("No match found: %s", err.Error())
	} else {
		log.Printf("Match found: ID=%d, Score=%.2f (in %s)", id, score, elapsed)
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
