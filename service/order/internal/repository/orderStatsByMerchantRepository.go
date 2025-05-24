package repository

import (
	"context"
	"database/sql"
	"time"

	db "github.com/MamangRust/monolith-point-of-sale-pkg/database/schema"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/record"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/order_errors"
	recordmapper "github.com/MamangRust/monolith-point-of-sale-shared/mapper/record"
)

type orderStatsByMerchantRepository struct {
	db      *db.Queries
	ctx     context.Context
	mapping recordmapper.OrderRecordMapping
}

func NewOrderStatsByMerchantRepository(db *db.Queries, ctx context.Context, mapping recordmapper.OrderRecordMapping) *orderStatsByMerchantRepository {
	return &orderStatsByMerchantRepository{
		db:      db,
		ctx:     ctx,
		mapping: mapping,
	}
}

func (r *orderStatsByMerchantRepository) GetMonthlyTotalRevenueByMerchant(req *requests.MonthTotalRevenueMerchant) ([]*record.OrderMonthlyTotalRevenueRecord, error) {
	currentMonthStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	currentMonthEnd := currentMonthStart.AddDate(0, 1, -1)
	prevMonthStart := currentMonthStart.AddDate(0, -1, 0)
	prevMonthEnd := prevMonthStart.AddDate(0, 1, -1)

	res, err := r.db.GetMonthlyTotalRevenueByMerchant(r.ctx, db.GetMonthlyTotalRevenueByMerchantParams{
		Extract:     currentMonthStart,
		CreatedAt:   sql.NullTime{Time: currentMonthEnd, Valid: true},
		CreatedAt_2: sql.NullTime{Time: prevMonthStart, Valid: true},
		CreatedAt_3: sql.NullTime{Time: prevMonthEnd, Valid: true},
		MerchantID:  int32(req.MerchantID),
	})

	if err != nil {
		return nil, order_errors.ErrGetMonthlyTotalRevenueByMerchant
	}

	so := r.mapping.ToOrderMonthlyTotalRevenuesByMerchant(res)

	return so, nil
}

func (r *orderStatsByMerchantRepository) GetYearlyTotalRevenueByMerchant(req *requests.YearTotalRevenueMerchant) ([]*record.OrderYearlyTotalRevenueRecord, error) {
	res, err := r.db.GetYearlyTotalRevenueByMerchant(r.ctx, db.GetYearlyTotalRevenueByMerchantParams{
		Column1:    int32(req.Year),
		MerchantID: int32(req.MerchantID),
	})

	if err != nil {
		return nil, order_errors.ErrGetYearlyTotalRevenueByMerchant
	}

	so := r.mapping.ToOrderYearlyTotalRevenuesByMerchant(res)

	return so, nil
}

func (r *orderStatsByMerchantRepository) GetMonthlyOrderByMerchant(req *requests.MonthOrderMerchant) ([]*record.OrderMonthlyRecord, error) {
	yearStart := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)

	res, err := r.db.GetMonthlyOrderByMerchant(r.ctx, db.GetMonthlyOrderByMerchantParams{
		Column1:    yearStart,
		MerchantID: int32(req.MerchantID),
	})
	if err != nil {
		return nil, order_errors.ErrGetMonthlyOrderByMerchant
	}

	return r.mapping.ToOrderMonthlyPricesByMerchant(res), nil
}

func (r *orderStatsByMerchantRepository) GetYearlyOrderByMerchant(req *requests.YearOrderMerchant) ([]*record.OrderYearlyRecord, error) {
	yearStart := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)

	res, err := r.db.GetYearlyOrderByMerchant(r.ctx, db.GetYearlyOrderByMerchantParams{
		Column1:    yearStart,
		MerchantID: int32(req.MerchantID),
	})

	if err != nil {
		return nil, order_errors.ErrGetYearlyOrderByMerchant
	}

	return r.mapping.ToOrderYearlyPricesByMerchant(res), nil
}
