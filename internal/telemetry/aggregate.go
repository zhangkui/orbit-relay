package telemetry

import (
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"time"
)

type Summary struct {
	SatelliteID        string
	From               time.Time
	To                 time.Time
	Count              int
	Good               int
	Degraded           int
	Bad                int
	MinSignal          float64
	MaxSignal          float64
	AverageTemperature float64
	AverageVoltage     float64
	MissingSequences   int
}

func Summarize(frames []model.TelemetryFrame) Summary {
	if len(frames) == 0 {
		return Summary{}
	}
	sorted := append([]model.TelemetryFrame(nil), frames...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ReceivedAt.Before(sorted[j].ReceivedAt) })
	s := Summary{SatelliteID: sorted[0].SatelliteID, From: sorted[0].ReceivedAt, To: sorted[len(sorted)-1].ReceivedAt, Count: len(sorted), MinSignal: sorted[0].SignalDB, MaxSignal: sorted[0].SignalDB}
	var temp, volt float64
	for _, f := range sorted {
		if f.SignalDB < s.MinSignal {
			s.MinSignal = f.SignalDB
		}
		if f.SignalDB > s.MaxSignal {
			s.MaxSignal = f.SignalDB
		}
		temp += f.Temperature
		volt += f.Voltage
		switch f.Quality {
		case "good":
			s.Good++
		case "degraded":
			s.Degraded++
		case "bad":
			s.Bad++
		}
	}
	s.AverageTemperature = temp / float64(len(sorted))
	s.AverageVoltage = volt / float64(len(sorted))
	s.MissingSequences = SequenceGaps(sorted)
	return s
}
func Bucket(frames []model.TelemetryFrame, step time.Duration) map[int64][]model.TelemetryFrame {
	out := map[int64][]model.TelemetryFrame{}
	if step <= 0 {
		return out
	}
	for _, f := range frames {
		k := f.ReceivedAt.Unix() / int64(step/time.Second)
		out[k] = append(out[k], f)
	}
	return out
}
