package circuitbreaker

import "time"

func valueOr(v, fallback uint32) uint32 {
	if v == 0 {
		return fallback
	}
	return v
}

func secondsOr(seconds, fallback int) time.Duration {
	if seconds <= 0 {
		seconds = fallback
	}
	return time.Duration(seconds) * time.Second
}
