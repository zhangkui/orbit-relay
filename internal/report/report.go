package report

import (
	"context"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/telemetry"
	"sort"
	"time"
)

type Source interface {
	List(context.Context, string) ([]model.TelemetryFrame, error)
}
type Builder struct {
	source Source
	now    func() time.Time
}

func New(s Source) *Builder {
	return &Builder{source: s, now: func() time.Time { return time.Now().UTC() }}
}
func (b *Builder) Build(ctx context.Context, w model.ContactWindow, commands int) (model.LinkReport, error) {
	frames, err := b.source.List(ctx, w.ID)
	if err != nil {
		return model.LinkReport{}, err
	}
	r := model.LinkReport{WindowID: w.ID, Frames: len(frames), Commands: commands, AverageSignal: telemetry.AverageSignal(frames), PacketLoss: 0, GeneratedAt: b.now()}
	if len(frames) > 0 {
		r.TemperatureMin = frames[0].Temperature
		r.TemperatureMax = frames[0].Temperature
	}
	for _, f := range frames {
		if f.Temperature < r.TemperatureMin {
			r.TemperatureMin = f.Temperature
		}
		if f.Temperature > r.TemperatureMax {
			r.TemperatureMax = f.Temperature
		}
	}
	if len(frames) > 0 {
		r.PacketLoss = float64(telemetry.SequenceGaps(frames)) / float64(len(frames)+telemetry.SequenceGaps(frames))
	}
	return r, nil
}
func SortBySignal(frames []model.TelemetryFrame) []model.TelemetryFrame {
	out := append([]model.TelemetryFrame(nil), frames...)
	sort.Slice(out, func(i, j int) bool { return out[i].SignalDB > out[j].SignalDB })
	return out
}
