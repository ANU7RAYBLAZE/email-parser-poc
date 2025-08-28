package outgoing

import (
	"context"
	"email-parser-poc/internal/domain/entities"
)

type StorageService interface {
	UploadEmails(ctx context.Context, emails *entities.EmailList) (string, error)
}
