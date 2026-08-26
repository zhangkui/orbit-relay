package command

import (
	"context"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/window"
	"time"
)

func CanSend(w model.ContactWindow, p model.CommandPacket, now time.Time) error {
	if !window.CanQueue(w, now) {
		return model.ErrWindowClosed
	}
	if p.SatelliteID != w.SatelliteID || p.WindowID != w.ID {
		return model.ErrConflict
	}
	return nil
}
func CheckContext(ctx context.Context) error { return ctx.Err() }
