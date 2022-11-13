//go:build integration

package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

func Test_pgTelegramUserStorage_Add(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateTelegramUserStorage()

	t.Run("duplicate user", func(t *testing.T) {
		// ACT
		user, err := s.Add(_ctx, 101)

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, user)
	})

	t.Run("success", func(t *testing.T) {
		// ACT
		user, err := s.Add(_ctx, 1234)

		// ASSERT
		assert.NoError(t, err)
		assert.IsType(t, &([]types.User{types.User(0)}[0]), user)

		// CLEANUP
		t.Cleanup(func() {
			_, _ = _testFactory.pool.Exec(
				_ctx,
				`delete
                 from users
                 where id = $1`,
				int64(*user),
			)
		})
	})
}

func Test_pgTelegramUserStorage_FetchByID(t *testing.T) {
	// ARRANGE
	s := _testFactory.CreateTelegramUserStorage()

	t.Run("no user", func(t *testing.T) {
		// ACT
		user, err := s.FetchByID(_ctx, 99999)

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, user)
	})

	t.Run("success", func(t *testing.T) {
		// ACT
		user, err := s.FetchByID(_ctx, 101)

		// ASSERT
		assert.NoError(t, err)
		assert.IsType(t, &([]types.User{types.User(0)}[0]), user)
	})
}
