package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows requests within limit", func(t *testing.T) {
		cfg := &config.SecurityConfig{
			RateLimitEnabled:  true,
			RateLimitRequests: 10,
			RateLimitWindow:   60,
			RateLimitBurst:    10,
		}
		rl := middleware.NewRateLimiter(cfg)
		defer rl.Stop()

		router := gin.New()
		router.Use(rl.Middleware())
		router.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// First request should succeed
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks requests exceeding burst", func(t *testing.T) {
		cfg := &config.SecurityConfig{
			RateLimitEnabled:  true,
			RateLimitRequests: 1,
			RateLimitWindow:   60,
			RateLimitBurst:    3,
		}
		rl := middleware.NewRateLimiter(cfg)
		defer rl.Stop()

		router := gin.New()
		router.Use(rl.Middleware())
		router.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Exhaust the burst
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("POST", "/login", nil)
			req.RemoteAddr = "10.0.0.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Next request should be rate limited
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
	})

	t.Run("different IPs have independent limits", func(t *testing.T) {
		cfg := &config.SecurityConfig{
			RateLimitEnabled:  true,
			RateLimitRequests: 1,
			RateLimitWindow:   60,
			RateLimitBurst:    2,
		}
		rl := middleware.NewRateLimiter(cfg)
		defer rl.Stop()

		router := gin.New()
		router.Use(rl.Middleware())
		router.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Exhaust burst for IP 1
		for i := 0; i < 2; i++ {
			req, _ := http.NewRequest("POST", "/login", nil)
			req.RemoteAddr = "10.0.0.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		// IP 1 should be blocked
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// IP 2 should still work
		req, _ = http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("concurrent requests are safe", func(t *testing.T) {
		cfg := &config.SecurityConfig{
			RateLimitEnabled:  true,
			RateLimitRequests: 100,
			RateLimitWindow:   1,
			RateLimitBurst:    100,
		}
		rl := middleware.NewRateLimiter(cfg)
		defer rl.Stop()

		router := gin.New()
		router.Use(rl.Middleware())
		router.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("POST", "/login", nil)
				req.RemoteAddr = "10.0.0.1:12345"
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				// Should be 200 or 429, not a panic
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusTooManyRequests)
			}()
		}
		wg.Wait()
	})
}
