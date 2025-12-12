package fragment

import (
	"github.com/menesesghz/go-sourceafis/config"
	"github.com/menesesghz/go-sourceafis/extractor/logger"
	"github.com/menesesghz/go-sourceafis/extractor/skeletons/filters/dot"
	"github.com/menesesghz/go-sourceafis/features"
)

type SkeletonFragmentFilter struct {
	logger logger.TransparencyLogger
}

func New(logger logger.TransparencyLogger) *SkeletonFragmentFilter {
	return &SkeletonFragmentFilter{
		logger: logger,
	}
}

func (f *SkeletonFragmentFilter) Apply(skeleton *features.Skeleton) error {
	for _, minutia := range skeleton.Minutiae {
		if len(minutia.Ridges) == 1 {
			ridge := minutia.Ridges[0]
			if len(ridge.End().Ridges) == 1 && ridge.Points.Size() < config.Config.MinFragmentLength {
				ridge.Detach()
			}
		}
	}

	if err := dot.Apply(skeleton); err != nil {
		return err
	}

	return f.logger.LogSkeleton("removed-fragments", skeleton)
}
