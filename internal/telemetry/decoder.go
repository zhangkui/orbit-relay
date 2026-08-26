package telemetry

import (
	"encoding/binary"
	"errors"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"math"
	"time"
)

type FieldType byte

const (
	Unsigned FieldType = iota
	Signed
	Float
	Boolean
	Text
)

type Field struct {
	Name     string
	Offset   int
	Width    int
	Type     FieldType
	Scale    float64
	Unit     string
	Min      float64
	Max      float64
	Required bool
}
type Schema struct {
	ID              string
	Version         int
	Fields          []Field
	ChecksumOffset  int
	SequenceOffset  int
	TimestampOffset int
}
type Value struct {
	Name   string
	Number float64
	Text   string
	Bool   bool
	Unit   string
	Valid  bool
}
type Decoder struct{ Schemas map[string]Schema }

func NewDecoder() *Decoder { return &Decoder{Schemas: map[string]Schema{}} }
func (d *Decoder) Register(s Schema) error {
	if s.ID == "" || s.Version <= 0 || len(s.Fields) == 0 {
		return model.ErrInvalid
	}
	d.Schemas[s.ID] = s
	return nil
}
func (d *Decoder) Decode(schemaID string, raw []byte) ([]Value, error) {
	s, ok := d.Schemas[schemaID]
	if !ok {
		return nil, model.ErrNotFound
	}
	out := []Value{}
	for _, f := range s.Fields {
		if f.Offset < 0 || f.Width <= 0 || f.Offset+f.Width > len(raw) {
			if f.Required {
				return nil, fmt.Errorf("field %s out of range: %w", f.Name, model.ErrInvalid)
			}
			continue
		}
		v, err := readField(f, raw[f.Offset:f.Offset+f.Width])
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
func readField(f Field, b []byte) (Value, error) {
	v := Value{Name: f.Name, Unit: f.Unit, Valid: true}
	switch f.Type {
	case Unsigned:
		v.Number = float64(readUint(b))
	case Signed:
		v.Number = float64(readInt(b))
	case Float:
		if len(b) != 4 {
			return v, errors.New("float field must be four bytes")
		}
		v.Number = float64(math.Float32frombits(binary.BigEndian.Uint32(b)))
	case Boolean:
		v.Bool = b[0] != 0
	case Text:
		v.Text = string(b)
	default:
		return v, model.ErrInvalid
	}
	if f.Type == Unsigned || f.Type == Signed || f.Type == Float {
		v.Number *= f.Scale
		if f.Min != 0 && v.Number < f.Min {
			v.Valid = false
		}
		if f.Max != 0 && v.Number > f.Max {
			v.Valid = false
		}
	}
	return v, nil
}
func readUint(b []byte) uint64 {
	var x uint64
	for _, v := range b {
		x = x<<8 | uint64(v)
	}
	return x
}
func readInt(b []byte) int64 {
	u := readUint(b)
	bits := uint(len(b) * 8)
	if bits > 0 && u&(uint64(1)<<(bits-1)) != 0 {
		u |= ^uint64(0) << bits
	}
	return int64(u)
}
func EncodeValues(fields []Field, values map[string]Value) []byte {
	size := 0
	for _, f := range fields {
		if f.Offset+f.Width > size {
			size = f.Offset + f.Width
		}
	}
	out := make([]byte, size)
	for _, f := range fields {
		v, ok := values[f.Name]
		if !ok {
			continue
		}
		n := int64(v.Number)
		for i := f.Width - 1; i >= 0; i-- {
			out[f.Offset+i] = byte(n)
			n >>= 8
		}
	}
	return out
}
func (d *Decoder) ExtractMetrics(values []Value) map[string]float64 {
	out := map[string]float64{}
	for _, v := range values {
		if v.Valid {
			out[v.Name] = v.Number
		}
	}
	return out
}
func NormalizeTimestamp(ts int64, epoch time.Time) time.Time {
	return epoch.Add(time.Duration(ts) * time.Millisecond).UTC()
}
func Age(ts, now time.Time) time.Duration { return now.Sub(ts) }
