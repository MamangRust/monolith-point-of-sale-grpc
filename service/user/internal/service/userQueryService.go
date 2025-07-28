package service

import (
	"context"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/user_errors"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
	"github.com/MamangRust/monolith-point-of-sale-user/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-user/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-user/internal/repository"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type userQueryService struct {
	errorhandler        errorhandler.UserQueryError
	mencache            mencache.UserQueryCache
	trace               trace.Tracer
	userQueryRepository repository.UserQueryRepository
	logger              logger.LoggerInterface
	mapping             response_service.UserResponseMapper
	requestCounter      *prometheus.CounterVec
	requestDuration     *prometheus.HistogramVec
}

func NewUserQueryService(
	errorhandler errorhandler.UserQueryError,
	mencache mencache.UserQueryCache,
	userQueryRepository repository.UserQueryRepository, logger logger.LoggerInterface, mapping response_service.UserResponseMapper) *userQueryService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_query_service_request_total",
			Help: "Total number of requests to the UserQueryService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_query_service_request_duration_seconds",
			Help:    "Histogram of request durations for the UserQueryService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &userQueryService{
		errorhandler:        errorhandler,
		mencache:            mencache,
		trace:               otel.Tracer("user-query-service"),
		userQueryRepository: userQueryRepository,
		logger:              logger,
		mapping:             mapping,
		requestCounter:      requestCounter,
		requestDuration:     requestDuration,
	}
}

func (s *userQueryService) FindAll(ctx context.Context, req *requests.FindAllUsers) ([]*response.UserResponse, *int, *response.ErrorResponse) {
	const method = "FindAll"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedUsersCache(ctx, req); found {
		logSuccess("Successfully fetched user records from cache", zap.Int("totalRecords", *total), zap.Int("page", page), zap.Int("pageSize", pageSize))

		return data, total, nil
	}

	users, totalRecords, err := s.userQueryRepository.FindAllUsers(ctx, req)
	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationError(err, method, "FAILED_FIND_ALL_USERS", span, &status, zap.Error(err))
	}

	userResponses := s.mapping.ToUsersResponse(users)

	s.mencache.SetCachedUsersCache(ctx, req, userResponses, totalRecords)

	logSuccess("Successfully fetched all user records", zap.Int("totalRecords", *totalRecords), zap.Int("page", page), zap.Int("pageSize", pageSize))

	return userResponses, totalRecords, nil
}

func (s *userQueryService) FindByID(ctx context.Context, id int) (*response.UserResponse, *response.ErrorResponse) {
	const method = "FindByID"

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("user.id", id))

	defer func() {
		end(status)
	}()

	user, err := s.userQueryRepository.FindById(ctx, id)

	if err != nil {
		return s.errorhandler.HandleRepositorySingleError(err, method, "FAILED_FIND_USER", span, &status, user_errors.ErrUserNotFoundRes, zap.Int("user.id", id), zap.Error(err))
	}

	userRes := s.mapping.ToUserResponse(user)

	s.mencache.SetCachedUserCache(ctx, userRes)

	logSuccess("Successfully fetched user record", zap.Int("user.id", id))

	return userRes, nil
}

func (s *userQueryService) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*response.UserResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByActive"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedUserActiveCache(ctx, req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	users, totalRecords, err := s.userQueryRepository.FindByActive(ctx, req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, "FindByActive", "FAILED_FIND_ACTIVE_USERS", span, &status, zap.Error(err))
	}

	so := s.mapping.ToUsersResponseDeleteAt(users)

	s.mencache.SetCachedUserActiveCache(ctx, req, so, totalRecords)

	logSuccess("Successfully fetched active users", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return so, totalRecords, nil
}

func (s *userQueryService) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*response.UserResponseDeleteAt, *int, *response.ErrorResponse) {
	const method = "FindByTrashed"

	page, pageSize := s.normalizePagination(req.Page, req.PageSize)
	search := req.Search

	ctx, span, end, status, logSuccess := s.startTracingAndLogging(ctx, method, attribute.Int("page", page), attribute.Int("pageSize", pageSize), attribute.String("search", search))

	defer func() {
		end(status)
	}()

	if data, total, found := s.mencache.GetCachedUserTrashedCache(ctx, req); found {
		logSuccess("Data found in cache", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

		return data, total, nil
	}

	users, totalRecords, err := s.userQueryRepository.FindByTrashed(ctx, req)

	if err != nil {
		return s.errorhandler.HandleRepositoryPaginationDeleteAtError(err, "FindByTrashed", "FAILED_FIND_TRASHED_USERS", span, &status, zap.Error(err))
	}

	so := s.mapping.ToUsersResponseDeleteAt(users)

	s.mencache.SetCachedUserTrashedCache(ctx, req, so, totalRecords)

	logSuccess("Successfully fetched trashed users", zap.Int("page", page), zap.Int("pageSize", pageSize), zap.String("search", search))

	return so, totalRecords, nil
}

func (s *userQueryService) startTracingAndLogging(ctx context.Context, method string, attrs ...attribute.KeyValue) (
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

	return ctx, span, end, status, logSuccess
}

func (s *userQueryService) normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func (s *userQueryService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
}
