package mencache

import (
	"context"

	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Mencache struct {
	RoleCommandCache RoleCommandCache
	RoleQueryCache   RoleQueryCache
}

type Deps struct {
	Ctx    context.Context
	Redis  *redis.Client
	Logger logger.LoggerInterface
}

func NewMencache(deps *Deps) *Mencache {
	cacheStore := NewCacheStore(deps.Ctx, deps.Redis, deps.Logger)

	return &Mencache{
		RoleCommandCache: NewRoleCommandCache(cacheStore),
		RoleQueryCache:   NewRoleQueryCache(cacheStore),
	}
}
