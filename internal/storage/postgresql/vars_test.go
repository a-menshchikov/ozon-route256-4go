//go:build integration

package postgresql

import (
	"context"
	"log"
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

var (
	_ctx = context.Background()

	_testFactory = func() *factory {
		factory, err := NewFactory(
			_ctx,
			"postgres://postgres:postgres@localhost:5433/testdb?sslmode=disable",
			3*time.Second,
			zap.NewNop(),
		)
		if err != nil {
			log.Fatalf("cannot init storage factory: %s", err.Error())
		}

		return factory
	}()

	_testUser101 = func() *types.User {
		var userID int64
		err := _testFactory.pool.QueryRow(
			_ctx,
			`select u.id
             from users u
                    join tg_users tg
                         on u.id = tg.user_id
             where tg.id = 101`,
		).Scan(&userID)
		if err != nil {
			log.Fatalf("cannot select test user ID (101): %s", err.Error())
			return nil
		}

		user := types.User(userID)

		return &user
	}()

	_testUser102 = func() *types.User {
		var userID int64
		err := _testFactory.pool.QueryRow(
			_ctx,
			`select u.id
             from users u
                    join tg_users tg
                         on u.id = tg.user_id
             where tg.id = 102`,
		).Scan(&userID)
		if err != nil {
			log.Fatalf("cannot select test user ID (101): %s", err.Error())
			return nil
		}

		user := types.User(userID)

		return &user
	}()
)
