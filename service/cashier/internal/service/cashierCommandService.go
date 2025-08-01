package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-cashier/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-cashier/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-cashier/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/merchant_errors"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/user_errors"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type cashierCommandService struct {
	mencache        mencache.CashierCommandCache
	errorHandler    errorhandler.CashierCommadError
	trace           trace.Tracer
	merchantQuery   repository.MerchantQueryRepository
	userQuery       repository.UserQueryRepository
	cashierCommand  repository.CashierCommandRepository
	mapping         response_service.CashierResponseMapper
	logger          logger.LoggerInterface
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewCashierCommandService(
	mencache mencache.CashierCommandCache,
	errorHandler errorhandler.CashierCommadError,
	merchantQuery repository.MerchantQueryRepository,
	userQuery repository.UserQueryRepository, cashierCommand repository.CashierCommandRepository,
	mapping response_service.CashierResponseMapper, logger logger.LoggerInterface,
) *cashierCommandService {

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cashier_command_service_requests_total",
			Help: "Total number of requests to the CashierCommandService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cashier_command_service_request_duration_seconds",
			Help:    "Histogram of request durations for the CashierCommandService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &cashierCommandService{
		mencache:        mencache,
		errorHandler:    errorHandler,
		trace:           otel.Tracer("cashier-command-service"),
		merchantQuery:   merchantQuery,
		userQuery:       userQuery,
		cashierCommand:  cashierCommand,
		mapping:         mapping,
		logger:          logger,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

func (s *cashierCommandService) CreateCashier(ctx context.Context, req *requests.CreateCashierRequest) (*response.CashierResponse, *response.ErrorResponse) {
	const method = "CreateCashier"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	_, err := s.merchantQuery.FindById(ctx, req.MerchantID)
	if err != nil {
		return errorhandler.HandleRepositorySingleError[*response.CashierResponse](s.logger, err, method, "FAILED_FIND_MERCHANT", span, &status, merchant_errors.ErrFailedFindMerchantById, zap.Error(err))
	}

	_, err = s.userQuery.FindById(ctx, req.UserID)
	if err != nil {
		return errorhandler.HandleRepositorySingleError[*response.CashierResponse](s.logger, err, method, "FAILED_FIND_USER", span, &status, user_errors.ErrUserNotFoundRes, zap.Error(err))
	}

	cashier, err := s.cashierCommand.CreateCashier(ctx, req)
	if err != nil {
		return s.errorHandler.HandleCreateCashierError(err, method, "FAILED_CREATE_CASHIER", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCashierResponse(cashier)

	logSuccess("Successfully created cashier", zap.Bool("success", true))

	return so, nil
}

func (s *cashierCommandService) UpdateCashier(ctx context.Context, req *requests.UpdateCashierRequest) (*response.CashierResponse, *response.ErrorResponse) {
	const method = "UpdateCashier"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	cashier, err := s.cashierCommand.UpdateCashier(ctx, req)
	if err != nil {
		return s.errorHandler.HandleUpdateCashierError(err, method, "FAILED_UPDATE_CASHIER", span, &status, zap.Error(err))
	}

	span.SetAttributes(
		attribute.String("cashier.name", cashier.Name),
	)

	so := s.mapping.ToCashierResponse(cashier)

	s.mencache.DeleteCashierCache(ctx, cashier.ID)

	logSuccess("Successfully updated cashier", zap.Bool("success", true))

	return so, nil
}

func (s *cashierCommandService) TrashedCashier(ctx context.Context, cashierID int) (*response.CashierResponseDeleteAt, *response.ErrorResponse) {
	const method = "TrashedCashier"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	cashier, err := s.cashierCommand.TrashedCashier(ctx, cashierID)

	if err != nil {
		return s.errorHandler.HandleTrashedCashierError(err, method, "FAILED_TRASHED_CASHIER", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCashierResponseDeleteAt(cashier)

	logSuccess("Successfully trashed cashier", zap.Bool("success", true))

	return so, nil
}

func (s *cashierCommandService) RestoreCashier(ctx context.Context, cashierID int) (*response.CashierResponseDeleteAt, *response.ErrorResponse) {
	const method = "RestoreCashier"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	cashier, err := s.cashierCommand.RestoreCashier(ctx, cashierID)

	if err != nil {
		return s.errorHandler.HandleRestoreCashierError(err, method, "FAILED_RESTORE_CASHIER", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCashierResponseDeleteAt(cashier)

	logSuccess("Successfully restored cashier", zap.Bool("success", true))

	return so, nil
}

func (s *cashierCommandService) DeleteCashierPermanent(ctx context.Context, cashierID int) (bool, *response.ErrorResponse) {
	const method = "DeleteCashierPermanent"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	success, err := s.cashierCommand.DeleteCashierPermanent(ctx, cashierID)

	if err != nil {
		return s.errorHandler.HandleDeleteCashierPermanentError(err, method, "FAILED_DELETE_CASHIER_PERMANENT", span, &status, zap.Error(err))
	}

	logSuccess("Successfully deleted cashier permanently", zap.Bool("success", success))

	return success, nil
}

func (s *cashierCommandService) RestoreAllCashier(ctx context.Context) (bool, *response.ErrorResponse) {
	const method = "RestoreAllCashier"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	success, err := s.cashierCommand.RestoreAllCashier(ctx)

	if err != nil {
		return s.errorHandler.HandleRestoreAllCashierError(err, method, "FAILED_RESTORE_ALL_CASHIERS", span, &status, zap.Error(err))
	}

	logSuccess("Successfully restored all trashed cashiers", zap.Bool("success", success))

	return success, nil
}

func (s *cashierCommandService) DeleteAllCashierPermanent(ctx context.Context) (bool, *response.ErrorResponse) {
	const method = "DeleteAllCashierPermanent"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	success, err := s.cashierCommand.DeleteAllCashierPermanent(ctx)

	if err != nil {
		return s.errorHandler.HandleDeleteAllCashierPermanentError(err, method, "FAILED_DELETE_ALL_CASHIERS_PERMANENT", span, &status, zap.Error(err))
	}

	logSuccess("Successfully deleted all trashed cashiers", zap.Bool("success", success))

	return success, nil
}

func (s *cashierCommandService) startTracingAndLogging(ctx context.Context, method string, attrs ...attribute.KeyValue) (
	context.Context,
	trace.Span,
	func(string),
	string,
	func(string, ...zap.Field),
) {
	start := time.Now()
	status := "success"

	ctx, span := s.trace.Start(ctx, method)

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	span.AddEvent("Start: " + method)

	s.logger.Info("Start: " + method)

	end := func(status string) {
		s.recordMetrics(method, status, start)
		code := codes.Ok
		if status != "success" {
			code = codes.Error
		}
		span.SetStatus(code, status)
		span.End()
	}

	logSuccess := func(msg string, fields ...zap.Field) {
		span.AddEvent(msg)
		s.logger.Debug(msg, fields...)
	}

	return ctx, span, end, status, logSuccess
}

func (s *cashierCommandService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
