package outgoing

import (
	"context"
	"email-parser-poc/internal/domain/entities"
)

type EmailRepository interface {
	FetchEmails(ctx context.Context, filter entities.EmailFilter) (*entities.EmailList, error)
	
}

type TokenProvider interface {
	GetAccessToken() string
}
