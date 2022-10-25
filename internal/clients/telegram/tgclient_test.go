package telegram

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	tgmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/clients/telegram"
	mmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	smocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

var (
	testUser     = &([]types.User{types.User(int64(123))}[0])
	testTgUserID = int64(123)
	simpleError  = errors.New("error")
)

func newTestCommandMessage(text string) *tgbotapi.Message {
	command := strings.SplitN(text, " ", 2)[0]
	message := &tgbotapi.Message{
		From: &([]tgbotapi.User{{
			ID:       testTgUserID,
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

func Test_client_ListenUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		api        func() api
		storage    func() storage.TelegramUserStorage
		controller func() model.Controller
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no controller",
			fields: fields{
				api: func() api {
					return nil
				},
				storage: func() storage.TelegramUserStorage {
					return nil
				},
			},
			wantErr: true,
		},
		{
			name: "cannot resolve user",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("anything"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(nil, simpleError)
					m.EXPECT().Add(testTgUserID).Return(nil, simpleError)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "unknown command",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/avadakedavra"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("я не знаю такой команды"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(nil, simpleError)
					m.EXPECT().Add(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					return mmocks.NewMockController(ctrl)
				},
			},
			wantErr: false,
		},
		{
			name: "start",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/start"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Привет!"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(nil, simpleError)
					m.EXPECT().Add(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					return mmocks.NewMockController(ctrl)
				},
			},
			wantErr: false,
		},
		{
			name: "currencies list",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/currency"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(gomock.All(
						test.MessageTextContains("Текущая валюта: USD"),
						test.MessageKeyboardContains("EUR"),
						test.MessageKeyboardContains("RUB"),
					))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().ListCurrencies(request.ListCurrencies{
						User: testUser,
					}).Return(response.ListCurrencies{
						Current: "USD",
						List: []string{
							"USD",
							"EUR",
							"RUB",
						},
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "currency switch cannot resolve user",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							CallbackQuery: &tgbotapi.CallbackQuery{
								ID: "some-id",
								From: &tgbotapi.User{
									ID:       testTgUserID,
									UserName: "tester",
								},
								Data: "USD",
							},
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(nil, simpleError)
					m.EXPECT().Add(testTgUserID).Return(nil, simpleError)
					return m
				},
				controller: func() model.Controller {
					return mmocks.NewMockController(ctrl)
				},
			},
			wantErr: false,
		},
		{
			name: "currency switch failed",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							CallbackQuery: &tgbotapi.CallbackQuery{
								ID: "some-id",
								From: &tgbotapi.User{
									ID:       testTgUserID,
									UserName: "tester",
								},
								Data: "EUR",
							},
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Не удалось сменить текущую валюту"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().SetCurrency(request.SetCurrency{
						User: testUser,
						Code: "EUR",
					}).Return(response.SetCurrency(false))
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "currency telegram error no meaning",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							CallbackQuery: &tgbotapi.CallbackQuery{
								ID: "some-id",
								From: &tgbotapi.User{
									ID:       testTgUserID,
									UserName: "tester",
								},
								Data: "EUR",
							},
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					var callback = reflect.TypeOf((*tgbotapi.CallbackConfig)(nil)).Elem()
					m.EXPECT().Request(gomock.AssignableToTypeOf(callback)).Return(nil, simpleError)
					m.EXPECT().Send(test.MessageTextContains("Готово"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().SetCurrency(request.SetCurrency{
						User: testUser,
						Code: "EUR",
					}).Return(response.SetCurrency(true))
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "currency switch success",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							CallbackQuery: &tgbotapi.CallbackQuery{
								From: &tgbotapi.User{
									ID:       testTgUserID,
									UserName: "tester",
								},
								Data: "RUB",
							},
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					var callback = reflect.TypeOf((*tgbotapi.CallbackConfig)(nil)).Elem()
					m.EXPECT().Request(gomock.AssignableToTypeOf(callback)).Return(nil, nil)
					m.EXPECT().Send(test.MessageTextContains("Готово"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().SetCurrency(request.SetCurrency{
						User: testUser,
						Code: "RUB",
					}).Return(response.SetCurrency(true))
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit invalid args error",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit foo"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Не удалось задать лимит"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit controller error",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit 200 "),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Не удалось задать лимит"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(nil, simpleError)
					m.EXPECT().Add(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().SetLimit(request.SetLimit{
						User:     testUser,
						Value:    2000000,
						Category: "",
					}).Return(response.SetLimit(false))
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit set success",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit 250 taxi & coffee "),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Готово"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(nil, simpleError)
					m.EXPECT().Add(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().SetLimit(request.SetLimit{
						User:     testUser,
						Value:    2500000,
						Category: "taxi & coffee",
					}).Return(response.SetLimit(true))
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit render not ready",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit "),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Выполняется обновление курсов валют"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().ListLimits(request.ListLimits{
						User: testUser,
					}).Return(response.ListLimits{
						Ready: false,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit render emergency",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().ListLimits(request.ListLimits{
						User: testUser,
					}).Return(response.ListLimits{
						Ready:   true,
						Success: false,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit render empty",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Лимиты ещё не заданы"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().ListLimits(request.ListLimits{
						User: testUser,
					}).Return(response.ListLimits{
						Ready:           true,
						Success:         true,
						CurrentCurrency: "RUB",
						List:            make(map[string]response.LimitItem),
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit render only base limit",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(gomock.All(
						test.MessageTextContains("Общий лимит (осталось/всего)"),
						test.MessageTextContains("50.00/500.00 RUB (1.00/10.00 USD)"),
					))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().ListLimits(request.ListLimits{
						User: testUser,
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
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "limit render complete",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/limit"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(gomock.All(
						test.MessageTextContains("Твои лимиты (осталось/всего)"),
						test.MessageTextContains("<b>0.00</b>/25.00 EUR (0.00/20.00 USD)"),
						test.MessageTextContains("5.00/10.00 EUR (250.00/500.00 RUB)"),
						test.MessageTextContains("<b>0.00</b>/100.00 EUR"),
					))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().ListLimits(request.ListLimits{
						User: testUser,
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
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "add syntax error",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/add 2d 10 taxi"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Не удалось добавить расход"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					return mmocks.NewMockController(ctrl)
				},
			},
			wantErr: false,
		},
		{
			name: "add invalid date",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/add 99.99.9999 10 taxi"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("дата указана неверно"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					return mmocks.NewMockController(ctrl)
				},
			},
			wantErr: false,
		},
		{
			name: "add not ready",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/add 20.10.2022 10 taxi"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Выполняется обновление курсов валют"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().AddExpense(request.AddExpense{
						User:     testUser,
						Date:     utils.TruncateToDate(time.Date(2022, 10, 20, 0, 0, 0, 0, time.UTC)),
						Amount:   100000,
						Category: "taxi",
					}).Return(response.AddExpense{
						Ready: false,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "add emergency",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/add -10d 2 coffee"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().AddExpense(request.AddExpense{
						User:     testUser,
						Date:     utils.TruncateToDate(time.Now()).Add(-10 * 24 * time.Hour),
						Amount:   20000,
						Category: "coffee",
					}).Return(response.AddExpense{
						Ready:   true,
						Success: false,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "add success",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/add -1d 2,02 coffee"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Готово"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().AddExpense(request.AddExpense{
						User:     testUser,
						Date:     utils.TruncateToDate(time.Now()).Add(-24 * time.Hour),
						Amount:   20200,
						Category: "coffee",
					}).Return(response.AddExpense{
						Ready:   true,
						Success: true,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "add success limit reached",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/add @ 2.5 coffee"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(gomock.All(
						test.MessageTextContains("Готово"),
						test.MessageTextContains("Ты исчерпал заданный лимит"),
					))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					m.EXPECT().AddExpense(request.AddExpense{
						User:     testUser,
						Date:     utils.TruncateToDate(time.Now()),
						Amount:   25000,
						Category: "coffee",
					}).Return(response.AddExpense{
						Ready:        true,
						Success:      true,
						LimitReached: true,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "report invalid syntax",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/report taxi"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Для просмотра расходов"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					return mmocks.NewMockController(ctrl)
				},
			},
			wantErr: false,
		},
		{
			name: "report last week not ready",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/report"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Выполняется обновление курсов валют"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					from := utils.TruncateToDate(time.Now()).Add(-7 * 24 * time.Hour)
					m.EXPECT().GetReport(request.GetReport{
						User: testUser,
						From: from,
					}).Return(response.GetReport{
						From:  from,
						Ready: false,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "report last week explicit emergency",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/report 1w"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("бот временно неисправен"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					from := utils.TruncateToDate(time.Now()).Add(-7 * 24 * time.Hour)
					m.EXPECT().GetReport(request.GetReport{
						User: testUser,
						From: from,
					}).Return(response.GetReport{
						From:    from,
						Ready:   true,
						Success: false,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "report last 2 months no expenses",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/report 2m"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					m.EXPECT().Send(test.MessageTextContains("Ты ещё не добавил ни одного расхода"))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					from := utils.TruncateToDate(time.Now()).Add(-60 * 24 * time.Hour)
					m.EXPECT().GetReport(request.GetReport{
						User: testUser,
						From: from,
					}).Return(response.GetReport{
						From:    from,
						Ready:   true,
						Data:    make(map[string]int64),
						Success: true,
					})
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "report last 3 years complete",
			fields: fields{
				api: func() api {
					m := tgmocks.NewMockapi(ctrl)
					updates := make(chan tgbotapi.Update)
					go func() {
						updates <- tgbotapi.Update{
							Message: newTestCommandMessage("/report 3y"),
						}
					}()
					var cfg = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()
					m.EXPECT().GetUpdatesChan(gomock.AssignableToTypeOf(cfg)).Return(updates)
					from := utils.TruncateToDate(time.Now()).Add(-3 * 365 * 24 * time.Hour)
					m.EXPECT().Send(gomock.All(
						test.MessageTextContains("Расходы с "+from.Format("02.01.2006")+" (валюта — RUB)"),
						test.MessageTextContains("hotel: 5000.00"),
						test.MessageTextContains("кофе: 120.50"),
						test.MessageTextContains("такси: 500.00"),
					))
					return m
				},
				storage: func() storage.TelegramUserStorage {
					m := smocks.NewMockTelegramUserStorage(ctrl)
					m.EXPECT().FetchByID(testTgUserID).Return(testUser, nil)
					return m
				},
				controller: func() model.Controller {
					m := mmocks.NewMockController(ctrl)
					from := utils.TruncateToDate(time.Now()).Add(-3 * 365 * 24 * time.Hour)
					m.EXPECT().GetReport(request.GetReport{
						User: testUser,
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
					return m
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				api:     tt.fields.api(),
				storage: tt.fields.storage(),
			}
			if tt.fields.controller != nil {
				c.RegisterController(tt.fields.controller())
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			if err := c.ListenUpdates(ctx); (err != nil) != tt.wantErr {
				t.Errorf("ListenUpdates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
