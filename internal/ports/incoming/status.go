package incoming

import (
	"context"
	"email-parser-poc/internal/domain/entities"
)

type HealthPort interface {
	CheckHealth(ctx context.Context) (*entities.Health, error)
}
