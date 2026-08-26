package model

import "errors"

var (
	ErrNotFound        = errors.New("orbit relay entity not found")
	ErrConflict        = errors.New("orbit relay state conflict")
	ErrInvalid         = errors.New("orbit relay invalid input")
	ErrWindowClosed    = errors.New("contact window is closed")
	ErrChecksum        = errors.New("telemetry checksum mismatch")
	ErrSequence        = errors.New("telemetry sequence out of order")
	ErrCommandRejected = errors.New("command rejected")
)
