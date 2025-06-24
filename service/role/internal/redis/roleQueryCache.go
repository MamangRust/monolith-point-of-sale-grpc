package mencache

import (
	"fmt"
	"time"

	"github.com/MamangRust/monolith-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/monolith-point-of-sale-shared/domain/response"
)

const (
	roleAllCacheKey     = "role:all:page:%d:pageSize:%d:search:%s"
	roleByIdCacheKey    = "role:id:%d"
	roleActiveCacheKey  = "role:active:page:%d:pageSize:%d:search:%s"
	roleTrashedCacheKey = "role:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)

type roleCachedResponse struct {
	Data         []*response.RoleResponse `json:"data"`
	TotalRecords *int                     `json:"total_records"`
}

type roleCachedResponseDeleteAt struct {
	Data         []*response.RoleResponseDeleteAt `json:"data"`
	TotalRecords *int                             `json:"total_records"`
}

type roleQueryCache struct {
	store *CacheStore
}

func NewRoleQueryCache(store *CacheStore) *roleQueryCache {
	return &roleQueryCache{store: store}
}

func (m *roleQueryCache) SetCachedRoles(req *requests.FindAllRoles, data []*response.RoleResponse, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*response.RoleResponse{}
	}

	key := fmt.Sprintf(roleAllCacheKey, req.Page, req.PageSize, req.Search)

	payload := &roleCachedResponse{Data: data, TotalRecords: total}
	SetToCache(m.store, key, payload, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleById(data *response.RoleResponse) {
	if data == nil {
		data = &response.RoleResponse{}
	}

	key := fmt.Sprintf(roleByIdCacheKey, data.ID)
	SetToCache(m.store, key, data, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleByUserId(userId int, data []*response.RoleResponse) {
	if data == nil {
		data = []*response.RoleResponse{}
	}

	key := fmt.Sprintf(roleByIdCacheKey, userId)

	SetToCache(m.store, key, &data, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleActive(req *requests.FindAllRoles, data []*response.RoleResponseDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*response.RoleResponseDeleteAt{}
	}

	key := fmt.Sprintf(roleActiveCacheKey, req.Page, req.PageSize, req.Search)

	payload := &roleCachedResponseDeleteAt{Data: data, TotalRecords: total}
	SetToCache(m.store, key, payload, ttlDefault)
}

func (m *roleQueryCache) SetCachedRoleTrashed(req *requests.FindAllRoles, data []*response.RoleResponseDeleteAt, total *int) {
	if total == nil {
		zero := 0
		total = &zero
	}

	if data == nil {
		data = []*response.RoleResponseDeleteAt{}
	}

	key := fmt.Sprintf(roleTrashedCacheKey, req.Page, req.PageSize, req.Search)
	payload := &roleCachedResponseDeleteAt{Data: data, TotalRecords: total}
	SetToCache(m.store, key, payload, ttlDefault)
}

func (m *roleQueryCache) GetCachedRoles(req *requests.FindAllRoles) ([]*response.RoleResponse, *int, bool) {
	key := fmt.Sprintf(roleAllCacheKey, req.Page, req.PageSize, req.Search)

	result, found := GetFromCache[roleCachedResponse](m.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (m *roleQueryCache) GetCachedRoleById(id int) (*response.RoleResponse, bool) {
	key := fmt.Sprintf(roleByIdCacheKey, id)

	result, found := GetFromCache[*response.RoleResponse](m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *roleQueryCache) GetCachedRoleByUserId(userId int) ([]*response.RoleResponse, bool) {
	key := fmt.Sprintf(roleByIdCacheKey, userId)

	result, found := GetFromCache[[]*response.RoleResponse](m.store, key)

	if !found || result == nil {
		return nil, false
	}

	return *result, true
}

func (m *roleQueryCache) GetCachedRoleActive(req *requests.FindAllRoles) ([]*response.RoleResponseDeleteAt, *int, bool) {
	key := fmt.Sprintf(roleActiveCacheKey, req.Page, req.PageSize, req.Search)

	result, found := GetFromCache[roleCachedResponseDeleteAt](m.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}

func (m *roleQueryCache) GetCachedRoleTrashed(req *requests.FindAllRoles) ([]*response.RoleResponseDeleteAt, *int, bool) {
	key := fmt.Sprintf(roleTrashedCacheKey, req.Page, req.PageSize, req.Search)

	result, found := GetFromCache[roleCachedResponseDeleteAt](m.store, key)

	if !found || result == nil {
		return nil, nil, false
	}

	return result.Data, result.TotalRecords, true
}
