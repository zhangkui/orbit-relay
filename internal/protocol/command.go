package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
)

var opcodes = map[string]byte{"PING": 1, "SET_MODE": 2, "CAPTURE": 3, "REBOOT": 4, "DUMP": 5}

func EncodeCommand(packet model.CommandPacket) ([]byte, error) {
	op, ok := opcodes[packet.Opcode]
	if !ok {
		return nil, fmt.Errorf("opcode %q: %w", packet.Opcode, model.ErrInvalid)
	}
	keys := make([]string, 0, len(packet.Arguments))
	for k := range packet.Arguments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := bytes.NewBuffer(nil)
	b.WriteByte(1)
	b.WriteByte(op)
	if err := binary.Write(b, binary.BigEndian, packet.Sequence); err != nil {
		return nil, err
	}
	b.WriteByte(byte(len(keys)))
	for _, k := range keys {
		v := packet.Arguments[k]
		if len(k) > 255 || len(v) > 255 {
			return nil, model.ErrInvalid
		}
		b.WriteByte(byte(len(k)))
		b.WriteString(k)
		b.WriteByte(byte(len(v)))
		b.WriteString(v)
	}
	return b.Bytes(), nil
}
func ValidateCommand(packet model.CommandPacket) error {
	if packet.ID == "" || packet.SatelliteID == "" || packet.WindowID == "" {
		return model.ErrInvalid
	}
	if _, ok := opcodes[packet.Opcode]; !ok {
		return model.ErrInvalid
	}
	if packet.Priority < 0 || packet.Priority > 100 {
		return model.ErrInvalid
	}
	return nil
}
