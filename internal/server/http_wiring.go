package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	deliveryhttp "github.com/fauzie/golang-sekeleton/internal/delivery/http"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/pkg/claims"
	"github.com/fauzie/golang-sekeleton/pkg/httputil"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/middleware"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// httpWiringConfig carries everything buildHTTPHandler needs, kept as one
// struct instead of a long parameter list.
type httpWiringConfig struct {
	GRPCAddr         string // loopback address of the local gRPC server, e.g. "127.0.0.1:9090"
	ReadTimeout      time.Duration
	ValidateJWT      bool
	SwaggerEnabled   bool
	ProtoDir         string
	ExcludedLogPaths map[string]bool
	JWT              *claims.JWTConfig
	CORS             *middleware.CorsConfig
	Repo             *repository.Repository
}

// buildHTTPHandler wires the gRPC-Gateway mux (REST, generated from the
// same proto contract as gRPC), health checks, and optionally Swagger UI,
// then wraps the whole thing in the middleware chain. Execution order for
// an incoming request, outermost first: Recovery -> Claims -> otelhttp ->
// Logging -> Timeout -> ResponseWrapper -> CORS -> mux. That is the
// doc-mirrored order: each layer adds something (trace ID, claims, log
// context) the next one relies on.
func buildHTTPHandler(ctx context.Context, log *logger.Logger, cfg httpWiringConfig) (http.Handler, error) {
	gwmux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := pb.RegisterUserGrpcServiceHandlerFromEndpoint(ctx, gwmux, cfg.GRPCAddr, dialOpts); err != nil {
		return nil, fmt.Errorf("server: register user gateway: %w", err)
	}
	// [INJECTION POINT: Gateway Handler Registration]

	topMux := http.NewServeMux()

	healthHandler := deliveryhttp.NewHealthHandler(cfg.Repo)
	topMux.HandleFunc("/health/live", healthHandler.Live)
	topMux.HandleFunc("/health/ready", healthHandler.Ready)
	topMux.HandleFunc("/health", healthHandler.Ready)
	topMux.HandleFunc("/hc", healthHandler.Ready)

	if cfg.SwaggerEnabled {
		topMux.Handle("/swagger/", deliveryhttp.SwaggerHandler(cfg.ProtoDir))
	}

	topMux.Handle("/", gwmux)

	excludedPaths := map[string]bool{"/health/live": true, "/health/ready": true, "/health": true, "/hc": true}
	for p := range cfg.ExcludedLogPaths {
		excludedPaths[p] = true
	}

	var handler http.Handler = topMux
	handler = middleware.CORSMiddleware(cfg.CORS)(handler)
	handler = httputil.ResponseWrapperMiddleware(log)(handler)
	handler = middleware.TimeoutMiddleware(cfg.ReadTimeout, map[string]bool{})(handler)
	handler = middleware.LoggingMiddleware(log, excludedPaths)(handler)
	handler = otelhttp.NewHandler(handler, "golang-sekeleton")
	handler = middleware.ClaimsMiddleware(cfg.JWT, cfg.ValidateJWT)(handler)
	handler = middleware.RecoveryMiddleware(log)(handler)

	return handler, nil
}
