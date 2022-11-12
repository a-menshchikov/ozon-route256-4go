package telegram

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	tgmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/clients/telegram"
	mmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	smocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
)

type clientMocksInitializer struct {
	api        func(m *tgmocks.Mockapi)
	storage    func(m *smocks.MockTelegramUserStorage)
	controller func(m *mmocks.MockController)
}

func newTestCommandMessage(text string) *tgbotapi.Message {
	command := strings.SplitN(text, " ", 2)[0]
	message := &tgbotapi.Message{
		From: &([]tgbotapi.User{{
			ID:       test.TgUserID,
			UserName: "tester",
		}}[0]),
		Text: text,
		Entities: []tgbotapi.MessageEntity{{
			Type:   "bot_command",
			Offset: 0,
			Length: len(command),
		}},
	}

	return message
}

func setupClient(t *testing.T, i clientMocksInitializer) (*client, context.Context, context.CancelFunc) {
	ctrl := gomock.NewController(t)

	apiMock := tgmocks.NewMockapi(ctrl)
	if i.api != nil {
		i.api(apiMock)
	}

	storageMock := smocks.NewMockTelegramUserStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	c := &client{
		api:     apiMock,
		storage: storageMock,
		logger:  zap.NewNop(),
	}

	if i.controller != nil {
		controllerMock := mmocks.NewMockController(ctrl)
		i.controller(controllerMock)
		c.RegisterController(controllerMock)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

	return c, ctx, cancel
}

func Test_client_ListenUpdates(t *testing.T) {
	t.Run("no controller", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("cannot resolve user", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("anything"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("unknown command", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/avadakedavra"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("я не знаю такой команды"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("start", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/start"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Привет!"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("currencies list", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/currency"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(gomock.All(
					test.MessageTextContains("Текущая валюта: USD"),
					test.MessageKeyboardContains("EUR"),
					test.MessageKeyboardContains("RUB"),
				))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().ListCurrencies(gomock.AssignableToTypeOf(test.CtxInterface), request.ListCurrencies{
					User: test.User,
				}).Return(response.ListCurrencies{
					Current: "USD",
					List: []string{
						"USD",
						"EUR",
						"RUB",
					},
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("currency switch cannot resolve user", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						CallbackQuery: &tgbotapi.CallbackQuery{
							ID: "some-id",
							From: &tgbotapi.User{
								ID:       test.TgUserID,
								UserName: "tester",
							},
							Data: "USD",
						},
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("currency switch failed", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						CallbackQuery: &tgbotapi.CallbackQuery{
							ID: "some-id",
							From: &tgbotapi.User{
								ID:       test.TgUserID,
								UserName: "tester",
							},
							Data: "EUR",
						},
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Не удалось сменить текущую валюту"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().SetCurrency(gomock.AssignableToTypeOf(test.CtxInterface), request.SetCurrency{
					User: test.User,
					Code: "EUR",
				}).Return(response.SetCurrency(false))
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("currency telegram error no meaning", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						CallbackQuery: &tgbotapi.CallbackQuery{
							ID: "some-id",
							From: &tgbotapi.User{
								ID:       test.TgUserID,
								UserName: "tester",
							},
							Data: "EUR",
						},
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				var callback = reflect.TypeOf((*tgbotapi.CallbackConfig)(nil)).Elem()
				m.EXPECT().Request(gomock.AssignableToTypeOf(callback)).Return(nil, test.SimpleError)
				m.EXPECT().Send(test.MessageTextContains("Готово"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().SetCurrency(gomock.AssignableToTypeOf(test.CtxInterface), request.SetCurrency{
					User: test.User,
					Code: "EUR",
				}).Return(response.SetCurrency(true))
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("currency switch success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						CallbackQuery: &tgbotapi.CallbackQuery{
							From: &tgbotapi.User{
								ID:       test.TgUserID,
								UserName: "tester",
							},
							Data: "RUB",
						},
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				var callback = reflect.TypeOf((*tgbotapi.CallbackConfig)(nil)).Elem()
				m.EXPECT().Request(gomock.AssignableToTypeOf(callback)).Return(nil, nil)
				m.EXPECT().Send(test.MessageTextContains("Готово"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().SetCurrency(gomock.AssignableToTypeOf(test.CtxInterface), request.SetCurrency{
					User: test.User,
					Code: "RUB",
				}).Return(response.SetCurrency(true))
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit invalid args error", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit foo"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Не удалось задать лимит"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit controller error", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit 200 "),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Не удалось задать лимит"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().SetLimit(gomock.AssignableToTypeOf(test.CtxInterface), request.SetLimit{
					User:     test.User,
					Value:    2000000,
					Category: "",
				}).Return(response.SetLimit(false))
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit set success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit 250 taxi & coffee "),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Готово"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(nil, test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().SetLimit(gomock.AssignableToTypeOf(test.CtxInterface), request.SetLimit{
					User:     test.User,
					Value:    2500000,
					Category: "taxi & coffee",
				}).Return(response.SetLimit(true))
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit render not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit "),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Выполняется обновление курсов валют"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().ListLimits(gomock.AssignableToTypeOf(test.CtxInterface), request.ListLimits{
					User: test.User,
				}).Return(response.ListLimits{
					Ready: false,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit render emergency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().ListLimits(gomock.AssignableToTypeOf(test.CtxInterface), request.ListLimits{
					User: test.User,
				}).Return(response.ListLimits{
					Ready:   true,
					Success: false,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit render empty", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Лимиты ещё не заданы"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().ListLimits(gomock.AssignableToTypeOf(test.CtxInterface), request.ListLimits{
					User: test.User,
				}).Return(response.ListLimits{
					Ready:           true,
					Success:         true,
					CurrentCurrency: "RUB",
					List:            make(map[string]response.LimitItem),
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit render only base limit", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(gomock.All(
					test.MessageTextContains("Общий лимит (осталось/всего)"),
					test.MessageTextContains("50.00/500.00 RUB (1.00/10.00 USD)"),
				))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().ListLimits(gomock.AssignableToTypeOf(test.CtxInterface), request.ListLimits{
					User: test.User,
				}).Return(response.ListLimits{
					Ready:           true,
					Success:         true,
					CurrentCurrency: "RUB",
					List: map[string]response.LimitItem{
						"": {
							Total:   5000000,
							Remains: 500000,
							Origin: types.LimitItem{
								Total:    100000,
								Remains:  10000,
								Currency: "USD",
							},
						},
					},
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("limit render complete", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/limit"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(gomock.All(
					test.MessageTextContains("Твои лимиты (осталось/всего)"),
					test.MessageTextContains("<b>0.00</b>/25.00 EUR (0.00/20.00 USD)"),
					test.MessageTextContains("5.00/10.00 EUR (250.00/500.00 RUB)"),
					test.MessageTextContains("<b>0.00</b>/100.00 EUR"),
				))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().ListLimits(gomock.AssignableToTypeOf(test.CtxInterface), request.ListLimits{
					User: test.User,
				}).Return(response.ListLimits{
					Ready:           true,
					Success:         true,
					CurrentCurrency: "EUR",
					List: map[string]response.LimitItem{
						"taxi": {
							Total:   250000,
							Remains: 0,
							Origin: types.LimitItem{
								Total:    200000,
								Remains:  0,
								Currency: "USD",
							},
						},
						"coffee": {
							Total:   100000,
							Remains: 50000,
							Origin: types.LimitItem{
								Total:    5000000,
								Remains:  2500000,
								Currency: "RUB",
							},
						},
						"": {
							Total:   1000000,
							Remains: 0,
							Origin: types.LimitItem{
								Total:    1000000,
								Remains:  0,
								Currency: "EUR",
							},
						},
					},
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("add syntax error", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/add 2d 10 taxi"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Не удалось добавить расход"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("add invalid date", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/add 99.99.9999 10 taxi"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("дата указана неверно"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("add not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/add 20.10.2022 10 taxi"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Выполняется обновление курсов валют"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), request.AddExpense{
					User:     test.User,
					Date:     utils.TruncateToDate(time.Date(2022, 10, 20, 0, 0, 0, 0, time.UTC)),
					Amount:   100000,
					Category: "taxi",
				}).Return(response.AddExpense{
					Ready: false,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("add emergency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/add -10d 2 coffee"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), request.AddExpense{
					User:     test.User,
					Date:     utils.TruncateToDate(time.Now()).Add(-10 * 24 * time.Hour),
					Amount:   20000,
					Category: "coffee",
				}).Return(response.AddExpense{
					Ready:   true,
					Success: false,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("add success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/add -1d 2,02 coffee"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Готово"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), request.AddExpense{
					User:     test.User,
					Date:     utils.TruncateToDate(time.Now()).Add(-24 * time.Hour),
					Amount:   20200,
					Category: "coffee",
				}).Return(response.AddExpense{
					Ready:   true,
					Success: true,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("add success limit reached", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/add @ 2.5 coffee"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(gomock.All(
					test.MessageTextContains("Готово"),
					test.MessageTextContains("Ты исчерпал заданный лимит"),
				))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), request.AddExpense{
					User:     test.User,
					Date:     utils.TruncateToDate(time.Now()),
					Amount:   25000,
					Category: "coffee",
				}).Return(response.AddExpense{
					Ready:        true,
					Success:      true,
					LimitReached: true,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("report invalid syntax", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/report taxi"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Для просмотра расходов"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(*mmocks.MockController) {},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("report last week not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/report"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Выполняется обновление курсов валют"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				from := utils.TruncateToDate(time.Now()).Add(-7 * 24 * time.Hour)
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), request.GetReport{
					User: test.User,
					From: from,
				}).Return(response.GetReport{
					From:  from,
					Ready: false,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("report last week explicit emergency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/report 1w"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Не удалось сформировать отчёт"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				from := utils.TruncateToDate(time.Now()).Add(-7 * 24 * time.Hour)
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), request.GetReport{
					User: test.User,
					From: from,
				}).Return(response.GetReport{
					From:    from,
					Ready:   true,
					Success: false,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("report last 2 months no expenses", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/report 2m"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				m.EXPECT().Send(test.MessageTextContains("Ты ещё не добавил ни одного расхода"))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				from := utils.TruncateToDate(time.Now()).Add(-60 * 24 * time.Hour)
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), request.GetReport{
					User: test.User,
					From: from,
				}).Return(response.GetReport{
					From:    from,
					Ready:   true,
					Data:    make(map[string]int64),
					Success: true,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("report last 3 years complete", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		c, ctx, cancel := setupClient(t, clientMocksInitializer{
			api: func(m *tgmocks.Mockapi) {
				updates := make(chan tgbotapi.Update)
				go func() {
					updates <- tgbotapi.Update{
						Message: newTestCommandMessage("/report 3y"),
					}
				}()
				m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(test.UpdateConfigInterface)).Return(updates)
				from := utils.TruncateToDate(time.Now()).Add(-3 * 365 * 24 * time.Hour)
				m.EXPECT().Send(gomock.All(
					test.MessageTextContains("Расходы с "+from.Format("02.01.2006")+" (валюта — RUB)"),
					test.MessageTextContains("hotel: 5000.00"),
					test.MessageTextContains("кофе: 120.50"),
					test.MessageTextContains("такси: 500.00"),
				))
			},
			storage: func(m *smocks.MockTelegramUserStorage) {
				m.EXPECT().FetchByID(gomock.AssignableToTypeOf(test.CtxInterface), test.TgUserID).Return(test.User, nil)
			},
			controller: func(m *mmocks.MockController) {
				from := utils.TruncateToDate(time.Now()).Add(-3 * 365 * 24 * time.Hour)
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), request.GetReport{
					User: test.User,
					From: from,
				}).Return(response.GetReport{
					From:     from,
					Ready:    true,
					Currency: "RUB",
					Data: map[string]int64{
						"такси": 5000000,
						"кофе":  1205000,
						"hotel": 50000000,
					},
					Success: true,
				})
			},
		})
		defer cancel()

		// ACT
		err := c.ListenUpdates(ctx)

		// ASSERT
		assert.NoError(t, err)
	})
}
