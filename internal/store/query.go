package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type EventRecord struct {
	ID        int64
	Subject   string
	Action    string
	CreatedAt time.Time
}

func (s *Store) Count(kind string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT count(*) FROM records WHERE kind=?`, kind).Scan(&n)
	return n, err
}
func (s *Store) Events(subject string, limit int) ([]EventRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id,subject,action,created_at FROM events WHERE subject=? ORDER BY id DESC LIMIT ?`, subject, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EventRecord{}
	for rows.Next() {
		var x EventRecord
		var raw string
		if err := rows.Scan(&x.ID, &x.Subject, &x.Action, &raw); err != nil {
			return nil, err
		}
		x.CreatedAt, _ = time.Parse(time.RFC3339Nano, raw)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) SaveMany(kind string, items map[string]any) error {
	return s.Transaction(func(tx *sql.Tx) error {
		for id, v := range items {
			raw, err := json.Marshal(v)
			if err != nil {
				return err
			}
			if _, err = tx.Exec(`INSERT INTO records(kind,id,payload,updated_at) VALUES(?,?,?,?) ON CONFLICT(kind,id) DO UPDATE SET payload=excluded.payload,updated_at=excluded.updated_at`, kind, id, raw, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
				return fmt.Errorf("save %s: %w", id, err)
			}
		}
		return nil
	})
}
func (s *Store) Purge(kind string, before time.Time) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM records WHERE kind=? AND updated_at<=?`, kind, before.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
func (s *Store) Health() error { var n int; return s.db.QueryRow(`SELECT 1`).Scan(&n) }
