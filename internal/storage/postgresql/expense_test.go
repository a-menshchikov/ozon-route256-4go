//go:build integration

package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

func Test_pgExpenseStorage_Add_List(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateExpenseStorage()

	t.Run("empty list", func(t *testing.T) {
		// ACT
		list, err := s.List(_ctx, _testUser101, time.Date(2022, 10, 22, 0, 0, 0, 0, time.UTC))

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("add with check", func(t *testing.T) {
		// ACT
		listBefore, listBeforeErr := s.List(_ctx, _testUser101, time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC))
		addErr := s.Add(_ctx, _testUser101, types.ExpenseItem{
			Date:     time.Date(2022, 10, 23, 0, 0, 0, 0, time.UTC),
			Amount:   1200000,
			Currency: "RUB",
		}, "coffee")
		listAfter, listAfterErr := s.List(_ctx, _testUser101, time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC))

		// ASSERT
		assert.NoError(t, listBeforeErr)
		assert.NoError(t, addErr)
		assert.NoError(t, listAfterErr)

		assert.Equal(t, map[string][]types.ExpenseItem{
			"taxi": {
				{
					Date:     time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC),
					Amount:   10000000,
					Currency: "RUB",
				},
			},
			"medicine": {
				{
					Date:     time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC),
					Amount:   250000,
					Currency: "USD",
				},
			},
		}, listBefore)
		assert.Equal(t, map[string][]types.ExpenseItem{
			"taxi": {
				{
					Date:     time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC),
					Amount:   10000000,
					Currency: "RUB",
				},
			},
			"medicine": {
				{
					Date:     time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC),
					Amount:   250000,
					Currency: "USD",
				},
			},
			"coffee": {
				{
					Date:     time.Date(2022, 10, 23, 0, 0, 0, 0, time.UTC),
					Amount:   1200000,
					Currency: "RUB",
				},
			},
		}, listAfter)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`delete
                 from expenses
                 where user_id = $1
                   and category = 'coffee'`,
				int64(*_testUser101),
			)
		})
	})
}
