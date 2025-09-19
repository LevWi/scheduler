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
	mu    sync.Mutex
	cache map[string]*cachePayload
	tc    TokenChecker
}

// TODO add test
func (tokenCache *TokenCache) TokenCheck(clientID common.ID, token string) (common.ID, error) {
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

	if !result.checked || result.token != token {
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

	return result.userID, result.err
}

func (tokenCache *TokenCache) Forget(clientID common.ID) {
	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()
	delete(tokenCache.cache, clientID)
}

//TODO clear unused cached variables after period

func NewTokenCache(tc TokenChecker) *TokenCache {
	return &TokenCache{
		cache: make(map[string]*cachePayload),
		tc:    tc,
	}
}
