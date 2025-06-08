package specs

import (
	"testing"
	"time"
)

func TestCookie_IsExpired(t *testing.T) {
	now := time.Now()

	t.Run("MaxAge > 0 sets Expires and returns false", func(t *testing.T) {
		c := Cookie{
			MaxAge: 10, // 10 seconds
		}

		expired := c.IsExpired(now)

		if expired {
			t.Error("Expected cookie to be not expired when MaxAge > 0")
		}
		if c.Expires.Before(now.Add(9*time.Second)) || c.Expires.After(now.Add(11*time.Second)) {
			t.Errorf("Expected Expires to be set ~10s into future, got: %v", c.Expires)
		}
		if c.MaxAge != 0 {
			t.Errorf("Expected MaxAge to be reset to 0, got: %d", c.MaxAge)
		}
	})

	t.Run("Expires in the past returns true", func(t *testing.T) {
		c := Cookie{
			Expires: now.Add(-time.Minute),
		}

		expired := c.IsExpired(now)

		if !expired {
			t.Error("Expected cookie to be expired when Expires is in the past")
		}
	})

	t.Run("Expires in the future returns false", func(t *testing.T) {
		c := Cookie{
			Expires: now.Add(time.Hour),
		}

		expired := c.IsExpired(now)

		if expired {
			t.Error("Expected cookie to be not expired when Expires is in the future")
		}
	})

	t.Run("Expires and MaxAge is zero returns false", func(t *testing.T) {
		c := Cookie{
			Expires: time.Time{},
		}

		expired := c.IsExpired(now)

		if expired {
			t.Error("Expected cookie to be not expired when Expires is zero")
		}
	})
}
