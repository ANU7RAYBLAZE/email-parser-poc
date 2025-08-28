package application_api

import (
	"context"
	"email-parser-poc/internal/domain/entities"
	"email-parser-poc/internal/ports/incoming"
	"email-parser-poc/internal/ports/outgoing"
	"fmt"
)

type EmailServie struct {
	EmailRepo      outgoing.EmailRepository
	StorageService outgoing.StorageService
	Dbservice      outgoing.DbService
}

func NewEmailService(emailRepo outgoing.EmailRepository, storageService outgoing.StorageService, dbservice outgoing.DbService) incoming.EmailService {
	return &EmailServie{
		EmailRepo:      emailRepo,
		StorageService: storageService,
		Dbservice:      dbservice,
	}
}
func (s EmailServie) GetEmails(ctx context.Context, filter entities.EmailFilter) (*entities.EmailList, string, error) {
	emailList, err := s.EmailRepo.FetchEmails(ctx, filter)

	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch emails: %w", err)
	}
	fmt.Println("emaillist got")
	if err := s.Dbservice.UploadHeaders(ctx, emailList); err != nil {
		return nil, "", fmt.Errorf("failed to store emails-headers in db: %w", err)
	}
	filename, err := s.StorageService.UploadEmails(ctx, emailList)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store emails: %w", err)
	}

	fmt.Println("all function called")

	return emailList, filename, nil

}
