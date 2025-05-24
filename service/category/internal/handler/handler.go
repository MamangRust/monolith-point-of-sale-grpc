package handler

import "github.com/MamangRust/monolith-point-of-sale-category/internal/service"

type Deps struct {
	Service service.Service
}

type Handler struct {
	Category CategoryHandleGrpc
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		Category: NewCategoryHandleGrpc(deps.Service),
	}
}
