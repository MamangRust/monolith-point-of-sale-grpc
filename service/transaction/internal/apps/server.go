package apps

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/MamangRust/monolith-point-of-sale-pkg/database"
	db "github.com/MamangRust/monolith-point-of-sale-pkg/database/schema"
	"github.com/MamangRust/monolith-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/monolith-point-of-sale-pkg/kafka"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	otel_pkg "github.com/MamangRust/monolith-point-of-sale-pkg/otel"
	"github.com/MamangRust/monolith-point-of-sale-shared/pb"
	"github.com/MamangRust/monolith-point-of-sale-transacton/internal/handler"
	"github.com/MamangRust/monolith-point-of-sale-transacton/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-transacton/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	port int
)

func init() {
	port = viper.GetInt("GRPC_TRANSACTION_ADDR")
	if port == 0 {
		port = 50060
	}

	flag.IntVar(&port, "port", port, "gRPC server port")
}

type Server struct {
	Logger   logger.LoggerInterface
	DB       *db.Queries
	Services *service.Service
	Handlers *handler.Handler
	Ctx      context.Context
}

func NewServer() (*Server, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if err := dotenv.Viper(); err != nil {
		logger.Fatal("Failed to load .env file", zap.Error(err))
	}
	flag.Parse()

	conn, err := database.NewClient(logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	DB := db.New(conn)

	ctx := context.Background()

	depsRepo := repository.Deps{
		DB:  DB,
		Ctx: ctx,
	}

	repositories := repository.NewRepositories(depsRepo)
	myKafka := kafka.NewKafka(logger, []string{viper.GetString("KAFKA_BROKERS")})

	shutdownTracerProvider, err := otel_pkg.InitTracerProvider("TRANSACTION-service", ctx)
	if err != nil {
		logger.Fatal("Failed to initialize tracer provider", zap.Error(err))
	}
	defer func() {
		if err := shutdownTracerProvider(ctx); err != nil {
			logger.Fatal("Failed to shutdown tracer provider", zap.Error(err))
		}
	}()

	services := service.NewService(service.Deps{
		Ctx:          ctx,
		Repositories: repositories,
		Logger:       logger,
		Kafka:        *myKafka,
	})

	handlers := handler.NewHandler(handler.Deps{
		Service: *services,
	})

	return &Server{
		Logger:   logger,
		DB:       DB,
		Services: services,
		Handlers: handlers,
		Ctx:      ctx,
	}, nil
}

func (s *Server) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		s.Logger.Fatal("Failed to listen", zap.Error(err))
	}
	metricsAddr := fmt.Sprintf(":%s", viper.GetString("METRIC_TRANSACTION_ADDR"))
	metricsLis, err := net.Listen("tcp", metricsAddr)
	if err != nil {
		s.Logger.Fatal("failed to listen on", zap.Error(err))
	}

	if err != nil {
		s.Logger.Fatal("Failed to listen for metrics", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
				otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
			),
		),
	)

	pb.RegisterTransactionServiceServer(grpcServer, s.Handlers.Transaction)

	metricsServer := http.NewServeMux()
	metricsServer.Handle("/metrics", promhttp.Handler())

	s.Logger.Info(fmt.Sprintf("Server running on port %d", port))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.Logger.Info("Metrics server listening on :8082")
		if err := http.Serve(metricsLis, metricsServer); err != nil {
			s.Logger.Fatal("Metrics server error", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		s.Logger.Info("gRPC server listening on :50052")
		if err := grpcServer.Serve(lis); err != nil {
			s.Logger.Fatal("gRPC server error", zap.Error(err))
		}
	}()

	wg.Wait()
}
