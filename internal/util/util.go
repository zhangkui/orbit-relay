package util

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ID(prefix string, at time.Time, sequence uint64) string {
	return fmt.Sprintf("%s-%s-%d", prefix, at.UTC().Format("20060102T150405.000000000Z"), sequence)
}
func Hash(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
func ParseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, errors.New("invalid boolean")
	}
}
func ParseDuration(value string, defaultValue time.Duration) (time.Duration, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, errors.New("duration must be positive")
	}
	return d, nil
}
func ParseInt(value string, defaultValue int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(value)
	return n, err
}
func Coalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func Truncate(value string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(value)
	if len(r) <= max {
		return value
	}
	return string(r[:max])
}
func Unique(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
func Contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func UTC(t time.Time) time.Time { return t.UTC().Truncate(time.Millisecond) }
func SameMinute(a, b time.Time) bool {
	a = UTC(a)
	b = UTC(b)
	return a.Year() == b.Year() && a.YearDay() == b.YearDay() && a.Hour() == b.Hour() && a.Minute() == b.Minute()
}
func Percent(part, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}
func Limit(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
func SplitNonEmpty(value, sep string) []string {
	parts := strings.Split(value, sep)
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
func JoinNonEmpty(values []string, sep string) string {
	out := []string{}
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, strings.TrimSpace(v))
		}
	}
	return strings.Join(out, sep)
}
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
func IsZero(t time.Time) bool { return t.IsZero() || t.Equal(time.Unix(0, 0)) }
func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
