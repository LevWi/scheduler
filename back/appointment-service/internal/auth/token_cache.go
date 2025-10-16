package auth

import (
	"errors"
	common "scheduler/appointment-service/internal"
	"sync"
	"time"
)

// TODO add token brute force protection?
type cachePayload struct {
	lastRead time.Time
	userID   common.ID
	token    string
	err      error
	checked  bool
	mu       sync.Mutex
}

type TokenCache struct {
	mu       sync.Mutex
	cache    map[string]*cachePayload
	lifetime time.Duration
	tc       TokenChecker
}

func (tokenCache *TokenCache) TokenCheck(clientID common.ID, token string) (common.ID, error) {
	if clientID == "" || token == "" {
		return "", common.ErrInvalidArgument
	}

	var result *cachePayload
	tokenCache.mu.Lock()
	result = tokenCache.cache[clientID]

	if result == nil {
		result = &cachePayload{}
		tokenCache.cache[clientID] = result
	}
	result.mu.Lock()
	tokenCache.mu.Unlock()

	defer result.mu.Unlock()
	result.lastRead = time.Now()

	if !result.checked {
		result.token = token
		userID, err := tokenCache.tc.TokenCheck(clientID, token)
		result.userID = userID
		if err != nil {
			if errors.Is(err, common.ErrNotFound) || errors.Is(err, ErrWrongToken) {
				result.err = err
			} else {
				return "", err
			}
		}
		result.checked = true
	}

	if result.token != token {
		return "", ErrWrongToken
	}

	return result.userID, result.err
}

func (tokenCache *TokenCache) ForgetExpired() uint {
	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()

	var i uint
	now := time.Now()
	for clientID, payload := range tokenCache.cache {
		if payload.mu.TryLock() {
			if now.Sub(payload.lastRead) > tokenCache.lifetime {
				i++
				delete(tokenCache.cache, clientID)
			}
			payload.mu.Unlock()
		}
	}
	return i
}

func (tokenCache *TokenCache) Forget(clientID common.ID) {
	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()
	delete(tokenCache.cache, clientID)
}

//TODO clear unused cached variables after period

func NewTokenCache(tc TokenChecker, cacheLifetime time.Duration) *TokenCache {
	return &TokenCache{
		cache:    make(map[string]*cachePayload),
		lifetime: cacheLifetime,
		tc:       tc,
	}
}

func NewTokenCacheDefault(tc TokenChecker) *TokenCache {
	return NewTokenCache(tc, 10*time.Minute)
}
