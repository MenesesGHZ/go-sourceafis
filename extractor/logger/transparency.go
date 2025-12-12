package logger

import (
	"github.com/menesesghz/go-sourceafis/features"
)

type TransparencyLogger interface {
	Log(key string, data interface{}) error
	LogSkeleton(keyword string, skeleton *features.Skeleton) error
}
