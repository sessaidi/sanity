package sanity

import "time"

func ClampDuration(p *time.Duration, min, max time.Duration) {
	Clamp(p, min, max)
}

func DefaultDurationClamp(v, def, min, max time.Duration) time.Duration {
	return DefaultIfClamp(v, def, min, max)
}
