package server

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	deliverygrpc "github.com/fauzie/golang-sekeleton/internal/delivery/grpc"
	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	"github.com/fauzie/golang-sekeleton/internal/mapper"
	"github.com/fauzie/golang-sekeleton/pkg/claims"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/middleware"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// publicGRPCMethods lists RPCs reachable without an authenticated caller.
// Add new public RPCs below the marker — this is the "bypass list" for
// middleware.UnaryServerAuthInterceptor.
var publicGRPCMethods = []string{
	"/grpc.health.v1.Health/Check",
	// [INJECTION POINT: Public gRPC Methods]
}

func buildGRPCServer(log *logger.Logger, jwtCfg *claims.JWTConfig, validateJWT bool, excludedLogPaths map[string]bool, userEndpoints endpoint.UserEndpoints) *grpc.Server {
	excluded := map[string]bool{"/grpc.health.v1.Health/Check": true}
	for p := range excludedLogPaths {
		excluded[p] = true
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecovery(log),
			middleware.UnaryServerClaims(jwtCfg, validateJWT),
			middleware.UnaryServerAuthInterceptor(publicGRPCMethods),
			middleware.UnaryServerLogger(log, excluded),
			middleware.UnaryServerErrorHandler(),
		),
	)

	userService := deliverygrpc.NewUserService(userEndpoints, mapper.NewUserMapper(), log)
	pb.RegisterUserGrpcServiceServer(grpcServer, userService)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	return grpcServer
}
