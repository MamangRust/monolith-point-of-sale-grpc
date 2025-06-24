package errorhandler

import (
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/user_errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type randomStringError struct {
	logger logger.LoggerInterface
}

func NewRandomStringError(logger logger.LoggerInterface) *randomStringError {
	return &randomStringError{
		logger: logger,
	}
}

func (r randomStringError) HandleRandomStringErrorRegister(err error, method, tracePrefix string, span trace.Span, status *string, fields ...zap.Field) (*response.UserResponse, *response.ErrorResponse) {
	return handleErrorGenerateRandomString[*response.UserResponse](
		r.logger,
		err,
		method,
		tracePrefix,
		span,
		status,
		user_errors.ErrInternalServerError,
		fields...,
	)
}

func (h *randomStringError) HandleRandomStringErrorForgotPassword(err error, method, tracePrefix string, span trace.Span, status *string, fields ...zap.Field) (bool, *response.ErrorResponse) {
	return handleErrorGenerateRandomString[bool](
		h.logger,
		err,
		method,
		tracePrefix,
		span,
		status,
		user_errors.ErrInternalServerError,
		fields...,
	)
}
