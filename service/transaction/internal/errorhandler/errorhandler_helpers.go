package errorhandler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	traceunic "github.com/MamangRust/monolith-point-of-sale-pkg/trace_unic"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func handleErrorInsufficientBalance[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"Insufficient balance", span, status, errorResp,
		fields...,
	)
}

func handleErrorCannotModified[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"Payment method Cannot modified", span, status, errorResp,
		fields...,
	)
}

func handleErrorInvalidQuantityOrderItem[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"Invalid quantity order item", span, status, errorResp,
		fields...,
	)
}

func handleErrorTemplate[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix, errorMessage string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	traceID := traceunic.GenerateTraceID(tracePrefix)
	logMsg := fmt.Sprintf("%s in %s", errorMessage, method)

	allFields := append(fields,
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	logger.Error(logMsg, allFields...)

	span.SetAttributes(attribute.String("trace.id", traceID))
	span.RecordError(err)
	span.AddEvent(logMsg)
	span.SetStatus(codes.Error, logMsg)

	*status = fmt.Sprintf("%s_error_%s", toSnakeCase(method), toSnakeCase(errorMessage))

	var zero T
	return zero, errorResp
}

func handleErrorRepository[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"Repository error", span, status, errorResp, fields...,
	)
}

func handleErrorPagination[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *int, *response.ErrorResponse) {
	result, errResp := handleErrorRepository[T](
		logger, err, method, tracePrefix, span, status, errorResp, fields...,
	)
	return result, nil, errResp
}

func HandleErrorJSONMarshal[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"JSON marshal error", span, status, errorResp, fields...,
	)
}

func HandleErrorKafkaSend[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"Kafka send error", span, status, errorResp, fields...,
	)
}

func HandleRepositorySingleError[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorRepository[T](logger, err, method, tracePrefix, span, status, errorResp, fields...)
}

func HandleCannotOrderItem[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](logger, err, method, tracePrefix, "Cannot order item", span, status, errorResp, fields...)
}

func HandleInvalidQuantityOrderItem[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](logger, err, method, tracePrefix, "Invalid quantity order item", span, status, errorResp, fields...)
}

func HandleErrorGenerateRandomString[T any](
	logger logger.LoggerInterface,
	err error,
	method, tracePrefix string,
	span trace.Span,
	status *string,
	errorResp *response.ErrorResponse,
	fields ...zap.Field,
) (T, *response.ErrorResponse) {
	return handleErrorTemplate[T](
		logger, err, method, tracePrefix,
		"Generate random string error", span, status, errorResp, fields...,
	)
}

func toSnakeCase(s string) string {
	re := regexp.MustCompile("([a-z])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}
