package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-product/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-product/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-product/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/product_errors"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type productQueryService struct {
	ctx                    context.Context
	errorhandler           errorhandler.ProductQueryError
	mencache               mencache.ProductQueryCache
	trace                  trace.Tracer
	productQueryRepository repository.ProductQueryRepository
	mapping                response_service.ProductResponseMapper
	logger                 logger.LoggerInterface
	requestCounter         *prometheus.CounterVec
	requestDuration        *prometheus.HistogramVec
}

func NewProductQueryService(
	ctx context.Context,
	errorhandler errorhandler.ProductQueryError,
	mencache mencache.ProductQueryCache,
	productQueryRepository repository.ProductQueryRepository,
	mapping response_service.ProductResponseMapper,
	logger logger.LoggerInterface,
) *productQueryService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_query_service_requests_total",
			Help: "Total number of requests to the ProductQueryService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_query_service_request_duration_seconds",
			Help:    "Histogram of request durations for the ProductQueryService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &productQueryService{
		ctx:                    ctx,
		errorhandler:           errorhandler,
		mencache:               mencache,
		trace:                  otel.Tracer("product-query-service"),
		productQueryRepository: productQueryRepository,
		mapping:                mapping,
		logger:                 logger,
		requestCounter:         requestCounter,
		requestDuration:        requestDuration,
	}
}

func (s *productQueryService) FindAll(req *requests.FindAllProducts) ([]*response.ProductResponse, *int, *response.ErrorResponse) {
	const method = "FindAll"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedProducts(req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))
		return data, total, nil
	}

	products, totalRecords, err := s.productQueryRepository.FindAllProducts(req)
	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationError(
			err, method, "FAILED_FIND_PRODUCTS", span, &status, zap.Error(err),
		)
	}

	result := s.mapping.ToProductsResponse(products)
	s.mencache.SetCachedProducts(req, result, totalRecords)

	logSuccess("Successfully fetched all products", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return result, totalRecords, nil
}

func (s *productQueryService) FindByMerchant(req *requests.ProductByMerchantRequest) ([]*response.ProductResponse, *int, *response.ErrorResponse) {
	const method = "FindByMerchant"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search
	merchantID := req.MerchantID

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search), attribute.Int("merchant.id", merchantID))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedProductsByMerchant(req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search), zap.Int("merchant.id", merchantID))
		return data, total, nil
	}

	products, totalRecords, err := s.productQueryRepository.FindByMerchant(req)
	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationError(
			err, method, "FAILED_FIND_PRODUCTS_BY_MERCHANT", span, &status, zap.Error(err),
		)
	}

	result := s.mapping.ToProductsResponse(products)
	s.mencache.SetCachedProductsByMerchant(req, result, totalRecords)

	logSuccess("Successfully fetched all products by merchant", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search), zap.Int("merchant.id", merchantID))

	return result, totalRecords, nil
}

func (s *productQueryService) FindByCategory(req *requests.ProductByCategoryRequest) ([]*response.ProductResponse, *int, *response.ErrorResponse) {
	const method = "FindByCategory"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search
	categoryName := req.CategoryName

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search), attribute.String("category.name", categoryName))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedProductsByCategory(req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search), zap.String("category.name", categoryName))
		return data, total, nil
	}

	products, totalRecords, err := s.productQueryRepository.FindByCategory(req)
	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationError(
			err, method, "FAILED_FIND_PRODUCTS_BY_CATEGORY", span, &status, zap.Error(err),
		)
	}

	result := s.mapping.ToProductsResponse(products)
	s.mencache.SetCachedProductsByCategory(req, result, totalRecords)

	logSuccess("Successfully fetched all products by category", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search), zap.String("category.name", categoryName))

	return result, totalRecords, nil
}

func (s *productQueryService) FindById(productID int) (*response.ProductResponse, *response.ErrorResponse) {
	const method = "FindById"

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("product.id", productID))

	defer func() {
		end(status)
	}()

	if data, found := s.mencache.GetCachedProduct(productID); found {
		logSuccess("Data found in cache", zap.Int("product.id", productID))
		return data, nil
	}

	product, err := s.productQueryRepository.FindById(productID)
	if err != nil {
		return errorhandler.HandleRepositorySingleError[*response.ProductResponse](
			s.logger, err, method, "FAILED_FIND_PRODUCT_BY_ID", span, &status,
			product_errors.ErrFailedFindProductById, zap.Error(err),
		)
	}

	so := s.mapping.ToProductResponse(product)
	s.mencache.SetCachedProduct(so)

	logSuccess("Successfully fetched product by id", zap.Int("product.id", productID))

	return so, nil
}

func (s *productQueryService) FindByActive(req *requests.FindAllProducts) ([]*response.ProductResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByActive"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedProductActive(req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	products, totalRecords, err := s.productQueryRepository.FindByActive(req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, method, "FAILED_FIND_PRODUCTS_ACTIVE", span, &status, product_errors.ErrFailedFindProductsByActive, zap.Error(err))
	}

	so := s.mapping.ToProductsResponseDeleteAt(products)

	s.mencache.SetCachedProductActive(req, so, totalRecords)

	logSuccess("Successfully fetched all products", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return so, totalRecords, nil
}

func (s *productQueryService) FindByTrashed(req *requests.FindAllProducts) ([]*response.ProductResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByTrashed"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedProductTrashed(req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	products, totalRecords, err := s.productQueryRepository.FindByTrashed(req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, method, "FAILED_FIND_PRODUCTS_TRASHED", span, &status, product_errors.ErrFailedFindProductsByTrashed, zap.Error(err))
	}

	so := s.mapping.ToProductsResponseDeleteAt(products)

	s.mencache.SetCachedProductTrashed(req, so, totalRecords)

	logSuccess("Successfully fetched all products", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return so, totalRecords, nil
}

func (s *productQueryService) startTracingAndLogging(method string, attrs ...attribute.KeyValue) (
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

func (s *productQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func (s *productQueryService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
