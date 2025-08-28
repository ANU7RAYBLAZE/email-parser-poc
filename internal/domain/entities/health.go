package entities

import "time"

// health constant type
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type Health struct {
	Status    HealthStatus           `json:"status"`
	TimeStamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Uptime    time.Duration          `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

type HealthCheck struct {
	Status     HealthStatus  `json:"status"`
	Message    string        `json:"message,omitempty"`
	Error      string        `json:"error,omitempty"`
	TimeStamps time.Time     `json:"timestamps"`
	Duration   time.Duration `json:"duration"`
}

func (h *Health) IsHealthy() bool {
	return h.Status == HealthStatusHealthy
}

func (h *Health) Addcheck(name string, check HealthCheck) {
	if h.Checks == nil {
		h.Checks = make(map[string]HealthCheck)
	}
	h.Checks[name] = check

	h.updateOverallStatus()
}

func (h *Health) updateOverallStatus() {
	if h.Checks == nil {
		h.Status = HealthStatusHealthy
		return
	}
	hasUnhealthy := false

	for _, checks := range h.Checks {
		switch checks.Status {
		case HealthStatusHealthy:
			hasUnhealthy = true
		}
	}

	if hasUnhealthy {
		h.Status = HealthStatusUnhealthy
	} else {
		h.Status = HealthStatusHealthy
	}
}
