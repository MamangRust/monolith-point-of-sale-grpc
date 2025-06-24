package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-category/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-category/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-category/internal/repository"
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

type categoryStatsByMerchantService struct {
	ctx                               context.Context
	mencache                          mencache.CategoryStatsByMerchantCache
	errorhandler                      errorhandler.CategoryStatsByMerchantError
	trace                             trace.Tracer
	categoryStatsByMerchantRepository repository.CategoryStatsByMerchantRepository
	logger                            logger.LoggerInterface
	mapping                           response_service.CategoryResponseMapper
	requestCounter                    *prometheus.CounterVec
	requestDuration                   *prometheus.HistogramVec
}

func NewCategoryStatsByMerchantService(ctx context.Context,
	mencache mencache.CategoryStatsByMerchantCache,
	errorhandler errorhandler.CategoryStatsByMerchantError,
	categoryStatsByMerchantRepository repository.CategoryStatsByMerchantRepository, logger logger.LoggerInterface, mapping response_service.CategoryResponseMapper) *categoryStatsByMerchantService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "category_stats_by_merchant_service_request_total",
			Help: "Total number of requests to the CategoryStatsByMerchantService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "category_stats_by_merchant_service_request_duration_seconds",
			Help:    "Duration of requests to the CategoryStatsByMerchantService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &categoryStatsByMerchantService{
		ctx:                               ctx,
		mencache:                          mencache,
		errorhandler:                      errorhandler,
		trace:                             otel.Tracer("category-stats-by-id-service"),
		categoryStatsByMerchantRepository: categoryStatsByMerchantRepository,
		logger:                            logger,
		mapping:                           mapping,
		requestCounter:                    requestCounter,
		requestDuration:                   requestDuration,
	}
}

func (s *categoryStatsByMerchantService) FindMonthlyTotalPriceByMerchant(req *requests.MonthTotalPriceMerchant) ([]*response.CategoriesMonthlyTotalPriceResponse, *response.ErrorResponse) {
	const method = "FindMonthlyTotalPriceByMerchant"

	year := req.Year
	month := req.Month

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year), attribute.Int("month", month))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedMonthTotalPriceByMerchantCache(req); found {
		logSuccess("Successfully fetched monthly total price by merchant from cache", zap.Int("year", year), zap.Int("month", month))

		return data, nil
	}

	res, err := s.categoryStatsByMerchantRepository.GetMonthlyTotalPriceByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleMonthTotalPriceError(err, method, "FAILED_FIND_MONTHLY_TOTAL_PRICE_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCategoryMonthlyTotalPrices(res)

	s.mencache.SetCachedMonthTotalPriceByMerchantCache(req, so)

	logSuccess("Successfully fetched monthly total price by merchant", zap.Int("year", year), zap.Int("month", month))

	return so, nil
}

func (s *categoryStatsByMerchantService) FindYearlyTotalPriceByMerchant(req *requests.YearTotalPriceMerchant) ([]*response.CategoriesYearlyTotalPriceResponse, *response.ErrorResponse) {
	const method = "FindYearlyTotalPriceByMerchant"

	year := req.Year
	merchantId := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year), attribute.Int("merchant.id", merchantId))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedYearTotalPriceByMerchantCache(req); found {
		logSuccess("Successfully fetched yearly total price by merchant from cache", zap.Int("year", year), zap.Int("merchant.id", merchantId))

		return data, nil
	}

	res, err := s.categoryStatsByMerchantRepository.GetYearlyTotalPricesByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleYearTotalPriceError(err, method, "FAILED_FIND_YEARLY_TOTAL_PRICE_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCategoryYearlyTotalPrices(res)

	s.mencache.SetCachedYearTotalPriceByMerchantCache(req, so)

	logSuccess("Successfully fetched yearly total price by merchant", zap.Int("year", year), zap.Int("merchant.id", merchantId))

	return so, nil
}

func (s *categoryStatsByMerchantService) FindMonthPriceByMerchant(req *requests.MonthPriceMerchant) ([]*response.CategoryMonthPriceResponse, *response.ErrorResponse) {
	const method = "FindMonthPriceByMerchant"

	year := req.Year
	merchant_id := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year), attribute.Int("merchant.id", merchant_id))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedMonthPriceByMerchantCache(req); found {
		logSuccess("Successfully fetched monthly category prices by merchant from cache", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

		return data, nil
	}

	res, err := s.categoryStatsByMerchantRepository.GetMonthPriceByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleMonthPrice(err, method, "FAILED_FIND_MONTH_PRICE_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCategoryMonthlyPrices(res)

	s.mencache.SetCachedMonthPriceByMerchantCache(req, so)

	logSuccess("Successfully fetched monthly category prices by merchant", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

	return so, nil
}

func (s *categoryStatsByMerchantService) FindYearPriceByMerchant(req *requests.YearPriceMerchant) ([]*response.CategoryYearPriceResponse, *response.ErrorResponse) {
	const method = "FindYearPriceByMerchant"

	year := req.Year
	merchant_id := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("year", year), attribute.Int("merchant.id", merchant_id))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedYearPriceByMerchantCache(req); found {
		logSuccess("Successfully fetched yearly category prices by merchant from cache", zap.Int("year", year), zap.Int("merchant.id", merchant_id))
		return data, nil
	}

	res, err := s.categoryStatsByMerchantRepository.GetYearPriceByMerchant(req)

	if err != nil {
		return s.errorhandler.HandleYearPrice(err, method, "FAILED_FIND_YEAR_PRICE_BY_MERCHANT", span, &status, zap.Error(err))
	}

	so := s.mapping.ToCategoryYearlyPrices(res)

	s.mencache.SetCachedYearPriceByMerchantCache(req, so)

	logSuccess("Successfully fetched yearly category prices by merchant", zap.Int("year", year), zap.Int("merchant.id", merchant_id))

	return so, nil
}

func (s *categoryStatsByMerchantService) startTracingAndLogging(method string, attrs ...attribute.KeyValue) (
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

func (s *categoryStatsByMerchantService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
