package telemetry

import (
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/protocol"
)

func Parse(raw []byte, satellite, station string) (model.TelemetryFrame, error) {
	f, err := protocol.DecodeFrame(raw, satellite, station)
	if err != nil {
		return f, err
	}
	if err := protocol.ValidateFrame(f); err != nil {
		return f, fmt.Errorf("frame rejected: %w", err)
	}
	return f, nil
}
func DecodeMetrics(frame *model.TelemetryFrame) {
	if len(frame.Payload) >= 12 {
		frame.SignalDB = float64(int16(frame.Payload[0])<<8|int16(frame.Payload[1])) / 10
		frame.Temperature = float64(int16(frame.Payload[2])<<8|int16(frame.Payload[3])) / 10
		frame.Voltage = float64(frame.Payload[4]) / 10
	}
	protocol.ApplyQuality(frame)
}
