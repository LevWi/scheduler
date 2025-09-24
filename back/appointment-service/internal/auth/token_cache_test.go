package auth

import (
	"errors"
	common "scheduler/appointment-service/internal"
	"sync"
	"testing"
	"time"
)

// MockTokenChecker implements the TokenChecker interface for testing
type MockTokenChecker struct {
	mu              sync.Mutex
	responses       map[string]mockResponse
	callLog         []mockCall
	defaultResponse mockResponse
}

type mockResponse struct {
	userID common.ID
	err    error
	delay  time.Duration // for testing concurrent access
}

type mockCall struct {
	clientID common.ID
	token    string
	time     time.Time
}

func NewMockTokenChecker() *MockTokenChecker {
	return &MockTokenChecker{
		responses: make(map[string]mockResponse),
		callLog:   make([]mockCall, 0),
	}
}

func (m *MockTokenChecker) SetResponse(clientID common.ID, token string, userID common.ID, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := clientID + ":" + token
	m.responses[key] = mockResponse{userID: userID, err: err}
}

func (m *MockTokenChecker) SetResponseWithDelay(clientID common.ID, token string, userID common.ID, err error, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := clientID + ":" + token
	m.responses[key] = mockResponse{userID: userID, err: err, delay: delay}
}

func (m *MockTokenChecker) SetDefaultResponse(userID common.ID, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultResponse = mockResponse{userID: userID, err: err}
}

func (m *MockTokenChecker) TokenCheck(clientID common.ID, token string) (common.ID, error) {
	m.mu.Lock()
	key := clientID + ":" + token
	response, exists := m.responses[key]
	if !exists {
		response = m.defaultResponse
	}

	// Log the call
	m.callLog = append(m.callLog, mockCall{
		clientID: clientID,
		token:    token,
		time:     time.Now(),
	})
	m.mu.Unlock()

	// Simulate delay if specified
	if response.delay > 0 {
		time.Sleep(response.delay)
	}

	return response.userID, response.err
}

func (m *MockTokenChecker) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.callLog)
}

func (m *MockTokenChecker) GetCalls() []mockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]mockCall, len(m.callLog))
	copy(calls, m.callLog)
	return calls
}

func (m *MockTokenChecker) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callLog = make([]mockCall, 0)
	m.responses = make(map[string]mockResponse)
	m.defaultResponse = mockResponse{}
}

func TestNewTokenCache(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	if cache == nil {
		t.Fatal("NewTokenCache returned nil")
	}
	if cache.cache == nil {
		t.Error("cache.cache is nil")
	}
	if cache.tc != mockTC {
		t.Error("token checker not set correctly")
	}
}

func TestTokenCache_TokenCheck_FirstCall(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "valid_token"
	expectedUserID := "user456"

	mockTC.SetResponse(clientID, token, expectedUserID, nil)

	userID, err := cache.TokenCheck(clientID, token)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if userID != expectedUserID {
		t.Errorf("expected userID %s, got %s", expectedUserID, userID)
	}
	if mockTC.GetCallCount() != 1 {
		t.Errorf("expected 1 call to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_TokenCheck_CachedCall(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "valid_token"
	expectedUserID := "user456"

	mockTC.SetResponse(clientID, token, expectedUserID, nil)

	// First call
	userID1, err1 := cache.TokenCheck(clientID, token)
	if err1 != nil {
		t.Errorf("first call: expected no error, got %v", err1)
	}
	if userID1 != expectedUserID {
		t.Errorf("first call: expected userID %s, got %s", expectedUserID, userID1)
	}

	// Second call (should be cached)
	userID2, err2 := cache.TokenCheck(clientID, token)
	if err2 != nil {
		t.Errorf("second call: expected no error, got %v", err2)
	}
	if userID2 != expectedUserID {
		t.Errorf("second call: expected userID %s, got %s", expectedUserID, userID2)
	}

	// Should only call the mock once
	if mockTC.GetCallCount() != 1 {
		t.Errorf("expected 1 call to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_TokenCheck_DifferentToken(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token1 := "token1"
	token2 := "token2"
	userID1 := "user1"

	mockTC.SetResponse(clientID, token1, userID1, nil)

	// First call with token1
	result1, err1 := cache.TokenCheck(clientID, token1)
	if err1 != nil {
		t.Errorf("first call: expected no error, got %v", err1)
	}
	if result1 != userID1 {
		t.Errorf("first call: expected userID %s, got %s", userID1, result1)
	}

	// Second call with token2 (should not be cached)
	_, err2 := cache.TokenCheck(clientID, token2)
	if err2 != ErrWrongToken {
		t.Errorf("second call: expected ErrWrongToken, got %v", err2)
	}

	// Should call the mock twice
	if mockTC.GetCallCount() != 1 {
		t.Errorf("expected 2 calls to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_TokenCheck_NotFoundError(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "invalid_token"

	mockTC.SetResponse(clientID, token, "", common.ErrNotFound)

	userID, err := cache.TokenCheck(clientID, token)

	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if userID != "" {
		t.Errorf("expected empty userID, got %s", userID)
	}
}

func TestTokenCache_TokenCheck_WrongTokenError(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "wrong_token"

	mockTC.SetResponse(clientID, token, "", ErrWrongToken)

	userID, err := cache.TokenCheck(clientID, token)

	if !errors.Is(err, ErrWrongToken) {
		t.Errorf("expected ErrWrongToken, got %v", err)
	}
	if userID != "" {
		t.Errorf("expected empty userID, got %s", userID)
	}
}

func TestTokenCache_TokenCheck_CachedError(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "invalid_token"

	mockTC.SetResponse(clientID, token, "", common.ErrNotFound)

	_, err1 := cache.TokenCheck(clientID, token)
	if !errors.Is(err1, common.ErrNotFound) {
		t.Errorf("first call: expected ErrNotFound, got %v", err1)
	}

	_, err2 := cache.TokenCheck(clientID, token)
	if !errors.Is(err2, common.ErrNotFound) {
		t.Errorf("second call: expected ErrNotFound, got %v", err2)
	}

	_, err2 = cache.TokenCheck(clientID, token)
	if !errors.Is(err2, common.ErrNotFound) {
		t.Errorf("second call: expected ErrNotFound, got %v", err2)
	}

	// Should only call the mock once
	if mockTC.GetCallCount() != 1 {
		t.Errorf("expected 1 call to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_TokenCheck_InternalError(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "some_token"
	internalErr := errors.New("database connection failed")

	mockTC.SetResponse(clientID, token, "", internalErr)

	const numCalls = 3
	for i := 0; i < numCalls; i++ {
		userID, err := cache.TokenCheck(clientID, token)

		// Internal errors should not be cached and should be returned immediately
		if err != internalErr {
			t.Errorf("call %d: expected internal error, got %v", i+1, err)
		}
		if userID != "" {
			t.Errorf("call %d: expected empty userID, got %s", i+1, userID)
		}
	}

	// Because internal errors are not cached, the mock should be called each time.
	if mockTC.GetCallCount() != numCalls {
		t.Errorf("expected %d calls to TokenChecker, got %d", numCalls, mockTC.GetCallCount())
	}
}

func TestTokenCache_Forget(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "valid_token"
	userID := "user456"

	mockTC.SetResponse(clientID, token, userID, nil)

	// First call to populate cache
	_, err := cache.TokenCheck(clientID, token)
	if err != nil {
		t.Errorf("first call failed: %v", err)
	}

	// Forget the client
	cache.Forget(clientID)

	// Second call should hit the TokenChecker again
	_, err = cache.TokenCheck(clientID, token)
	if err != nil {
		t.Errorf("second call failed: %v", err)
	}

	// Should call the mock twice (not cached after Forget)
	if mockTC.GetCallCount() != 2 {
		t.Errorf("expected 2 calls to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_ConcurrentAccess(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	clientID := "client123"
	token := "valid_token"
	expectedUserID := "user456"

	// Add a small delay to make race conditions more likely
	mockTC.SetResponseWithDelay(clientID, token, expectedUserID, nil, 10*time.Millisecond)

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([]struct {
		userID string
		err    error
	}, numGoroutines)

	// Launch multiple goroutines calling TokenCheck simultaneously
	for i := range numGoroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			userID, err := cache.TokenCheck(clientID, token)
			results[index] = struct {
				userID string
				err    error
			}{userID, err}
		}(i)
	}

	wg.Wait()

	// All calls should succeed
	for i, result := range results {
		if result.err != nil {
			t.Errorf("goroutine %d: expected no error, got %v", i, result.err)
		}
		if result.userID != expectedUserID {
			t.Errorf("goroutine %d: expected userID %s, got %s", i, expectedUserID, result.userID)
		}
	}

	// The underlying TokenChecker should only be called once despite concurrent access
	if mockTC.GetCallCount() != 1 {
		t.Errorf("expected 1 call to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_MultipleDifferentClients(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	client1 := "client1"
	client2 := "client2"
	token := "same_token"
	user1 := "user1"
	user2 := "user2"

	mockTC.SetResponse(client1, token, user1, nil)
	mockTC.SetResponse(client2, token, user2, nil)

	// Call for client1
	result1, err1 := cache.TokenCheck(client1, token)
	if err1 != nil {
		t.Errorf("client1: expected no error, got %v", err1)
	}
	if result1 != user1 {
		t.Errorf("client1: expected userID %s, got %s", user1, result1)
	}

	// Call for client2
	result2, err2 := cache.TokenCheck(client2, token)
	if err2 != nil {
		t.Errorf("client2: expected no error, got %v", err2)
	}
	if result2 != user2 {
		t.Errorf("client2: expected userID %s, got %s", user2, result2)
	}

	// Should call the mock twice (different clients)
	if mockTC.GetCallCount() != 2 {
		t.Errorf("expected 2 calls to TokenChecker, got %d", mockTC.GetCallCount())
	}

	// Subsequent calls should be cached
	cache.TokenCheck(client1, token)
	cache.TokenCheck(client2, token)

	// Still should be 2 calls total
	if mockTC.GetCallCount() != 2 {
		t.Errorf("expected 2 calls to TokenChecker after caching, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_ForgetNonExistentClient(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	// Forgetting a non-existent client should not panic
	cache.Forget("non_existent_client")

	// Should be no calls to the mock
	if mockTC.GetCallCount() != 0 {
		t.Errorf("expected 0 calls to TokenChecker, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_TokenCheck_EmptyValues(t *testing.T) {
	mockTC := NewMockTokenChecker()
	cache := NewTokenCacheDefault(mockTC)

	// Test with empty clientID
	_, err := cache.TokenCheck("", "token")
	if !errors.Is(err, common.ErrInvalidArgument) {
		t.Errorf("empty clientID: expected ErrInvalidArgument, got %v", err)
	}

	// Test with empty token
	_, err = cache.TokenCheck("client", "")
	if !errors.Is(err, common.ErrInvalidArgument) {
		t.Errorf("empty token: expected ErrInvalidArgument, got %v", err)
	}

	// Test with both empty
	_, err = cache.TokenCheck("", "")
	if !errors.Is(err, common.ErrInvalidArgument) {
		t.Errorf("both empty: expected ErrInvalidArgument, got %v", err)
	}

	// The mock should not be called at all
	if mockTC.GetCallCount() != 0 {
		t.Errorf("expected 0 calls to TokenChecker for invalid arguments, got %d", mockTC.GetCallCount())
	}
}

func TestTokenCache_ForgetExpired(t *testing.T) {
	mockTC := NewMockTokenChecker()
	// Use a short lifetime for testing
	cache := NewTokenCache(mockTC, 50*time.Millisecond)

	client1 := "client1"
	token1 := "token1"
	user1 := "user1"

	client2 := "client2"
	token2 := "token2"
	user2 := "user2"

	mockTC.SetResponse(client1, token1, user1, nil)
	mockTC.SetResponse(client2, token2, user2, nil)

	// Populate the cache for client1
	_, _ = cache.TokenCheck(client1, token1)

	// Wait for a short period, but less than the cache lifetime
	time.Sleep(20 * time.Millisecond)

	// Populate the cache for client2
	_, _ = cache.TokenCheck(client2, token2)

	// Wait for a period that makes client1's entry expire, but not client2's
	time.Sleep(40 * time.Millisecond)

	cache.ForgetExpired()

	// After ForgetExpired, client1 should be gone, but client2 should remain.

	// This call for client1 should hit the mock again
	_, _ = cache.TokenCheck(client1, token1)
	if mockTC.GetCallCount() != 3 {
		t.Errorf("expected 3 calls to TokenChecker (2 initial, 1 after expiry), got %d", mockTC.GetCallCount())
	}

	// This call for client2 should be cached
	_, _ = cache.TokenCheck(client2, token2)
	if mockTC.GetCallCount() != 3 { // Should not increase
		t.Errorf("expected 3 calls to TokenChecker (client2 should be cached), got %d", mockTC.GetCallCount())
	}
}
