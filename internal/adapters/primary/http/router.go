package http

import (
	"email-parser-poc/internal/adapters/primary/http/handlers"
	"email-parser-poc/internal/adapters/seondary/dynamodb"
	"email-parser-poc/internal/adapters/seondary/gmail"
	"email-parser-poc/internal/adapters/seondary/s3bucket"
	"email-parser-poc/internal/adapters/seondary/token"
	"email-parser-poc/internal/application_api"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type RouterConfig struct {
	Version     string
	AccessToken string
}

func NewRouter(config RouterConfig) http.Handler {
	// new router by initilizing chi NewRouter method
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Recoverer)
	// r.Use(middleware.Timeout(30 * time.Millisecond))
	r.Use(middleware.Heartbeat("/ping"))

	// Intilizing healthservice for dependency for healthhandler
	healthService := application_api.NewHealthService(config.Version)
	healthHandler := handlers.NewHealthHandler(healthService)

	token1 := config.AccessToken
	fmt.Printf("token : %s", token1)

	tokenProvider := token.NewStaticTokenProvider(config.AccessToken)
	emailRepo := gmail.NewGmailRepository(tokenProvider)
	storageService, err := s3bucket.NewS3Storage("sample-bucket", "http://localhost:4566")
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}
	dbService, err := dynamodb.NewDynamoDbInstance("http://localhost:4566")
	if err != nil {
		log.Fatalf("Failed to initialize Db storage: %v", err)
	}
	emailService := application_api.NewEmailService(emailRepo, storageService, dbService)
	emailHandler := handlers.NewEmailHandler(emailService)

	r.Route("/health", func(r chi.Router) {
		r.Get("/", healthHandler.CheckHealth)
	})

	r.Get("/emails/all", emailHandler.GetAllEmails)

	return r
}
