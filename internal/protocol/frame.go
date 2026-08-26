package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"hash/crc32"
	"time"
)

const HeaderSize = 20

type Header struct {
	Version     byte
	Flags       byte
	Satellite   uint16
	Sequence    uint32
	PayloadSize uint16
	Checksum    uint32
	Epoch       uint32
}

func EncodeFrame(frame model.TelemetryFrame) ([]byte, error) {
	if len(frame.Payload) > 65535 {
		return nil, fmt.Errorf("payload too large: %w", model.ErrInvalid)
	}
	b := bytes.NewBuffer(nil)
	h := Header{Version: 1, Satellite: uint16(frame.Sequence), Sequence: frame.Sequence, PayloadSize: uint16(len(frame.Payload)), Checksum: crc32.ChecksumIEEE(frame.Payload), Epoch: uint32(frame.ReceivedAt.Unix())}
	if err := binary.Write(b, binary.BigEndian, h); err != nil {
		return nil, err
	}
	b.Write(frame.Payload)
	return b.Bytes(), nil
}
func DecodeFrame(raw []byte, satelliteID, stationID string) (model.TelemetryFrame, error) {
	if len(raw) < HeaderSize {
		return model.TelemetryFrame{}, fmt.Errorf("short frame: %w", model.ErrInvalid)
	}
	var h Header
	if err := binary.Read(bytes.NewReader(raw[:HeaderSize]), binary.BigEndian, &h); err != nil {
		return model.TelemetryFrame{}, err
	}
	if int(h.PayloadSize) != len(raw)-HeaderSize {
		return model.TelemetryFrame{}, fmt.Errorf("size mismatch: %w", model.ErrInvalid)
	}
	p := append([]byte(nil), raw[HeaderSize:]...)
	return model.TelemetryFrame{ID: fmt.Sprintf("%s-%d", satelliteID, h.Sequence), SatelliteID: satelliteID, StationID: stationID, ReceivedAt: time.Unix(int64(h.Epoch), 0).UTC(), Sequence: h.Sequence, Payload: p, Checksum: h.Checksum}, nil
}
func ValidateFrame(frame model.TelemetryFrame) error {
	if crc32.ChecksumIEEE(frame.Payload) != frame.Checksum {
		return model.ErrChecksum
	}
	if frame.SatelliteID == "" || frame.StationID == "" {
		return model.ErrInvalid
	}
	return nil
}
