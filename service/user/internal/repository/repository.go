package repository

import (
	db "github.com/MamangRust/monolith-point-of-sale-pkg/database/schema"
	recordmapper "github.com/MamangRust/monolith-point-of-sale-shared/mapper/record"
)

type Repositories struct {
	UserCommand UserCommandRepository
	UserQuery   UserQueryRepository
	Role        RoleQueryRepository
}

func NewRepositories(DB *db.Queries) *Repositories {
	mapper := recordmapper.NewUserRecordMapper()
	mapperrole := recordmapper.NewRoleRecordMapper()

	return &Repositories{
		UserCommand: NewUserCommandRepository(DB, mapper),
		UserQuery:   NewUserQueryRepository(DB, mapper),
		Role:        NewRoleRepository(DB, mapperrole),
	}
}
