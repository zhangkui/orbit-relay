package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address            string        `json:"address"`
	Database           string        `json:"database"`
	WorkerCount        int           `json:"worker_count"`
	CommandTimeout     time.Duration `json:"command_timeout"`
	TelemetryRetention time.Duration `json:"telemetry_retention"`
	AllowedBands       []string      `json:"allowed_bands"`
	Debug              bool          `json:"debug"`
}

func Default() Config {
	return Config{Address: ":8080", Database: "data/orbit-relay.db", WorkerCount: 2, CommandTimeout: 10 * time.Second, TelemetryRetention: 30 * 24 * time.Hour, AllowedBands: []string{"UHF", "VHF", "S", "X"}}
}
func Load(path string) (Config, error) {
	c := Default()
	raw, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, err
	}
	return c, Validate(c)
}
func FromEnv() Config {
	c := Default()
	if x := os.Getenv("ORBIT_RELAY_ADDR"); x != "" {
		c.Address = x
	}
	if x := os.Getenv("ORBIT_RELAY_DB"); x != "" {
		c.Database = x
	}
	if x := os.Getenv("ORBIT_RELAY_WORKERS"); x != "" {
		if n, e := strconv.Atoi(x); e == nil {
			c.WorkerCount = n
		}
	}
	if os.Getenv("ORBIT_RELAY_DEBUG") == "1" {
		c.Debug = true
	}
	return c
}
func Validate(c Config) error {
	if strings.TrimSpace(c.Address) == "" || strings.TrimSpace(c.Database) == "" {
		return errors.New("address and database required")
	}
	if c.WorkerCount < 1 || c.WorkerCount > 64 {
		return errors.New("worker count outside range")
	}
	if c.CommandTimeout <= 0 || c.TelemetryRetention <= 0 {
		return errors.New("durations must be positive")
	}
	return nil
}
func HasBand(c Config, band string) bool {
	for _, x := range c.AllowedBands {
		if strings.EqualFold(x, band) {
			return true
		}
	}
	return false
}
func Save(path string, c Config) error {
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
