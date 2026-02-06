package middleware_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
)

func TestTokenBlacklist(t *testing.T) {
	t.Run("blacklisted token is detected", func(t *testing.T) {
		bl := middleware.NewTokenBlacklist()
		defer bl.Stop()

		token := "test-jwt-token-string"
		bl.Add(token, time.Now().Add(time.Hour))

		assert.True(t, bl.IsBlacklisted(token))
	})

	t.Run("non-blacklisted token is not detected", func(t *testing.T) {
		bl := middleware.NewTokenBlacklist()
		defer bl.Stop()

		bl.Add("token-a", time.Now().Add(time.Hour))

		assert.False(t, bl.IsBlacklisted("token-b"))
	})

	t.Run("different tokens are independent", func(t *testing.T) {
		bl := middleware.NewTokenBlacklist()
		defer bl.Stop()

		bl.Add("token-1", time.Now().Add(time.Hour))
		bl.Add("token-2", time.Now().Add(time.Hour))

		assert.True(t, bl.IsBlacklisted("token-1"))
		assert.True(t, bl.IsBlacklisted("token-2"))
		assert.False(t, bl.IsBlacklisted("token-3"))
	})

	t.Run("concurrent access is safe", func(t *testing.T) {
		bl := middleware.NewTokenBlacklist()
		defer bl.Stop()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(2)
			token := "token-" + string(rune(i))
			go func() {
				defer wg.Done()
				bl.Add(token, time.Now().Add(time.Hour))
			}()
			go func() {
				defer wg.Done()
				bl.IsBlacklisted(token)
			}()
		}
		wg.Wait()
	})
}
