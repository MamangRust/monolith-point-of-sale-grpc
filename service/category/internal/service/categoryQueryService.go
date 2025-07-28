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
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/category_errors"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type categoryQueryService struct {
	errorhandler            errorhandler.CategoryQueryError
	mencache                mencache.CategoryQueryCache
	trace                   trace.Tracer
	categoryQueryRepository repository.CategoryQueryRepository
	logger                  logger.LoggerInterface
	mapping                 response_service.CategoryResponseMapper
	requestCounter          *prometheus.CounterVec
	requestDuration         *prometheus.HistogramVec
}

func NewCategoryQueryService(
	errorhandler errorhandler.CategoryQueryError,
	mencache mencache.CategoryQueryCache,
	categoryQueryRepository repository.CategoryQueryRepository, logger logger.LoggerInterface, mapping response_service.CategoryResponseMapper) *categoryQueryService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "category_query_service_request_total",
			Help: "Total number of requests to the CategoryQueryService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "category_query_service_request_duration_seconds",
			Help:    "Duration of requests to the CategoryQueryService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &categoryQueryService{
		errorhandler:            errorhandler,
		mencache:                mencache,
		trace:                   otel.Tracer("category-query-service"),
		categoryQueryRepository: categoryQueryRepository,
		logger:                  logger,
		mapping:                 mapping,
		requestCounter:          requestCounter,
		requestDuration:         requestDuration,
	}
}

func (s *categoryQueryService) FindAll(ctx context.Context, req *requests.FindAllCategory) ([]*response.CategoryResponse, *int, *response.ErrorResponse) {
	const method = "FindAll"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedCategoriesCache(ctx, req); found {
		logSuccess("Successfully fetched categories from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))
		return data, total, nil
	}

	category, totalRecords, err := s.categoryQueryRepository.FindAllCategory(ctx, req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationError(err, method, "FAILED_FIND_ALL_CATEGORY", span, &status, zap.Error(err))
	}

	categoriesResponse := s.mapping.ToCategorysResponse(category)

	s.mencache.SetCachedCategoriesCache(ctx, req, categoriesResponse, totalRecords)

	logSuccess("Successfully fetched categories", zap.Int("totalRecords", *totalRecords), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return categoriesResponse, totalRecords, nil
}

func (s *categoryQueryService) FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*response.CategoryResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByActive"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedCategoryActiveCache(ctx, req); found {
		logSuccess("Successfully fetched categories from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))

		return data, total, nil
	}

	category, totalRecords, err := s.categoryQueryRepository.FindByActive(ctx, req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, method, "FAILED_FIND_BY_ACTIVE_CATEGORY", span, &status, category_errors.ErrFailedFindActiveCategories, zap.Error(err))
	}

	so := s.mapping.ToCategoryResponsesDeleteAt(category)

	s.mencache.SetCachedCategoryActiveCache(ctx, req, so, totalRecords)

	logSuccess("Successfully fetched categories", zap.Int("totalRecords", *totalRecords), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return so, totalRecords, nil
}

func (s *categoryQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*response.CategoryResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByTrashed"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedCategoryTrashedCache(ctx, req); found {
		logSuccess("Successfully fetched categories from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))

		return data, total, nil
	}

	categories, totalRecords, err := s.categoryQueryRepository.FindByTrashed(ctx, req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, method, "FAILED_FIND_BY_TRASHED_CATEGORY", span, &status, category_errors.ErrFailedFindTrashedCategories, zap.Error(err))
	}

	so := s.mapping.ToCategoryResponsesDeleteAt(categories)

	s.mencache.SetCachedCategoryTrashedCache(ctx, req, so, totalRecords)

	logSuccess("Successfully fetched categories", zap.Int("totalRecords", *totalRecords), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return so, totalRecords, nil
}

func (s *categoryQueryService) FindById(ctx context.Context, category_id int) (*response.CategoryResponse, *response.ErrorResponse) {
	const method = "FindById"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("category.id", category_id))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedCategoryCache(ctx, category_id); found {
		logSuccess("Successfully fetched category from cache", zap.Int("category.id", category_id))

		return data, nil
	}

	category, err := s.categoryQueryRepository.FindById(ctx, category_id)

	if err != nil {
		return s.errorhandler.HandleRepositorySingleError(err, method, "FAILED_FIND_CATEGORY_BY_ID", span, &status, category_errors.ErrFailedFindCategoryById, zap.Error(err))
	}

	so := s.mapping.ToCategoryResponse(category)

	s.mencache.SetCachedCategoryCache(ctx, so)

	logSuccess("Successfully fetched category", zap.Int("category.id", category_id))

	return so, nil
}

func (s *categoryQueryService) startTracingAndLogging(ctx context.Context, method string, attrs ...attribute.KeyValue) (
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

func (s *categoryQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func (s *categoryQueryService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
