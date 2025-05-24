package repository

import (
	"context"

	db "github.com/MamangRust/monolith-point-of-sale-pkg/database/schema"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/record"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	userrole_errors "github.com/MamangRust/monolith-point-of-sale-shared/errors/user_role_errors"
	recordmapper "github.com/MamangRust/monolith-point-of-sale-shared/mapper/record"
)

type userRoleRepository struct {
	db      *db.Queries
	ctx     context.Context
	mapping recordmapper.UserRoleRecordMapping
}

func NewUserRoleRepository(db *db.Queries, ctx context.Context, mapping recordmapper.UserRoleRecordMapping) *userRoleRepository {
	return &userRoleRepository{
		db:      db,
		ctx:     ctx,
		mapping: mapping,
	}
}

func (r *userRoleRepository) AssignRoleToUser(req *requests.CreateUserRoleRequest) (*record.UserRoleRecord, error) {
	res, err := r.db.AssignRoleToUser(r.ctx, db.AssignRoleToUserParams{
		UserID: int32(req.UserId),
		RoleID: int32(req.RoleId),
	})

	if err != nil {
		return nil, userrole_errors.ErrAssignRoleToUser
	}

	return r.mapping.ToUserRoleRecord(res), nil
}

func (r *userRoleRepository) RemoveRoleFromUser(req *requests.RemoveUserRoleRequest) error {
	err := r.db.RemoveRoleFromUser(r.ctx, db.RemoveRoleFromUserParams{
		UserID: int32(req.UserId),
		RoleID: int32(req.RoleId),
	})

	if err != nil {
		return userrole_errors.ErrRemoveRole
	}

	return nil
}
