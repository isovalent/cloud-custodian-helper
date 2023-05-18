package date

import "time"

func ParseOrDefault(s string, d time.Time) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t
	}
	return d
}
