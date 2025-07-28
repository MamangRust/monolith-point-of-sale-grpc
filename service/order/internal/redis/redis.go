package mencache

import (
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Mencache struct {
	OrderQueryCache           OrderQueryCache
	OrderCommandCache         OrderCommandCache
	OrderStatsCache           OrderStatsCache
	OrderStatsByMerchantCache OrderStatsByMerchantCache
}

type Deps struct {
	Redis  *redis.Client
	Logger logger.LoggerInterface
}

func NewMencache(deps *Deps) *Mencache {
	cacheStore := NewCacheStore(deps.Redis, deps.Logger)

	return &Mencache{
		OrderQueryCache:           NewOrderQueryCache(cacheStore),
		OrderCommandCache:         NewOrderCommandCache(cacheStore),
		OrderStatsCache:           NewOrderStatsCache(cacheStore),
		OrderStatsByMerchantCache: NewOrderStatsByMerchantCache(cacheStore),
	}
}
