package mencache

import (
	"github.com/MamangRust/monolith-point-of-sale-pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Mencache struct {
	CashierQueryCache           CashierQueryCache
	CashierCommandCache         CashierCommandCache
	CashierStatsCache           CashierStatsCache
	CashierStatsByIdCache       CashierStatsByIdCache
	CashierStatsByMerchantCache CashierStatsByMerchantCache
}

type Deps struct {
	Redis  *redis.Client
	Logger logger.LoggerInterface
}

func NewMencache(deps *Deps) *Mencache {
	cacheStore := NewCacheStore(deps.Redis, deps.Logger)

	return &Mencache{
		CashierQueryCache:           NewCashierQueryCache(cacheStore),
		CashierCommandCache:         NewCashierCommandCache(cacheStore),
		CashierStatsCache:           NewCashierStatsCache(cacheStore),
		CashierStatsByIdCache:       NewCashierStatsByIdCache(cacheStore),
		CashierStatsByMerchantCache: NewCashierStatsByMerchantCache(cacheStore),
	}
}
