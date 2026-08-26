package export

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/report"
	"io"
	"strconv"
	"strings"
	"time"
)

type Bundle struct {
	GeneratedAt time.Time              `json:"generated_at"`
	Windows     []model.ContactWindow  `json:"windows"`
	Frames      []model.TelemetryFrame `json:"frames"`
	Commands    []model.CommandPacket  `json:"commands"`
	Reports     []model.LinkReport     `json:"reports"`
}

func NewBundle() Bundle {
	return Bundle{GeneratedAt: time.Now().UTC(), Windows: []model.ContactWindow{}, Frames: []model.TelemetryFrame{}, Commands: []model.CommandPacket{}, Reports: []model.LinkReport{}}
}
func JSON(b Bundle) ([]byte, error) { return json.MarshalIndent(b, "", "  ") }
func WriteJSON(w io.Writer, b Bundle) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}
func WriteReports(w io.Writer, items []model.LinkReport) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"window_id", "frames", "commands", "average_signal", "packet_loss", "generated_at"}); err != nil {
		return err
	}
	for _, r := range items {
		if err := cw.Write([]string{r.WindowID, strconv.Itoa(r.Frames), strconv.Itoa(r.Commands), fmt.Sprintf("%.2f", r.AverageSignal), fmt.Sprintf("%.4f", r.PacketLoss), r.GeneratedAt.Format(time.RFC3339)}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
func ReadLines(r io.Reader) ([]string, error) {
	s := bufio.NewScanner(r)
	out := []string{}
	for s.Scan() {
		line := strings.Trim(s.Text(), " \t")
		if line != "" {
			out = append(out, line)
		}
	}
	return out, s.Err()
}
func SummaryText(r model.LinkReport) string { return report.Text(r) }
func ParseReport(raw []byte) (model.LinkReport, error) {
	var r model.LinkReport
	err := json.Unmarshal(raw, &r)
	return r, err
}
