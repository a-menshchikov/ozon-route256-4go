package model

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
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
	bot := NewBot(sender, storage)

	sender.EXPECT().SendMessage(testUserID, test.Contains("Привет!"))
	storage.EXPECT().Init(testUserID)

	err := bot.HandleMessage(dto.Message{
		Text:   "/start me",
		UserID: testUserID,
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

			bot := NewBot(sender, storage)

			sender.EXPECT().SendMessage(testUserID, tc.want)

			err := bot.HandleMessage(dto.Message{
				Text:   tc.text,
				UserID: testUserID,
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
		name string
		text string
		want want
	}{
		{
			name: "add today",
			text: "/add @ 100 milk",
			want: want{
				date:     test.SameDate(time.Now()),
				amount:   int64(10000),
				category: "milk",
			},
		},
		{
			name: "add 2 days ago",
			text: "/add -2d 100,50 coffee",
			want: want{
				date:     test.SameDate(time.Now().AddDate(0, 0, -2)),
				amount:   int64(10050),
				category: "coffee",
			},
		},
		{
			name: "add by date",
			text: "/add 09.05.1945 0.99 ice cream",
			want: want{
				date:     test.SameDate(time.Date(1945, 5, 9, 0, 0, 0, 0, time.UTC)),
				amount:   int64(99),
				category: "ice cream",
			},
		},
	}

	ctrl := gomock.NewController(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sender := mocks.NewMockMessageSender(ctrl)
			storage := mocks.NewMockExpenseStorage(ctrl)

			bot := NewBot(sender, storage)

			sender.EXPECT().SendMessage(testUserID, "Готово!")
			storage.EXPECT().Add(testUserID, tc.want.date, tc.want.amount, tc.want.category)

			err := bot.HandleMessage(dto.Message{
				Text:   tc.text,
				UserID: testUserID,
			})
			assert.NoError(t, err)
		})
	}
}

func Test_HandleMessage_ReportCommand(t *testing.T) {
	type want struct {
		message interface{}
		from    interface{}
	}

	var tt = []struct {
		name string
		text string
		list []interface{}
		want want
	}{
		{
			name: "report empty (week)",
			text: "/report",
			list: []interface{}{
				map[string]int64{},
			},
			want: want{
				message: "Вы ещё не добавили ни одного расхода.",
				from:    test.SameDate(time.Now().AddDate(0, 0, -7)),
			},
		},
		{
			name: "report week",
			text: "/report",
			list: []interface{}{
				map[string]int64{
					"кофе":   12500,
					"молоко": 6500,
				},
			},
			want: want{
				message: test.Contains("Расходы с " + time.Now().AddDate(0, 0, -7).Format("02.01.2006") + ":\nкофе: 125.00\nмолоко: 65.00"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -7)),
			},
		},
		{
			name: "report 3 weeks",
			text: "/report 3w",
			list: []interface{}{
				map[string]int64{
					"мебель": 1250000,
				},
			},
			want: want{
				message: test.Contains("мебель: 12500.00"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -21)),
			},
		},
		{
			name: "report 2 months",
			text: "/report 2m",
			list: []interface{}{
				map[string]int64{
					"кофе": 10000,
				},
			},
			want: want{
				message: test.Contains("кофе: 100.00"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -60)),
			},
		},
		{
			name: "report year",
			text: "/report y",
			list: []interface{}{
				map[string]int64{
					"кофе": 10000,
				},
			},
			want: want{
				message: test.Contains("кофе: 100.00"),
				from:    test.SameDate(time.Now().AddDate(0, 0, -365)),
			},
		},
	}

	ctrl := gomock.NewController(t)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sender := mocks.NewMockMessageSender(ctrl)
			storage := mocks.NewMockExpenseStorage(ctrl)

			bot := NewBot(sender, storage)

			sender.EXPECT().SendMessage(testUserID, tc.want.message)
			storage.EXPECT().List(testUserID, tc.want.from).Return(tc.list...)

			err := bot.HandleMessage(dto.Message{
				Text:   tc.text,
				UserID: testUserID,
			})
			assert.NoError(t, err)
		})
	}
}
