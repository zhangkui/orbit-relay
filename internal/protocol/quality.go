package protocol

import "gitlab.com/zhangkui/orbit-relay/internal/model"

func LinkQuality(signal, temperature, voltage float64) string {
	if signal < -15 || temperature < -30 || temperature > 70 || voltage < 3 {
		return "bad"
	}
	if signal < -8 || temperature < -10 || temperature > 55 || voltage < 3.3 {
		return "degraded"
	}
	return "good"
}
func ApplyQuality(frame *model.TelemetryFrame) {
	frame.Quality = LinkQuality(frame.SignalDB, frame.Temperature, frame.Voltage)
}
