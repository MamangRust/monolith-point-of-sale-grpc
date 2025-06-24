package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-auth/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-auth/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-auth/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-pkg/email"
	"github.com/MamangRust/monolith-point-of-sale-pkg/hash"
	"github.com/MamangRust/monolith-point-of-sale-pkg/kafka"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-pkg/randomstring"
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

type registerService struct {
	ctx               context.Context
	trace             trace.Tracer
	errohandler       errorhandler.RegisterErrorHandler
	errorPassword     errorhandler.PasswordErrorHandler
	errorRandomString errorhandler.RandomStringErrorHandler
	errorMarshal      errorhandler.MarshalErrorHandler
	errorKafka        errorhandler.KafkaErrorHandler
	mencache          mencache.RegisterCache
	user              repository.UserRepository
	role              repository.RoleRepository
	userRole          repository.UserRoleRepository
	hash              hash.HashPassword
	kafka             *kafka.Kafka
	logger            logger.LoggerInterface
	mapping           response_service.UserResponseMapper
	requestCounter    *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
}

func NewRegisterService(ctx context.Context,
	errorhandler errorhandler.RegisterErrorHandler,
	errorPassword errorhandler.PasswordErrorHandler,
	errorRandomString errorhandler.RandomStringErrorHandler,
	errorMarshal errorhandler.MarshalErrorHandler,
	errorKafka errorhandler.KafkaErrorHandler,
	mencache mencache.RegisterCache,
	user repository.UserRepository, role repository.RoleRepository, userRole repository.UserRoleRepository, hash hash.HashPassword, kafka *kafka.Kafka, logger logger.LoggerInterface, mapping response_service.UserResponseMapper) *registerService {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "register_service_requests_total",
			Help: "Total number of requests to the RegisterService",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "register_service_request_duration_seconds",
			Help:    "Histogram of request durations for the RegisterService",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	prometheus.MustRegister(requestCounter, requestDuration)

	return &registerService{
		ctx:               ctx,
		errorPassword:     errorPassword,
		errohandler:       errorhandler,
		errorRandomString: errorRandomString,
		errorMarshal:      errorMarshal,
		errorKafka:        errorKafka,
		mencache:          mencache,
		trace:             otel.Tracer("register-service"),
		user:              user,
		role:              role,
		userRole:          userRole,
		hash:              hash,
		kafka:             kafka,
		logger:            logger,
		mapping:           mapping,
		requestCounter:    requestCounter,
		requestDuration:   requestDuration,
	}
}

func (s *registerService) Register(request *requests.RegisterRequest) (*response.UserResponse, *response.ErrorResponse) {
	const method = "Register"

	span, end, status, logSuccess := s.startTracingAndLogging(method, attribute.String("email", request.Email))

	defer func() {
		end(status)
	}()

	existingUser, err := s.user.FindByEmail(request.Email)
	if err == nil && existingUser != nil {
		return s.errohandler.HandleFindEmailError(err, "Register", "REGISTER_ERR", span, &status,
			zap.String("email", request.Email), zap.Error(err))
	}

	passwordHash, err := s.hash.HashPassword(request.Password)
	if err != nil {
		return s.errorPassword.HandleHashPasswordError(err, "Register", "REGISTER_ERR", span, &status)
	}
	request.Password = passwordHash

	const defaultRoleName = "Admin Access 1"

	role, err := s.role.FindByName(defaultRoleName)

	if err != nil || role == nil {
		return s.errohandler.HandleFindRoleError(err, "Register", "REGISTER_ERR", span, &status,
			zap.String("role_name", defaultRoleName), zap.Error(err))
	}

	random, err := randomstring.GenerateRandomString(10)
	if err != nil {
		return s.errorRandomString.HandleRandomStringErrorRegister(err, "Register", "REGISTER_ERR", span, &status, zap.Error(err))
	}

	request.VerifiedCode = random
	request.IsVerified = false

	newUser, err := s.user.CreateUser(request)
	if err != nil {
		return s.errohandler.HandleCreateUserError(err, "Register", "REGISTER_ERR", span, &status, zap.Error(err))
	}

	htmlBody := email.GenerateEmailHTML(map[string]string{
		"Title":   "Welcome to SanEdge",
		"Message": "Your account has been successfully created.",
		"Button":  "Login Now",
		"Link":    "https://sanedge.example.com/login?verify_code=" + request.VerifiedCode,
	})

	emailPayload := map[string]any{
		"email":   request.Email,
		"subject": "Welcome to SanEdge",
		"body":    htmlBody,
	}

	payloadBytes, err := json.Marshal(emailPayload)
	if err != nil {
		return s.errorMarshal.HandleMarshalRegisterError(err, "Register", "MARSHAL_ERR", span, &status, zap.Error(err))
	}

	err = s.kafka.SendMessage("email-service-topic-auth-register", strconv.Itoa(newUser.ID), payloadBytes)
	if err != nil {
		return s.errorKafka.HandleSendEmailRegister(err, "Register", "SEND_EMAIL_ERR", span, &status, zap.Error(err))
	}

	_, err = s.userRole.AssignRoleToUser(&requests.CreateUserRoleRequest{
		UserId: newUser.ID,
		RoleId: role.ID,
	})
	if err != nil {
		return s.errohandler.HandleAssignRoleError(err, "Register", "ASSIGN_ROLE_ERR", span, &status, zap.Error(err))
	}

	s.mencache.SetVerificationCodeCache(request.Email, random, 15*time.Minute)

	userResponse := s.mapping.ToUserResponse(newUser)

	logSuccess("User registered successfully",
		zap.String("email", request.Email),
		zap.String("first_name", request.FirstName),
		zap.String("last_name", request.LastName),
	)

	return userResponse, nil
}

func (s *registerService) startTracingAndLogging(method string, attrs ...attribute.KeyValue) (
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
		s.logger.Info(msg, fields...)
	}

	return span, end, status, logSuccess
}

func (s *registerService) recordMetrics(method string, status string, start time.Time) {
	s.requestCounter.WithLabelValues(method, status).Inc()
	s.requestDuration.WithLabelValues(method, status).Observe(time.Since(start).Seconds())
}
