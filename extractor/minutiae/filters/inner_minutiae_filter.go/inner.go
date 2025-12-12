package inner

import (
	"github.com/menesesghz/go-sourceafis/config"
	"github.com/menesesghz/go-sourceafis/features"
	"github.com/menesesghz/go-sourceafis/primitives"
)

func Apply(minutiae *primitives.GenericList[*features.FeatureMinutia], mask *primitives.BooleanMatrix) {
	for e := minutiae.Front(); e != nil; {
		minutia := e.Value.(*features.FeatureMinutia)

		arrow := primitives.FloatAngle(minutia.Direction).ToVector().Multiply(-config.Config.MaskDisplacement).Round()

		if !mask.GetPointWithFallback(minutia.Position.Plus(arrow), false) {
			next := e.Next()
			minutiae.Remove(e)
			e = next
		} else {
			e = e.Next()
		}
	}

}
