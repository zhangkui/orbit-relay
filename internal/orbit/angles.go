package orbit

import (
	"math"
	"sort"
)

func NormalizeDegrees(v float64) float64 {
	v = math.Mod(v, 360)
	if v < 0 {
		v += 360
	}
	return v
}
func AngularDistance(a, b float64) float64 {
	d := math.Abs(NormalizeDegrees(a) - NormalizeDegrees(b))
	if d > 180 {
		d = 360 - d
	}
	return d
}
func InterpolateAngle(a, b, f float64) float64 {
	d := NormalizeDegrees(b - a)
	if d > 180 {
		d -= 360
	}
	return NormalizeDegrees(a + d*f)
}
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	m := len(v) / 2
	if len(v)%2 == 1 {
		return v[m]
	}
	return (v[m-1] + v[m]) / 2
}
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var s float64
	for _, v := range values {
		s += v
	}
	return s / float64(len(values))
}
func StdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := Mean(values)
	var s float64
	for _, v := range values {
		d := v - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(values)-1))
}
func Clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
func Bearing(a, b Position) float64 {
	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	dlon := (b.Longitude - a.Longitude) * math.Pi / 180
	y := math.Sin(dlon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dlon)
	return NormalizeDegrees(math.Atan2(y, x) * 180 / math.Pi)
}
