package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-order/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-order/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-order/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type orderStatsService struct {
	ctx                  context.Context
	errorhandler         errorhandler.OrderStatsError
	mencache             mencache.OrderStatsCache
	trace                trace.Tracer
	orderStatsRepository repository.OrderStatsRepository
	logger               logger.LoggerInterface
	mapping              response_service.OrderResponseMapper
	requestCounter       *prometheus.CounterVec
	requestDuration      *prometheus.HistogramVec
}

func NewOrderStatsService(
	ctx context.Context,
	errorhandler errorhandler.OrderStatsError,
	mencache mencache.OrderStatsCache,
	orderStatsRepository repository.OrderStatsRepository,
	logger logger.LoggerInterface,
	mapping response_service.OrderResponseMapper,
) *orderStatsService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_stats_service_request_count",
			Help: "Total number of requests to the OrderStatsService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_stats_service_request_duration",
			Help:    "Histogram of request durations for the OrderStatsService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &orderStatsService{
		ctx:                  ctx,
		errorhandler:         errorhandler,
		mencache:             mencache,
		trace:                otel.Tracer("order-stats-service"),
		orderStatsRepository: orderStatsRepository,
		logger:               logger,
		mapping:              mapping,
		requestCounter:       requestCounter,
		requestDuration:      requestDuration,
	}
}

func (s *orderStatsService) FindMonthlyTotalRevenue(req *requests.MonthTotalRevenue) ([]*response.OrderMonthlyTotalRevenueResponse, *response.ErrorResponse) {
	const method = "FindMonthlyTotalRevenue"

	year := req.Year
	month := req.Month

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year), attribute.Int("month", month))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetMonthlyTotalRevenueCache(req); found {
		logSuccess("Successfully fetched monthly total revenue from cache", zap.Int("year", year), zap.Int("month", month))

		return data, nil
	}

	res, err := s.orderStatsRepository.GetMonthlyTotalRevenue(req)

	if err != nil {
		return s.errorhandler.HandleMonthlyTotalRevenueError(err, method, "FAILED_FIND_MONTHLY_TOTAL_REVENUE", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderMonthlyTotalRevenues(res)

	s.mencache.SetMonthlyTotalRevenueCache(req, so)

	logSuccess("Successfully fetched monthly total revenue", zap.Int("year", year), zap.Int("month", month))

	return so, nil
}

func (s *orderStatsService) FindYearlyTotalRevenue(year int) ([]*response.OrderYearlyTotalRevenueResponse, *response.ErrorResponse) {
	const method = "FindYearlyTotalRevenue"

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetYearlyTotalRevenueCache(year); found {
		logSuccess("Successfully fetched yearly total revenue from cache", zap.Int("year", year))

		return data, nil
	}

	res, err := s.orderStatsRepository.GetYearlyTotalRevenue(year)

	if err != nil {
		return s.errorhandler.HandleYearlyTotalRevenueError(err, method, "FAILED_FIND_YEARLY_TOTAL_REVENUE", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderYearlyTotalRevenues(res)

	s.mencache.SetYearlyTotalRevenueCache(year, so)

	logSuccess("Successfully fetched yearly total revenue", zap.Int("year", year))

	return so, nil
}

func (s *orderStatsService) FindMonthlyOrder(year int) ([]*response.OrderMonthlyResponse, *response.ErrorResponse) {
	const method = "FindMonthlyOrder"

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetMonthlyOrderCache(year); found {
		logSuccess("Successfully fetched monthly orders from cache", zap.Int("year", year))

		return data, nil
	}

	res, err := s.orderStatsRepository.GetMonthlyOrder(year)

	if err != nil {
		return s.errorhandler.HandleMonthOrderStatsError(err, method, "FAILED_FIND_MONTHLY_ORDER", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderMonthlyPrices(res)

	s.mencache.SetMonthlyOrderCache(year, so)

	logSuccess("Successfully fetched monthly orders", zap.Int("year", year))

	return so, nil
}

func (s *orderStatsService) FindYearlyOrder(year int) ([]*response.OrderYearlyResponse, *response.ErrorResponse) {
	const method = "FindYearlyOrder"

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetYearlyOrderCache(year); found {
		logSuccess("Successfully fetched yearly orders from cache", zap.Int("year", year))

		return data, nil
	}

	res, err := s.orderStatsRepository.GetYearlyOrder(year)

	if err != nil {
		return s.errorhandler.HandleYearOrderStatsError(err, method, "FAILED_FIND_YEARLY_ORDER", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderYearlyPrices(res)

	s.mencache.SetYearlyOrderCache(year, so)

	logSuccess("Successfully fetched yearly orders", zap.Int("year", year))

	return so, nil
}

func (s *orderStatsService) startTracingAndLogging(method string, attrs ...attribute.KeyValue) (
	trace.Span,
	func(string),
	string,
	func(string, ...zap.Field),
) {
	start := time.Now()
	status := "success"

	_, span := s.trace.Start(s.ctx, method)

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	span.AddEvent("Start: " + method)

	s.logger.Debug("Start: " + method)

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

	return span, end, status, logSuccess
}

func (s *orderStatsService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
