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

type categoryStatsByIdRepository struct {
	db      *db.Queries
	ctx     context.Context
	mapping recordmapper.CategoryRecordMapper
}

func NewCategoryStatsByIdRepository(db *db.Queries, ctx context.Context, mapping recordmapper.CategoryRecordMapper) *categoryStatsByIdRepository {
	return &categoryStatsByIdRepository{
		db:      db,
		ctx:     ctx,
		mapping: mapping,
	}
}

func (r *categoryStatsByIdRepository) GetMonthlyTotalPriceById(req *requests.MonthTotalPriceCategory) ([]*record.CategoriesMonthlyTotalPriceRecord, error) {
	currentMonthStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	currentMonthEnd := currentMonthStart.AddDate(0, 1, -1)
	prevMonthStart := currentMonthStart.AddDate(0, -1, 0)
	prevMonthEnd := prevMonthStart.AddDate(0, 1, -1)

	res, err := r.db.GetMonthlyTotalPriceById(r.ctx, db.GetMonthlyTotalPriceByIdParams{
		Extract:     currentMonthStart,
		CreatedAt:   sql.NullTime{Time: currentMonthEnd, Valid: true},
		CreatedAt_2: sql.NullTime{Time: prevMonthStart, Valid: true},
		CreatedAt_3: sql.NullTime{Time: prevMonthEnd, Valid: true},
		CategoryID:  int32(req.CategoryID),
	})

	if err != nil {
		return nil, category_errors.ErrGetMonthlyTotalPriceById
	}

	so := r.mapping.ToCategoryMonthlyTotalPricesById(res)

	return so, nil
}

func (r *categoryStatsByIdRepository) GetYearlyTotalPricesById(req *requests.YearTotalPriceCategory) ([]*record.CategoriesYearlyTotalPriceRecord, error) {
	res, err := r.db.GetYearlyTotalPriceById(r.ctx, db.GetYearlyTotalPriceByIdParams{
		Column1:    int32(req.Year),
		CategoryID: int32(req.CategoryID),
	})

	if err != nil {
		return nil, category_errors.ErrGetYearlyTotalPricesById
	}

	so := r.mapping.ToCategoryYearlyTotalPricesById(res)

	return so, nil
}

func (r *categoryStatsByIdRepository) GetMonthPriceById(req *requests.MonthPriceId) ([]*record.CategoriesMonthPriceRecord, error) {
	yearStart := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)

	res, err := r.db.GetMonthlyCategoryById(r.ctx, db.GetMonthlyCategoryByIdParams{
		Column1:    yearStart,
		CategoryID: int32(req.CategoryID),
	})
	if err != nil {
		return nil, category_errors.ErrGetMonthPriceById
	}

	return r.mapping.ToCategoryMonthlyPricesById(res), nil
}

func (r *categoryStatsByIdRepository) GetYearPriceById(req *requests.YearPriceId) ([]*record.CategoriesYearPriceRecord, error) {
	yearStart := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)

	res, err := r.db.GetYearlyCategoryById(r.ctx, db.GetYearlyCategoryByIdParams{
		Column1:    yearStart,
		CategoryID: int32(req.CategoryID),
	})

	if err != nil {
		return nil, category_errors.ErrGetYearPriceById
	}

	return r.mapping.ToCategoryYearlyPricesById(res), nil
}
