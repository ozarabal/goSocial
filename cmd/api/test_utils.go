package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ozarabal/goSocial/internal/auth"
	"github.com/ozarabal/goSocial/internal/ratelimiter"
	"github.com/ozarabal/goSocial/internal/store"
	"github.com/ozarabal/goSocial/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application{
	t.Helper()

	// logger := zap.NewNop().Sugar()
	logger := zap.Must(zap.NewProduction()).Sugar()
	
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	ratelimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	testAuth := &auth.TestAuthenticator{}

	return &application{
		logger:logger,
		store: mockStore,
		cacheStorage: mockCacheStore,
		authenticator: testAuth,
		config: cfg,
		rateLimiter: ratelimiter,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}