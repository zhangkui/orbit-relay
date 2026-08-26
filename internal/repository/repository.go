package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
)

type Repository struct{ db *store.Store }

func New(db *store.Store) *Repository                  { return &Repository{db: db} }
func (r *Repository) Put(kind, id string, v any) error { return r.db.Save(kind, id, v) }
func (r *Repository) Get(kind, id string, v any) error {
	err := r.db.Load(kind, id, v)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return err
}
func (r *Repository) Delete(kind, id string) error                  { return r.db.Delete(kind, id) }
func (r *Repository) List(kind string, fn func([]byte) error) error { return r.db.List(kind, fn) }
func Decode(raw []byte, v any) error                                { return json.Unmarshal(raw, v) }
func (r *Repository) Event(subject, action string) error            { return r.db.Event(subject, action) }
func (r *Repository) Tx(fn func(*sql.Tx) error) error               { return r.db.Transaction(fn) }
