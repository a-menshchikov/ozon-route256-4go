//go:build integration

package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

func Test_pgExpenseLimitStorage_Get(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateExpenseLimitStorage()

	t.Run("user without limits", func(t *testing.T) {
		// ACT
		limit, ok, err := s.Get(_ctx, _testUser102, "tea")

		// ASSERT
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, limit)
	})

	t.Run("remains limit", func(t *testing.T) {
		// ACT
		limit, ok, err := s.Get(_ctx, _testUser101, "tea")

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    25000000,
			Remains:  20000000,
			Currency: "RUB",
		}, limit)
	})

	t.Run("category limit", func(t *testing.T) {
		// ACT
		limit, ok, err := s.Get(_ctx, _testUser101, "taxi")

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    1000000,
			Remains:  700000,
			Currency: "USD",
		}, limit)
	})
}

func Test_pgExpenseLimitStorage_Set(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateExpenseLimitStorage()

	t.Run("insert", func(t *testing.T) {
		// ACT
		setErr := s.Set(_ctx, _testUser102, 100000, "EUR", "tea")
		limit, ok, getErr := s.Get(_ctx, _testUser102, "tea")

		// ASSERT
		assert.NoError(t, setErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    100000,
			Remains:  100000,
			Currency: "EUR",
		}, limit)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`delete
                 from limits
                 where user_id = $1`,
				int64(*_testUser102),
			)
		})
	})

	t.Run("update", func(t *testing.T) {
		// ACT
		setErr := s.Set(_ctx, _testUser101, 12000000, "USD", "taxi")
		limit, ok, getErr := s.Get(_ctx, _testUser101, "taxi")

		// ASSERT
		assert.NoError(t, setErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    12000000,
			Remains:  12000000,
			Currency: "USD",
		}, limit)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`update limits
                 set total = 1000000,
                     remains = 700000
                 where user_id = $1
                   and category = 'taxi'`,
				int64(*_testUser101),
			)
		})
	})
}

func Test_pgExpenseLimitStorage_Decrease(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateExpenseLimitStorage()

	t.Run("limit not reached", func(t *testing.T) {
		// ACT
		reached, decreaseErr := s.Decrease(_ctx, _testUser101, 650000, "taxi")
		limit, ok, getErr := s.Get(_ctx, _testUser101, "taxi")

		// ASSERT
		assert.NoError(t, decreaseErr)
		assert.NoError(t, getErr)
		assert.False(t, reached)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    1000000,
			Remains:  50000,
			Currency: "USD",
		}, limit)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`update limits
                 set total = 1000000,
                     remains = 700000
                 where user_id = $1
                   and category = 'taxi'`,
				int64(*_testUser101),
			)
		})
	})

	t.Run("limit reached", func(t *testing.T) {
		// ACT
		reached, decreaseErr := s.Decrease(_ctx, _testUser101, 750000, "taxi")
		limit, ok, getErr := s.Get(_ctx, _testUser101, "taxi")

		// ASSERT
		assert.NoError(t, decreaseErr)
		assert.NoError(t, getErr)
		assert.True(t, reached)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    1000000,
			Remains:  0,
			Currency: "USD",
		}, limit)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`update limits
                 set total = 1000000,
                     remains = 700000
                 where user_id = $1
                   and category = 'taxi'`,
				int64(*_testUser101),
			)
		})
	})
}

func Test_pgExpenseLimitStorage_Unset(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateExpenseLimitStorage()

	t.Run("success", func(t *testing.T) {
		// ACT
		unsetErr := s.Unset(_ctx, _testUser101, "taxi")
		limit, ok, getErr := s.Get(_ctx, _testUser101, "taxi")

		// ASSERT
		assert.NoError(t, unsetErr)
		assert.NoError(t, getErr)
		assert.True(t, ok)
		assert.Equal(t, types.LimitItem{
			Total:    25000000,
			Remains:  20000000,
			Currency: "RUB",
		}, limit)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`insert into limits (user_id, category, total, remains, currency_code)
                 values ($1, 'taxi', 1000000, 700000, 'USD')`,
				int64(*_testUser101),
			)
		})
	})
}

func Test_pgExpenseLimitStorage_List(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateExpenseLimitStorage()

	t.Run("empty list", func(t *testing.T) {
		// ACT
		list, ok, err := s.List(_ctx, _testUser102)

		// ASSERT
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, list)
	})

	t.Run("complete list", func(t *testing.T) {
		// ACT
		list, ok, err := s.List(_ctx, _testUser101)

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, map[string]types.LimitItem{
			"": {
				Total:    25000000,
				Remains:  20000000,
				Currency: "RUB",
			},
			"taxi": {
				Total:    1000000,
				Remains:  700000,
				Currency: "USD",
			},
		}, list)
	})
}
