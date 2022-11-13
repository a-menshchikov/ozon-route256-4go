//go:build integration

package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pgCurrencyStorage_Get(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateCurrencyStorage()

	t.Run("user without currency", func(t *testing.T) {
		// ACT
		currency, ok, err := s.Get(_ctx, _testUser101)

		// ASSERT
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, currency)
	})

	t.Run("success", func(t *testing.T) {
		// ACT
		currency, ok, err := s.Get(_ctx, _testUser102)

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "USD", currency)
	})
}

func Test_pgCurrencyStorage_Set(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateCurrencyStorage()

	t.Run("insert", func(t *testing.T) {
		// ACT
		setErr := s.Set(_ctx, _testUser101, "USD")
		currency, ok, getErr := s.Get(_ctx, _testUser101)

		// ASSERT
		assert.NoError(t, setErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, "USD", currency)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`delete from currencies where user_id = $1`,
				int64(*_testUser101),
			)
		})
	})

	t.Run("update", func(t *testing.T) {
		// ACT
		setErr := s.Set(_ctx, _testUser102, "RUB")
		currency, ok, getErr := s.Get(_ctx, _testUser102)

		// ASSERT
		assert.NoError(t, setErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, "RUB", currency)

		// CLEANUP
		_ = s.Set(_ctx, _testUser102, "USD")
	})
}
