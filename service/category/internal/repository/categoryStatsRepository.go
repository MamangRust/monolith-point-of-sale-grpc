package repository

import (
	"context"
	"database/sql"
	"time"

	db "github.com/MamangRust/monolith-point-of-sale-pkg/database/schema"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/record"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/errors/category_errors"
	recordmapper "github.com/MamangRust/monolith-point-of-sale-shared/mapper/record"
)

type categoryStatsRepository struct {
	db      *db.Queries
	mapping recordmapper.CategoryRecordMapper
}

func NewCategoryStatsRepository(db *db.Queries, mapping recordmapper.CategoryRecordMapper) *categoryStatsRepository {
	return &categoryStatsRepository{
		db:      db,
		mapping: mapping,
	}
}

func (r *categoryStatsRepository) GetMonthlyTotalPrice(ctx context.Context, req *requests.MonthTotalPrice) ([]*record.CategoriesMonthlyTotalPriceRecord, error) {
	currentMonthStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	currentMonthEnd := currentMonthStart.AddDate(0, 1, -1)
	prevMonthStart := currentMonthStart.AddDate(0, -1, 0)
	prevMonthEnd := prevMonthStart.AddDate(0, 1, -1)

	res, err := r.db.GetMonthlyTotalPrice(ctx, db.GetMonthlyTotalPriceParams{
		Extract:     currentMonthStart,
		CreatedAt:   sql.NullTime{Time: currentMonthEnd, Valid: true},
		CreatedAt_2: sql.NullTime{Time: prevMonthStart, Valid: true},
		CreatedAt_3: sql.NullTime{Time: prevMonthEnd, Valid: true},
	})

	if err != nil {
		return nil, category_errors.ErrGetMonthlyTotalPrice
	}

	so := r.mapping.ToCategoryMonthlyTotalPrices(res)

	return so, nil
}

func (r *categoryStatsRepository) GetYearlyTotalPrices(ctx context.Context, year int) ([]*record.CategoriesYearlyTotalPriceRecord, error) {
	res, err := r.db.GetYearlyTotalPrice(ctx, int32(year))

	if err != nil {
		return nil, category_errors.ErrGetYearlyTotalPrices
	}

	so := r.mapping.ToCategoryYearlyTotalPrices(res)

	return so, nil
}

func (r *categoryStatsRepository) GetMonthPrice(ctx context.Context, year int) ([]*record.CategoriesMonthPriceRecord, error) {
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	res, err := r.db.GetMonthlyCategory(ctx, yearStart)

	if err != nil {
		return nil, category_errors.ErrGetMonthPrice
	}

	return r.mapping.ToCategoryMonthlyPrices(res), nil
}

func (r *categoryStatsRepository) GetYearPrice(ctx context.Context, year int) ([]*record.CategoriesYearPriceRecord, error) {
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	res, err := r.db.GetYearlyCategory(ctx, yearStart)

	if err != nil {
		return nil, category_errors.ErrGetYearPrice
	}

	return r.mapping.ToCategoryYearlyPrices(res), nil
}
