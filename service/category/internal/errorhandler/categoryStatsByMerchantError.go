package errorhandler

import (
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/category_errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type categoryStatsByMerchantError struct {
	logger logger.LoggerInterface
}

func NewCategoryStatsByMerchantError(logger logger.LoggerInterface) *categoryStatsByMerchantError {
	return &categoryStatsByMerchantError{logger: logger}
}

func (c *categoryStatsByMerchantError) HandleMonthTotalPriceError(err error, method, tracePrefix string, span trace.Span, status *string, fields ...zap.Field) ([]*response.CategoriesMonthlyTotalPriceResponse, *response.ErrorResponse) {
	return handleErrorRepository[[]*response.CategoriesMonthlyTotalPriceResponse](c.logger, err, method, tracePrefix, span, status, category_errors.ErrFailedFindMonthlyTotalPrice, fields...)
}

func (c *categoryStatsByMerchantError) HandleYearTotalPriceError(err error, method, tracePrefix string, span trace.Span, status *string, fields ...zap.Field) ([]*response.CategoriesYearlyTotalPriceResponse, *response.ErrorResponse) {
	return handleErrorRepository[[]*response.CategoriesYearlyTotalPriceResponse](c.logger, err, method, tracePrefix, span, status, category_errors.ErrFailedFindYearlyTotalPrice, fields...)
}

func (c *categoryStatsByMerchantError) HandleMonthPrice(err error, method, tracePrefix string, span trace.Span, status *string, fields ...zap.Field) ([]*response.CategoryMonthPriceResponse, *response.ErrorResponse) {
	return handleErrorRepository[[]*response.CategoryMonthPriceResponse](c.logger, err, method, tracePrefix, span, status, category_errors.ErrFailedFindMonthPrice, fields...)
}

func (c *categoryStatsByMerchantError) HandleYearPrice(err error, method, tracePrefix string, span trace.Span, status *string, fields ...zap.Field) ([]*response.CategoryYearPriceResponse, *response.ErrorResponse) {
	return handleErrorRepository[[]*response.CategoryYearPriceResponse](c.logger, err, method, tracePrefix, span, status, category_errors.ErrFailedFindYearPrice, fields...)
}
