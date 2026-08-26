package telemetry

import (
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
)

func AverageSignal(frames []model.TelemetryFrame) float64 {
	if len(frames) == 0 {
		return 0
	}
	var sum float64
	for _, f := range frames {
		sum += f.SignalDB
	}
	return sum / float64(len(frames))
}
func SequenceGaps(frames []model.TelemetryFrame) int {
	if len(frames) < 2 {
		return 0
	}
	copyFrames := frames
	sort.Slice(copyFrames, func(i, j int) bool { return copyFrames[i].Sequence < copyFrames[j].Sequence })
	gaps := 0
	for i := 1; i < len(copyFrames); i++ {
		if copyFrames[i].Sequence > copyFrames[i-1].Sequence+1 {
			gaps += int(copyFrames[i].Sequence - copyFrames[i-1].Sequence - 1)
		}
	}
	return gaps
}
