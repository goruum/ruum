package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goruum/ruum/core"
)

func TestCache(t *testing.T) {
	config := DefaultCacheConfig()
	config.TTL = time.Second
	middleware := Cache(config)

	callCount := 0
	handler := func(ctx core.Context) error {
		callCount++
		return ctx.JSON(200, map[string]string{"message": "test"})
	}

	t.Run("cache miss then hit", func(t *testing.T) {
		callCount = 0

		// First request - cache miss
		req1 := httptest.NewRequest("GET", "/test", nil)
		res1 := httptest.NewRecorder()
		ctx1 := core.NewContext(context.Background(), req1, res1, core.NewContainer())

		err := middleware(handler)(ctx1)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}

		if callCount != 1 {
			t.Errorf("Call count = %d, want 1", callCount)
		}

		// Second request - cache hit
		req2 := httptest.NewRequest("GET", "/test", nil)
		res2 := httptest.NewRecorder()
		ctx2 := core.NewContext(context.Background(), req2, res2, core.NewContainer())

		err = middleware(handler)(ctx2)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}

		if callCount != 1 {
			t.Errorf("Call count = %d, want 1 (cached)", callCount)
		}
	})

	t.Run("skip non-GET method", func(t *testing.T) {
		callCount = 0

		req := httptest.NewRequest("POST", "/test", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}

		if callCount != 1 {
			t.Errorf("POST should not be cached")
		}
	})

	t.Run("skip function", func(t *testing.T) {
		config := DefaultCacheConfig()
		config.SkipFunc = func(ctx core.Context) bool {
			return true
		}
		middleware := Cache(config)

		callCount = 0

		req := httptest.NewRequest("GET", "/test", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}
	})
}

func TestCacheStore_Cleanup(t *testing.T) {
	store := newCacheStore()
	
	// Add entries that will expire
	store.Set("key1", &CacheEntry{
		StatusCode: 200,
		Body:       []byte("value1"),
		Headers:    make(map[string][]string),
		Expiration: time.Now().Add(time.Millisecond * 50),
	})
	store.Set("key2", &CacheEntry{
		StatusCode: 200,
		Body:       []byte("value2"),
		Headers:    make(map[string][]string),
		Expiration: time.Now().Add(time.Second * 10),
	})
	
	// Wait for cleanup to run and remove expired entries
	time.Sleep(time.Millisecond * 150)
	
	// key1 should be cleaned up
	_, found := store.Get("key1")
	if found {
		t.Error("Expired entry should have been cleaned up")
	}
	
	// key2 should still exist
	_, found = store.Get("key2")
	if !found {
		t.Error("Non-expired entry should still exist")
	}
}

