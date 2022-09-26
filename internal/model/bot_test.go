package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
)

type containsMatcher struct {
	s string
}

func contains(s string) containsMatcher {
	return containsMatcher{s}
}

func (c containsMatcher) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}

	return strings.Contains(s, c.s)
}

func (c containsMatcher) String() string {
	return fmt.Sprintf("contains %v (%T)", c.s, c.s)
}

func Test_OnStartCommand_ShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	storage := mocks.NewMockExpenseStorage(ctrl)
	bot := NewBot(sender, storage)

	sender.EXPECT().SendMessage(int64(123), contains("Привет\\!"))
	storage.EXPECT().Init(int64(123))

	err := bot.HandleMessage(dto.Message{
		Text:   "/start me",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnStartCommand_ShouldAnswerWithUnknownCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	storage := mocks.NewMockExpenseStorage(ctrl)
	bot := NewBot(sender, storage)

	sender.EXPECT().SendMessage(int64(777), contains("не знаю такой команды"))

	err := bot.HandleMessage(dto.Message{
		Text:   "/avadakedavra",
		UserID: 777,
	})

	assert.NoError(t, err)
}
