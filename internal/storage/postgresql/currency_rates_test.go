//go:build integration

package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

func Test_pgCurrencyRatesStorage_Get(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateCurrencyRatesStorage()

	t.Run("no rate", func(t *testing.T) {
		// ACT
		rate, ok, err := s.Get(_ctx, "CNY", time.Now())

		// ASSERT
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, rate)
	})

	t.Run("lead rate", func(t *testing.T) {
		// ACT
		rate, ok, err := s.Get(_ctx, "USD", time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC))

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, int64(550000), rate)
	})

	t.Run("lag rate", func(t *testing.T) {
		// ACT
		rate, ok, err := s.Get(_ctx, "USD", time.Date(2022, 10, 25, 0, 0, 0, 0, time.UTC))

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, int64(500000), rate)
	})

	t.Run("on day rate", func(t *testing.T) {
		// ACT
		rate, ok, err := s.Get(_ctx, "USD", time.Date(2022, 10, 20, 0, 0, 0, 0, time.UTC))

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, int64(525000), rate)
	})
}

func Test_pgCurrencyRatesStorage_Add(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateCurrencyRatesStorage()

	t.Run("insert", func(t *testing.T) {
		// ARRANGE
		date := utils.TruncateToDate(time.Now())

		// ACT
		setErr := s.Add(_ctx, "CNY", date, 84500)
		rate, ok, getErr := s.Get(_ctx, "CNY", date)

		// ASSERT
		assert.NoError(t, setErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, int64(84500), rate)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`delete from rates where code = 'CNY'`,
			)
		})
	})

	t.Run("update", func(t *testing.T) {
		// ARRANGE
		date := time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)

		// ACT
		setErr := s.Add(_ctx, "USD", date, 480000)
		rate, ok, getErr := s.Get(_ctx, "USD", date)

		// ASSERT
		assert.NoError(t, setErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, int64(480000), rate)

		// CLEANUP
		_ = s.Add(_ctx, "USD", date, 500000)
	})
}
