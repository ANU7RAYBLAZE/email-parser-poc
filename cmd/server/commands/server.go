/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	httpAdapter "email-parser-poc/internal/adapters/primary/http"
	"email-parser-poc/internal/adapters/seondary/config"
	httpserver "email-parser-poc/pkg/http-server"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	port string
	host string
)

// server/commands/serverCmd represents the server/commands/server command
var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long: `Start the HTTP server with graceful shutdown capabilities.

The server will listen for incoming HTTP requests and handle them
according to the configured routes and middleware. The server
supports graceful shutdown on SIGINT and SIGTERM signals.`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// server/commands/serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// server/commands/serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	serverCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host to bind to")
}

func runServe(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Config loaded successfully!\n")
	fmt.Printf("Access Token: '%s'\n", cfg.Auth.AccessToken)
	fmt.Printf("Access Token Length: %d\n", len(cfg.Auth.AccessToken))

	routerConfig := httpAdapter.RouterConfig{
		Version:     "1.0.0",
		AccessToken: cfg.GetAccessToken(),
	}

	router := httpAdapter.NewRouter(routerConfig)

	serverConfig := httpserver.Config{
		Port:            port,
		Host:            host,
		Handler:         router,
		ShutdownTimeout: 0,
		ReadTimeout:     0,
		WriteTimeout:    0,
		IdleTimeout:     0,
	}

	server := httpserver.NewConfig(serverConfig)
	return server.StaertWithGracefulShutdown()
}
