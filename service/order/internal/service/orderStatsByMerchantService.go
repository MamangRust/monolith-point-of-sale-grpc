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

type orderStatsByMerchantService struct {
	ctx                            context.Context
	mencache                       mencache.OrderStatsByMerchantCache
	errorhandler                   errorhandler.OrderStatsByMerchantError
	trace                          trace.Tracer
	orderStatsByMerchantRepository repository.OrderStatByMerchantRepository
	mapping                        response_service.OrderResponseMapper
	logger                         logger.LoggerInterface
	requestCounter                 *prometheus.CounterVec
	requestDuration                *prometheus.HistogramVec
}

func NewOrderStatsByMerchantService(
	ctx context.Context,
	mencache mencache.OrderStatsByMerchantCache,
	errorhandler errorhandler.OrderStatsByMerchantError,
	orderStatsByMerchantRepository repository.OrderStatByMerchantRepository,
	logger logger.LoggerInterface,
	mapping response_service.OrderResponseMapper,
) *orderStatsByMerchantService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_stats_by_merchant_service_request_count",
			Help: "Total number of requests to the OrderStatsByMerchantService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_stats_by_merchant_service_request_duration",
			Help:    "Histogram of request durations for the OrderStatsByMerchantService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &orderStatsByMerchantService{
		ctx:                            ctx,
		mencache:                       mencache,
		errorhandler:                   errorhandler,
		trace:                          otel.Tracer("order-stats-by-merchant-service"),
		orderStatsByMerchantRepository: orderStatsByMerchantRepository,
		logger:                         logger,
		mapping:                        mapping,
		requestCounter:                 requestCounter,
		requestDuration:                requestDuration,
	}
}

func (s *orderStatsByMerchantService) FindMonthlyTotalRevenueByMerchant(req *requests.MonthTotalRevenueMerchant) ([]*response.OrderMonthlyTotalRevenueResponse, *response.ErrorResponse) {
	const method = "FindMonthlyTotalRevenueByMerchant"

	merchantId := req.MerchantID
	year := req.Year
	month := req.Month

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("merchant.id", merchantId), attribute.Int("year", year), attribute.Int("month", month))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetMonthlyTotalRevenueByMerchantCache(req); found {
		logSuccess("Successfully fetched monthly total revenue from cache", zap.Int("year", year), zap.Int("month", month))

		return data, nil
	}

	res, err := s.orderStatsByMerchantRepository.GetMonthlyTotalRevenueByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleMonthTotalRevenueByMerchantError(err, method, "FAILED_FIND_MONTHLY_TOTAL_REVENUE_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderMonthlyTotalRevenues(res)
	s.mencache.SetMonthlyTotalRevenueByMerchantCache(req, so)

	logSuccess("Successfully fetched monthly total revenue", zap.Int("year", year), zap.Int("month", month))

	return so, nil
}

func (s *orderStatsByMerchantService) FindYearlyTotalRevenueByMerchant(req *requests.YearTotalRevenueMerchant) ([]*response.OrderYearlyTotalRevenueResponse, *response.ErrorResponse) {
	const method = "FindYearlyTotalRevenueByMerchant"

	year := req.Year
	merchantId := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("merchant.id", merchantId), attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetYearlyTotalRevenueByMerchantCache(req); found {
		logSuccess("Successfully fetched yearly total revenue from cache", zap.Int("year", year))

		return data, nil
	}

	res, err := s.orderStatsByMerchantRepository.GetYearlyTotalRevenueByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleYearTotalRevenueByMerchantError(err, method, "FAILED_FIND_YEARLY_TOTAL_REVENUE_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderYearlyTotalRevenues(res)
	s.mencache.SetYearlyTotalRevenueByMerchantCache(req, so)

	logSuccess("Successfully fetched yearly total revenue", zap.Int("year", year))

	return so, nil
}

func (s *orderStatsByMerchantService) FindMonthlyOrderByMerchant(req *requests.MonthOrderMerchant) ([]*response.OrderMonthlyResponse, *response.ErrorResponse) {
	const method = "FindMonthlyOrderByMerchant"

	year := req.Year
	merchant_id := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("merchant.id", merchant_id), attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetMonthlyOrderByMerchantCache(req); found {
		logSuccess("Successfully fetched monthly orders from cache", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

		return data, nil
	}

	res, err := s.orderStatsByMerchantRepository.GetMonthlyOrderByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleMonthOrderStatsByMerchantError(err, method, "FAILED_FIND_MONTHLY_ORDER_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderMonthlyPrices(res)
	s.mencache.SetMonthlyOrderByMerchantCache(req, so)

	logSuccess("Successfully fetched monthly orders", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

	return so, nil
}

func (s *orderStatsByMerchantService) FindYearlyOrderByMerchant(req *requests.YearOrderMerchant) ([]*response.OrderYearlyResponse, *response.ErrorResponse) {
	const method = "FindYearlyOrderByMerchant"

	year := req.Year
	merchant_id := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("merchant.id", merchant_id), attribute.Int("year", year))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetYearlyOrderByMerchantCache(req); found {
		logSuccess("Successfully fetched yearly orders from cache", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

		return data, nil
	}

	res, err := s.orderStatsByMerchantRepository.GetYearlyOrderByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleYearOrderStatsByMerchantError(err, "FindYearlyOrderByMerchant", "FAILED_FIND_YEARLY_ORDER_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToOrderYearlyPrices(res)
	s.mencache.SetYearlyOrderByMerchantCache(req, so)

	logSuccess("Successfully fetched yearly orders", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

	return so, nil
}

func (s *orderStatsByMerchantService) startTracingAndLogging(method string, attrs ...attribute.KeyValue) (
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

func (s *orderStatsByMerchantService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
