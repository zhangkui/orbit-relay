package model

import "time"

type Satellite struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	NORAD   int    `json:"norad"`
	Orbit   string `json:"orbit"`
	Enabled bool   `json:"enabled"`
}
type GroundStation struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Antennas  []string `json:"antennas"`
}
type ContactWindow struct {
	ID          string    `json:"id"`
	SatelliteID string    `json:"satellite_id"`
	StationID   string    `json:"station_id"`
	Antenna     string    `json:"antenna"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	State       string    `json:"state"`
	Priority    int       `json:"priority"`
	Failure     string    `json:"failure,omitempty"`
}
type TelemetryFrame struct {
	ID          string    `json:"id"`
	SatelliteID string    `json:"satellite_id"`
	StationID   string    `json:"station_id"`
	ReceivedAt  time.Time `json:"received_at"`
	Sequence    uint32    `json:"sequence"`
	Payload     []byte    `json:"payload"`
	Checksum    uint32    `json:"checksum"`
	SignalDB    float64   `json:"signal_db"`
	Temperature float64   `json:"temperature"`
	Voltage     float64   `json:"voltage"`
	Quality     string    `json:"quality"`
}
type CommandPacket struct {
	ID          string            `json:"id"`
	SatelliteID string            `json:"satellite_id"`
	WindowID    string            `json:"window_id"`
	Sequence    uint32            `json:"sequence"`
	Opcode      string            `json:"opcode"`
	Arguments   map[string]string `json:"arguments"`
	Priority    int               `json:"priority"`
	State       string            `json:"state"`
	CreatedAt   time.Time         `json:"created_at"`
	SentAt      *time.Time        `json:"sent_at,omitempty"`
	Error       string            `json:"error,omitempty"`
}
type LinkReport struct {
	WindowID       string    `json:"window_id"`
	Frames         int       `json:"frames"`
	Commands       int       `json:"commands"`
	AverageSignal  float64   `json:"average_signal"`
	PacketLoss     float64   `json:"packet_loss"`
	TemperatureMin float64   `json:"temperature_min"`
	TemperatureMax float64   `json:"temperature_max"`
	GeneratedAt    time.Time `json:"generated_at"`
}

const (
	WindowPlanned   = "planned"
	WindowActive    = "active"
	WindowCompleted = "completed"
	WindowFailed    = "failed"
	WindowCancelled = "cancelled"
	CommandQueued   = "queued"
	CommandSending  = "sending"
	CommandSent     = "sent"
	CommandFailed   = "failed"
)
