package model

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
)

const testUserID = int64(123)

func Test_HandleMessage_StartMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	storage := mocks.NewMockExpenseStorage(ctrl)
	exchanger := mocks.NewMockExchanger(ctrl)
	currencyKeeper := mocks.NewMockCurrencyKeeper(ctrl)

	bot := NewBot(sender, storage, exchanger, currencyKeeper)

	sender.EXPECT().SendMessage(testUserID, test.Contains("Привет!"))
	storage.EXPECT().Init(testUserID)

	err := bot.HandleMessage(dto.Message{
		UserID: testUserID,
		Text:   "/start me",
	})

	assert.NoError(t, err)
}

func Test_HandleMessage_InvalidCommands(t *testing.T) {
	var tt = []struct {
		name string
		text string
		want interface{}
	}{
		{
			name: "unknown command",
			text: "/avadakedavra",
			want: test.Contains("не знаю такой команды"),
		},
		{
			name: "invalid add format",
			text: "/add @@",
			want: test.Contains("Не удалось определить расход"),
		},
		{
			name: "invalid add date",
			text: "/add 30.02.2022 100 milk",
			want: test.Contains("не удалось определить дату"),
		},
		{
			name: "invalid add amount",
			text: "/add @ 100. milk",
			want: test.Contains("Не удалось определить расход"),
		},
		{
			name: "invalid report format",
			text: "/report 100d",
			want: test.Contains("Для просмотра расходов"),
		},
	}

	ctrl := gomock.NewController(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sender := mocks.NewMockMessageSender(ctrl)
			storage := mocks.NewMockExpenseStorage(ctrl)
			exchanger := mocks.NewMockExchanger(ctrl)
			currencyKeeper := mocks.NewMockCurrencyKeeper(ctrl)

			bot := NewBot(sender, storage, exchanger, currencyKeeper)

			sender.EXPECT().SendMessage(testUserID, tc.want)

			err := bot.HandleMessage(dto.Message{
				UserID: testUserID,
				Text:   tc.text,
			})
			assert.NoError(t, err)
		})
	}
}

func Test_HandleMessage_AddExpenseCommand(t *testing.T) {
	type want struct {
		date     interface{}
		amount   int64
		category string
	}

	var tt = []struct {
		name      string
		text      string
		amount    int64
		currency  string
		exchanged int64
		want      want
	}{
		{
			name:      "add today",
			text:      "/add @ 5 milk",
			amount:    int64(50000),
			currency:  "EUR",
			exchanged: int64(1000000),
			want: want{
				date:     test.SameDate(time.Now().UTC()),
				amount:   int64(1000000),
				category: "milk",
			},
		},
		{
			name:      "add 2 days ago",
			text:      "/add -2d 100,50 coffee",
			amount:    int64(1005000),
			currency:  "CNY",
			exchanged: int64(8040000),
			want: want{
				date:     test.SameDate(time.Now().UTC().AddDate(0, 0, -2)),
				amount:   int64(8040000),
				category: "coffee",
			},
		},
		{
			name:      "add by date",
			text:      "/add 09.05.1945 0.99 ice cream",
			amount:    int64(9900),
			currency:  "RUB",
			exchanged: int64(9900),
			want: want{
				date:     test.SameDate(time.Date(1945, 5, 9, 0, 0, 0, 0, time.UTC)),
				amount:   int64(9900),
				category: "ice cream",
			},
		},
	}

	ctrl := gomock.NewController(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sender := mocks.NewMockMessageSender(ctrl)
			storage := mocks.NewMockExpenseStorage(ctrl)
			exchanger := mocks.NewMockExchanger(ctrl)
			currencyKeeper := mocks.NewMockCurrencyKeeper(ctrl)

			bot := NewBot(sender, storage, exchanger, currencyKeeper)

			sender.EXPECT().SendMessage(testUserID, doneMessage)
			storage.EXPECT().Add(testUserID, tc.want.date, tc.want.amount, tc.want.category)
			exchanger.EXPECT().ExchangeToBase(tc.amount, tc.currency).Return(tc.exchanged, nil)
			currencyKeeper.EXPECT().Get(testUserID).Return(tc.currency)

			err := bot.HandleMessage(dto.Message{
				UserID: testUserID,
				Text:   tc.text,
			})
			assert.NoError(t, err)
		})
	}
}

func Test_HandleMessage_ReportCommand(t *testing.T) {
	type want struct {
		exchange map[int64]int64
		message  interface{}
		from     interface{}
	}

	var tt = []struct {
		name          string
		text          string
		list          map[string]int64
		currency      string
		exchangeError bool
		want          want
	}{
		{
			name:          "report empty (week)",
			text:          "/report",
			list:          map[string]int64{},
			currency:      "RUB",
			exchangeError: false,
			want: want{
				message: "Вы ещё не добавили ни одного расхода.",
				from:    test.SameDate(time.Now().AddDate(0, 0, -7)),
			},
		},
		{
			name: "report week",
			text: "/report",
			list: map[string]int64{
				"кофе":   1250000,
				"молоко": 650000,
			},
			currency:      "USD",
			exchangeError: false,
			want: want{
				exchange: map[int64]int64{
					1250000: 50000,
					650000:  26000,
				},
				message: test.Contains("Расходы с " + time.Now().AddDate(0, 0, -7).Format("02.01.2006") + " (валюта — USD):\nкофе: 5.00\nмолоко: 2.60"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -7)),
			},
		},
		{
			name: "report 3 weeks",
			text: "/report 3w",
			list: map[string]int64{
				"мебель": 125000000,
			},
			currency:      "USD",
			exchangeError: false,
			want: want{
				exchange: map[int64]int64{
					125000000: 5000000,
				},
				message: test.Contains("мебель: 500.00"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -21)),
			},
		},
		{
			name: "report 2 months",
			text: "/report 2m",
			list: map[string]int64{
				"кофе": 1000000,
			},
			currency:      "EUR",
			exchangeError: false,
			want: want{
				exchange: map[int64]int64{
					1000000: 20000,
				},
				message: test.Contains("кофе: 2.00"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -60)),
			},
		},
		{
			name: "report year",
			text: "/report y",
			list: map[string]int64{
				"кофе": 1000000,
			},
			currency:      "CNY",
			exchangeError: false,
			want: want{
				exchange: map[int64]int64{
					1000000: 125000,
				},
				message: test.Contains("кофе: 12.50"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -365)),
			},
		},
		{
			name: "unknown currency",
			text: "/report",
			list: map[string]int64{
				"кофе": 1000000,
			},
			currency:      "AUD",
			exchangeError: true,
			want: want{
				message: test.Contains("Ошибка при формировании отчёта"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -7)),
			},
		},
	}

	ctrl := gomock.NewController(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sender := mocks.NewMockMessageSender(ctrl)
			storage := mocks.NewMockExpenseStorage(ctrl)
			exchanger := mocks.NewMockExchanger(ctrl)
			currencyKeeper := mocks.NewMockCurrencyKeeper(ctrl)

			bot := NewBot(sender, storage, exchanger, currencyKeeper)

			sender.EXPECT().SendMessage(testUserID, tc.want.message)
			storage.EXPECT().List(testUserID, tc.want.from).Return(tc.list)

			if len(tc.list) > 0 {
				currencyKeeper.EXPECT().Get(testUserID).Return(tc.currency)
			}

			if tc.exchangeError {
				exchanger.EXPECT().ExchangeFromBase(gomock.Any(), tc.currency).Return(int64(0), errors.New("some error"))
			} else if len(tc.want.exchange) > 0 {
				for value, result := range tc.want.exchange {
					exchanger.EXPECT().ExchangeFromBase(value, tc.currency).Return(result, nil)
				}
			}

			err := bot.HandleMessage(dto.Message{
				UserID: testUserID,
				Text:   tc.text,
			})
			assert.NoError(t, err)
		})
	}
}

func Test_HandleMessage_CurrencyCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	storage := mocks.NewMockExpenseStorage(ctrl)
	exchanger := mocks.NewMockExchanger(ctrl)
	currencyKeeper := mocks.NewMockCurrencyKeeper(ctrl)

	bot := NewBot(sender, storage, exchanger, currencyKeeper)

	sender.EXPECT().SendMessageWithInlineKeyboard(
		testUserID,
		gomock.All(
			test.Contains(currencyCurrentMessage),
			test.Contains("RUB"),
			test.Contains(currencyChooseMessage),
		),
		gomock.Any(),
	)
	exchanger.EXPECT().ListCurrencies().Return([]string{})
	currencyKeeper.EXPECT().Get(testUserID).Return("RUB")

	err := bot.HandleMessage(dto.Message{
		UserID: testUserID,
		Text:   "/currency",
	})

	assert.NoError(t, err)
}

func Test_HandleCallbackQuery_Currency(t *testing.T) {
	tt := []struct {
		name            string
		data            string
		exchangerReady  bool
		unknownCurrency bool
		want            interface{}
	}{
		{
			name:            "unknown currency",
			data:            "AUD",
			exchangerReady:  true,
			unknownCurrency: true,
			want:            test.Contains("Не удалось сменить текущую валюту."),
		},
		{
			name:            "known currency",
			data:            "USD",
			exchangerReady:  true,
			unknownCurrency: false,
			want:            doneMessage,
		},
		{
			name:            "rates refresh in progress",
			data:            "USD",
			exchangerReady:  false,
			unknownCurrency: false,
			want:            currencyLaterMessage,
		},
	}

	ctrl := gomock.NewController(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sender := mocks.NewMockMessageSender(ctrl)
			storage := mocks.NewMockExpenseStorage(ctrl)
			exchanger := mocks.NewMockExchanger(ctrl)
			currencyKeeper := mocks.NewMockCurrencyKeeper(ctrl)

			bot := NewBot(sender, storage, exchanger, currencyKeeper)

			sender.EXPECT().SendMessage(testUserID, tc.want)
			exchanger.EXPECT().Ready().Return(tc.exchangerReady)
			if tc.exchangerReady {
				if tc.unknownCurrency {
					currencyKeeper.EXPECT().Set(testUserID, tc.data).Return(errors.New("some error"))
				} else {
					currencyKeeper.EXPECT().Set(testUserID, tc.data).Return(nil)
				}
			}

			err := bot.HandleCallbackQuery(dto.CallbackQuery{
				UserID: testUserID,
				Data:   tc.data,
			})
			assert.NoError(t, err)
		})
	}
}
