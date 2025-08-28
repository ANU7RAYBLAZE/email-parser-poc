package application_api

import (
	"context"
	"email-parser-poc/internal/domain/entities"
	"time"
)

type HealthService struct {
	StartTime time.Time
	Version   string
}

func NewHealthService(version string) *HealthService {
	return &HealthService{
		StartTime: time.Now(),
		Version:   version,
	}
}

func (s *HealthService) CheckHealth(ctx context.Context) (*entities.Health, error) {
	health := &entities.Health{
		Status:    entities.HealthStatusHealthy,
		TimeStamp: time.Now(),
		Version:   s.Version,
		Uptime:    time.Since(s.StartTime),
	}
	return health, nil
}
