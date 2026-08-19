// Package server assembles every layer into one running service: it is
// the only package that knows about every concrete infrastructure
// dependency (MySQL/PostgreSQL, Redis, Kafka/RabbitMQ/Pub/Sub) at once,
// and owns the full startup/shutdown lifecycle.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/config"
	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/internal/usecase"
	"github.com/fauzie/golang-sekeleton/internal/validator"
	"github.com/fauzie/golang-sekeleton/pkg/claims"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/messaging"
	"github.com/fauzie/golang-sekeleton/pkg/middleware"

	"google.golang.org/grpc"
)

// Server owns every long-lived resource the process holds: the two
// listeners (HTTP/gRPC), the repository (DB/cache/MQ), and every message
// consumer. Shutdown order is the mirror of startup order — see Shutdown.
type Server struct {
	cfg  *config.Config
	log  *logger.Logger
	repo *repository.Repository

	httpServer *http.Server
	grpcServer *grpc.Server
	grpcLis    net.Listener

	consumers []messaging.Consumer
}

// New builds every layer (repository -> validator/mapper/usecase/endpoint
// -> gRPC service -> REST gateway -> message consumers) but does not start
// listening yet — call Start for that. Splitting the two means New can
// fail fast on misconfiguration before anything is bound to a port.
func New(ctx context.Context, cfg *config.Config, log *logger.Logger) (*Server, error) {
	repo, err := buildRepository(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("server: build repository: %w", err)
	}

	userValidator := validator.NewUserValidator()
	userUseCase := usecase.NewUserUseCase(repo, userValidator, log)
	userEndpoints := endpoint.MakeUserEndpoints(userUseCase)

	jwtCfg := &claims.JWTConfig{Secret: cfg.JWT.Secret, Issuer: cfg.JWT.Issuer}
	excludedLogPaths := map[string]bool{}
	for _, p := range cfg.Logging.ExcludedLogPaths {
		excludedLogPaths[p] = true
	}

	grpcServer := buildGRPCServer(log, jwtCfg, cfg.Server.ValidateJWT, excludedLogPaths, userEndpoints)

	grpcAddr := fmt.Sprintf("127.0.0.1:%d", cfg.Server.GRPCPort)
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("server: listen grpc: %w", err)
	}

	httpHandler, err := buildHTTPHandler(ctx, log, httpWiringConfig{
		GRPCAddr:         grpcAddr,
		ReadTimeout:      time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		ValidateJWT:      cfg.Server.ValidateJWT,
		SwaggerEnabled:   cfg.Server.SwaggerEnabled,
		ProtoDir:         "proto",
		ExcludedLogPaths: excludedLogPaths,
		JWT:              jwtCfg,
		CORS: &middleware.CorsConfig{
			AllowedOrigins:   cfg.CORS.AllowedOrigins,
			AllowedMethods:   cfg.CORS.AllowedMethods,
			AllowedHeaders:   cfg.CORS.AllowedHeaders,
			AllowCredentials: cfg.CORS.AllowCredentials,
		},
		Repo: repo,
	})
	if err != nil {
		return nil, fmt.Errorf("server: build http handler: %w", err)
	}

	consumers, err := startConsumers(ctx, cfg, log, userEndpoints)
	if err != nil {
		return nil, fmt.Errorf("server: start consumers: %w", err)
	}

	return &Server{
		cfg:  cfg,
		log:  log,
		repo: repo,
		httpServer: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Server.HTTPPort),
			Handler:           httpHandler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		grpcServer: grpcServer,
		grpcLis:    grpcLis,
		consumers:  consumers,
	}, nil
}

// Start begins serving gRPC and HTTP. It blocks until either listener
// fails; a caller normally runs it in a goroutine and waits on the
// returned channel alongside an OS signal (see cmd/server/main.go).
func (s *Server) Start() <-chan error {
	errCh := make(chan error, 2)

	go func() {
		s.log.Info("grpc server listening", zap.Int("port", s.cfg.Server.GRPCPort))
		if err := s.grpcServer.Serve(s.grpcLis); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	go func() {
		s.log.Info("http server listening", zap.Int("port", s.cfg.Server.HTTPPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	return errCh
}

// Shutdown runs the graceful-shutdown sequence in the order that keeps
// zero-drop guarantees: stop accepting new HTTP requests, stop gRPC
// (letting in-flight RPCs finish), stop message consumers so no new
// message is picked up mid-shutdown, then close every repository
// connection last so anything still finishing up above can still reach
// the database/cache/broker.
func (s *Server) Shutdown(ctx context.Context) error {
	var errs []error

	if err := s.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http shutdown: %w", err))
	}

	s.grpcServer.GracefulStop()

	for _, c := range s.consumers {
		if err := c.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("consumer stop: %w", err))
		}
	}

	if err := s.repo.Close(); err != nil {
		errs = append(errs, fmt.Errorf("repository close: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("server: shutdown errors: %v", errs)
	}
	return nil
}
