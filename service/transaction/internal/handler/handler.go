package handler

import "github.com/MamangRust/monolith-point-of-sale-transacton/internal/service"

type Deps struct {
	Service *service.Service
}

type Handler struct {
	Transaction TransactionHandleGrpc
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Transaction: NewTransactionHandleGrpc(deps.Service),
	}
}
