package apps

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-apigateway/internal/handler"
	"github.com/MamangRust/monolith-point-of-sale-apigateway/internal/middlewares"
	"github.com/MamangRust/monolith-point-of-sale-pkg/auth"
	"github.com/MamangRust/monolith-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	otel_pkg "github.com/MamangRust/monolith-point-of-sale-pkg/otel"
	"github.com/MamangRust/monolith-point-of-sale-pkg/upload_image"
	response_api "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceAddresses struct {
	Auth        string
	Role        string
	User        string
	Category    string
	Cashier     string
	Merchant    string
	OrderItem   string
	Order       string
	Product     string
	Transaction string
}

func loadServiceAddresses() ServiceAddresses {
	return ServiceAddresses{
		Auth:        getEnvOrDefault("GRPC_AUTH_ADDR", "localhost:50051"),
		Role:        getEnvOrDefault("GRPC_ROLE_ADDR", "localhost:50052"),
		User:        getEnvOrDefault("GRPC_USER_ADDR", "localhost:50053"),
		Cashier:     getEnvOrDefault("GRPC_USER_ADDR", "localhost:50053"),
		Category:    getEnvOrDefault("GRPC_CATEGORY_ADDR", "localhost:50054"),
		Merchant:    getEnvOrDefault("GRPC_MERCHANT_ADDR", "localhost:50055"),
		OrderItem:   getEnvOrDefault("GRPC_ORDER_ITEM_ADDR", "localhost:50056"),
		Order:       getEnvOrDefault("GRPC_ORDER_ADDR", "localhost:50057"),
		Product:     getEnvOrDefault("GRPC_PRODUCT_ADDR", "localhost:50058"),
		Transaction: getEnvOrDefault("GRPC_TRANSACTION_ADDR", "localhost:50059"),
	}
}

func createServiceConnections(addresses ServiceAddresses, logger logger.LoggerInterface) (handler.ServiceConnections, error) {
	var connections handler.ServiceConnections

	conns := map[string]*string{
		"Auth":        &addresses.Auth,
		"Role":        &addresses.Role,
		"User":        &addresses.User,
		"Cashier":     &addresses.User,
		"Category":    &addresses.Category,
		"Merchant":    &addresses.Merchant,
		"OrderItem":   &addresses.OrderItem,
		"Order":       &addresses.Order,
		"Product":     &addresses.Product,
		"Transaction": &addresses.Transaction,
	}

	for name, addr := range conns {
		conn, err := createConnection(*addr, name, logger)
		if err != nil {
			return connections, err
		}
		switch name {
		case "Auth":
			connections.Auth = conn
		case "Role":
			connections.Role = conn
		case "User":
			connections.User = conn
		case "Cashier":
			connections.Cashier = conn
		case "Category":
			connections.Category = conn
		case "Merchant":
			connections.Merchant = conn
		case "OrderItem":
			connections.OrderItem = conn
		case "Order":
			connections.Order = conn
		case "Product":
			connections.Product = conn
		case "Transaction":
			connections.Transaction = conn
		}
	}

	return connections, nil
}

// @title PointOfsale gRPC
// @version 1.0
// @description gRPC based Point Of Sale service

// @host localhost:5000
// @BasePath /api/

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token obtained from login
// @security ApiKeyAuth
type Client struct {
	App    *echo.Echo
	Logger logger.LoggerInterface
}

func (c *Client) Shutdown(ctx context.Context) error {
	return c.App.Shutdown(ctx)
}

func RunClient() (*Client, func(), error) {
	flag.Parse()

	addresses := loadServiceAddresses()

	log, err := logger.NewLogger("apigateway")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	log.Debug("Loading environment variables")
	if err := dotenv.Viper(); err != nil {
		log.Fatal("Failed to load .env file", zap.Error(err))
	}

	log.Debug("Creating gRPC connections...")
	conns, err := createServiceConnections(addresses, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect services: %w", err)
	}

	e := setupEcho()

	token, err := auth.NewManager(viper.GetString("SECRET_KEY"))
	if err != nil {
		log.Fatal("Failed to create token manager", zap.Error(err))
	}

	ctx := context.Background()
	shutdownTracer, err := otel_pkg.InitTracerProvider("apigateway", ctx)
	if err != nil {
		log.Fatal("Failed to initialize tracer provider", zap.Error(err))
	}

	mapping := response_api.NewResponseApiMapper()
	image_upload := upload_image.NewImageUpload()

	depsHandler := &handler.Deps{
		Token:       token,
		E:           e,
		Logger:      log,
		Mapping:     mapping,
		ImageUpload: image_upload,
	}

	handler.NewHandler(depsHandler)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Info("Starting API Gateway server on :5000")
		if err := e.Start(":5000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Echo server error", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		log.Info("Starting Prometheus metrics server on :8091")
		if err := http.ListenAndServe(":8100", promhttp.Handler()); err != nil {
			log.Fatal("Metrics server error", zap.Error(err))
		}
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("Shutting down API Gateway...")
		if err := e.Shutdown(ctx); err != nil {
			log.Error("Echo shutdown failed", zap.Error(err))
		}

		closeConnections(conns, log)

		if shutdownTracer != nil {
			if err := shutdownTracer(context.Background()); err != nil {
				log.Error("Tracer shutdown failed", zap.Error(err))
			}
		}
	}

	return &Client{App: e, Logger: log}, shutdown, nil
}

func setupEcho() *echo.Echo {
	e := echo.New()

	limiter := middlewares.NewRateLimiter(20, 50)
	e.Use(limiter.Limit, middleware.Recover(), middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:1420", "http://localhost:33451"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key"},
		AllowCredentials: true,
	}))

	middlewares.WebSecurityConfig(e)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e
}

func createConnection(address, serviceName string, logger logger.LoggerInterface) (*grpc.ClientConn, error) {
	logger.Info(fmt.Sprintf("Connecting to %s service at %s", serviceName, address))
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to %s service", serviceName), zap.Error(err))
		return nil, err
	}
	return conn, nil
}

func closeConnections(conns handler.ServiceConnections, log logger.LoggerInterface) {
	connsMap := map[string]*grpc.ClientConn{
		"Auth":        conns.Auth,
		"Role":        conns.Role,
		"User":        conns.User,
		"Cashier":     conns.Cashier,
		"Category":    conns.Category,
		"Merchant":    conns.Merchant,
		"OrderItem":   conns.OrderItem,
		"Order":       conns.Order,
		"Product":     conns.Product,
		"Transaction": conns.Transaction,
	}

	for name, conn := range connsMap {
		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Error(fmt.Sprintf("Failed to close %s connection", name), zap.Error(err))
			}
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
