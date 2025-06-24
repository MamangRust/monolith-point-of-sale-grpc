package service

import (
	"context"

	"github.com/MamangRust/monolith-point-of-sale-merchant/internal/errorhandler"
	mencache "github.com/MamangRust/monolith-point-of-sale-merchant/internal/redis"
	"github.com/MamangRust/monolith-point-of-sale-merchant/internal/repository"
	"github.com/MamangRust/monolith-point-of-sale-pkg/kafka"
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	response_service "github.com/MamangRust/monolith-point-of-sale-shared/mapper/response/service"
)

type Service struct {
	MerchantQuery           MerchantQueryService
	MerchantCommand         MerchantCommandService
	MerchantDocumentCommand MerchantDocumentCommandService
	MerchantDocumentQuery   MerchantDocumentQueryService
}

type Deps struct {
	Kafka        *kafka.Kafka
	Ctx          context.Context
	Repositories *repository.Repositories
	ErrorHander  *errorhandler.ErrorHandler
	Mencache     *mencache.Mencache
	Logger       logger.LoggerInterface
}

func NewService(deps *Deps) *Service {
	merchantMapper := response_service.NewMerchantResponseMapper()
	merchantDocument := response_service.NewMerchantDocumentResponseMapper()

	return &Service{
		MerchantQuery:           NewMerchantQueryService(deps.Ctx, deps.ErrorHander.MerchantQueryError, deps.Mencache.MerchantQueryCache, deps.Repositories.MerchantQuery, deps.Logger, merchantMapper),
		MerchantCommand:         NewMerchantCommandService(deps.Kafka, deps.Ctx, deps.ErrorHander.MerchantCommandError, deps.Mencache.MerchantCommandCache, deps.Repositories.UserQuery, deps.Repositories.MerchantQuery, deps.Repositories.MerchantCommand, deps.Logger, merchantMapper),
		MerchantDocumentCommand: NewMerchantDocumentCommandService(deps.Kafka, deps.Ctx, deps.Mencache.MerchantDocumentCommandCache, deps.ErrorHander.MerchantDocumentCommandError, deps.Repositories.MerchantDocumentCommand, deps.Repositories.MerchantQuery, deps.Repositories.UserQuery, deps.Logger, merchantDocument),
		MerchantDocumentQuery:   NewMerchantDocumentQueryService(deps.Ctx, deps.ErrorHander.MerchantDocumentQueryError, deps.Mencache.MerchantDocumentQueryCache, deps.Repositories.MerchantDocumentQuery, deps.Logger, merchantDocument),
	}
}
