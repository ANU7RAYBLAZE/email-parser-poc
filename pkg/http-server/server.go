package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Server          *http.Server
	ShutdownTimeout time.Duration
}

type Config struct {
	Port            string
	Host            string
	Handler         http.Handler
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

func NewConfig(config Config) *Server {
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 10 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 60 * time.Second
	}

	httpServer := http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:      config.Handler,
		WriteTimeout: config.WriteTimeout,
		ReadTimeout:  config.ReadTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		Server:          &httpServer,
		ShutdownTimeout: config.ShutdownTimeout,
	}

}

func (s *Server) Start() error {
	fmt.Printf("Starting server on the %s\n", s.Server.Addr)

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *Server) StaertWithGracefulShutdown() error {
	// Channel to receive OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// channel to recieve server errors
	serverError := make(chan error, 1)

	go func() {
		log.Default().Output(1, "HTTP server starting")

		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Default().Output(1, "HTTP server starting")
			serverError <- fmt.Errorf("failed to start server: %w", err)
		}
	}()

	select {
	case err := <-serverError:
		return err
	case <-stop:
		shutdownStart := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
		defer cancel()

		if err := s.Server.Shutdown(ctx); err != nil {
			// shutdownDuration := time.Since(shutdownStart)
			return err
		}

		shutdownDuration := time.Since(shutdownStart)

		fmt.Println(shutdownDuration)
		return nil
	}

}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
