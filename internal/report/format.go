package report

import (
	"encoding/json"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"strings"
	"time"
)

func JSON(r model.LinkReport) (string, error) {
	b, e := json.MarshalIndent(r, "", "  ")
	return string(b), e
}
func Text(r model.LinkReport) string {
	return fmt.Sprintf("窗口 %s\n遥测帧: %d\n指令: %d\n平均信号: %.1f dB\n丢包率: %.2f%%\n温度范围: %.1f..%.1f C\n生成时间: %s", r.WindowID, r.Frames, r.Commands, r.AverageSignal, r.PacketLoss*100, r.TemperatureMin, r.TemperatureMax, r.GeneratedAt.Format(time.RFC3339))
}
func CSV(reports []model.LinkReport) string {
	lines := []string{"window_id,frames,commands,average_signal,packet_loss,temperature_min,temperature_max"}
	for _, r := range reports {
		lines = append(lines, fmt.Sprintf("%s,%d,%d,%.2f,%.4f,%.2f,%.2f", r.WindowID, r.Frames, r.Commands, r.AverageSignal, r.PacketLoss, r.TemperatureMin, r.TemperatureMax))
	}
	return strings.Join(lines, "\n")
}
