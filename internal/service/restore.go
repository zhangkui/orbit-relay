package service

import (
	"context"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
)

func Restore(r *repository.Repository) (map[string]model.ContactWindow, error) {
	out := map[string]model.ContactWindow{}
	err := r.List("window", func(raw []byte) error {
		var w model.ContactWindow
		if err := repository.Decode(raw, &w); err != nil {
			return err
		}
		out[w.ID] = w
		return nil
	})
	return out, err
}
func Ping(ctx context.Context) error { return ctx.Err() }
