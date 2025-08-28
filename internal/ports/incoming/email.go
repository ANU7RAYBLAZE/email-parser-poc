package incoming

import (
	"context"
	"email-parser-poc/internal/domain/entities"
)

type EmailService interface {
	GetEmails(ctx context.Context, filter entities.EmailFilter) (*entities.EmailList, string, error)
}
