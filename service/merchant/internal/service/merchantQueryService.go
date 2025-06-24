package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-merchant/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-merchant/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-merchant/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/merchant_errors"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type merchantQueryService struct {
	ctx                     context.Context
	errorhandler            errorhandler.MerchantQueryErrorHandler
	mencache                mencache.MerchantQueryCache
	trace                   trace.Tracer
	merchantQueryRepository repository.MerchantQueryRepository
	logger                  logger.LoggerInterface
	mapping                 response_service.MerchantResponseMapper
	requestCounter          *prometheus.CounterVec
	requestDuration         *prometheus.HistogramVec
}

func NewMerchantQueryService(ctx context.Context,
	errorhandler errorhandler.MerchantQueryErrorHandler,
	mencache mencache.MerchantQueryCache,
	merchantQueryRepository repository.MerchantQueryRepository, logger logger.LoggerInterface, mapping response_service.MerchantResponseMapper) *merchantQueryService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "merchant_query_service_requests_total",
			Help: "Total number of requests to the MerchantQueryService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "merchant_query_service_request_duration_seconds",
			Help:    "Histogram of request durations for the MerchantQueryService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &merchantQueryService{
		ctx:                     ctx,
		errorhandler:            errorhandler,
		mencache:                mencache,
		trace:                   otel.Tracer("merchant-query-service"),
		merchantQueryRepository: merchantQueryRepository,
		logger:                  logger,
		mapping:                 mapping,
		requestCounter:          requestCounter,
		requestDuration:         requestDuration,
	}
}

func (s *merchantQueryService) FindAll(req *requests.FindAllMerchants) ([]*response.MerchantResponse, *int, *response.ErrorResponse) {
	const method = "FindAll"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedMerchants(req); found {
		logSuccess("Successfully fetched merchants from cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	merchants, totalRecords, err := s.merchantQueryRepository.FindAllMerchants(req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationError(err, "FindAll", "FAILED_FIND_ALL_MERCHANTS", span, &status, zap.Error(err))
	}

	merchantResponses := s.mapping.ToMerchantsResponse(merchants)

	s.mencache.SetCachedMerchants(req, merchantResponses, totalRecords)

	logSuccess("Successfully fetched merchants from database", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return merchantResponses, totalRecords, nil
}

func (s *merchantQueryService) FindById(merchant_id int) (*response.MerchantResponse, *response.ErrorResponse) {
	const method = "FindById"

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("merchant.id", merchant_id))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedMerchant(merchant_id); found {
		logSuccess("Successfully fetched merchant from cache", zap.Int("merchant.id", merchant_id))

		return data, nil
	}

	res, err := s.merchantQueryRepository.FindById(merchant_id)

	if err != nil {
		return errorhandler.HandleRepositorySingleError[*response.MerchantResponse](s.logger, err, method, "FAILED_FIND_MERCHANT_BY_ID", span, &status, merchant_errors.ErrFailedFindMerchantById, zap.Int("merchant.id", merchant_id))
	}

	so := s.mapping.ToMerchantResponse(res)

	s.mencache.SetCachedMerchant(so)

	logSuccess("Successfully fetched merchant from database", zap.Int("merchant.id", merchant_id))

	return so, nil
}

func (s *merchantQueryService) FindByActive(req *requests.FindAllMerchants) ([]*response.MerchantResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByActive"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedMerchantActive(req); found {
		logSuccess("Successfully fetched active merchants from cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	merchants, totalRecords, err := s.merchantQueryRepository.FindByActive(req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, method, "FAILED_FIND_ALL_ACTIVE_MERCHANTS", span, &status, merchant_errors.ErrFailedFindMerchantsByActive, zap.Error(err))
	}

	so := s.mapping.ToMerchantsResponseDeleteAt(merchants)

	logSuccess("Successfully fetched active merchants from database", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return so, totalRecords, nil
}

func (s *merchantQueryService) FindByTrashed(req *requests.FindAllMerchants) ([]*response.MerchantResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByTrashed"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedMerchantTrashed(req); found {
		logSuccess("Successfully fetched trashed merchants from cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	merchants, totalRecords, err := s.merchantQueryRepository.FindByTrashed(req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, method, "FAILED_FIND_ALL_TRASHED_MERCHANTS", span, &status, merchant_errors.ErrFailedFindMerchantsByTrashed, zap.Error(err))
	}

	so := s.mapping.ToMerchantsResponseDeleteAt(merchants)

	s.mencache.SetCachedMerchantTrashed(req, so, totalRecords)

	logSuccess("Successfully fetched trashed merchants from database", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return so, totalRecords, nil
}

func (s *merchantQueryService) startTracingAndLogging(method string, attrs ...attribute.KeyValue) (
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

func (s *merchantQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func (s *merchantQueryService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
