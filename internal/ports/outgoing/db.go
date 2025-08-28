package outgoing

import (
	"context"
	"email-parser-poc/internal/domain/entities"
)

type DbService interface {
	UploadHeaders(ctx context.Context, emails *entities.EmailList) error
}
