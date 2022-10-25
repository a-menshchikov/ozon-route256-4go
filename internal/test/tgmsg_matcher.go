package test

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageTextContainsMatcher struct {
	s string
}

type MessageKeyboardContainsMatcher struct {
	s string
}

func MessageTextContains(s string) MessageTextContainsMatcher {
	return MessageTextContainsMatcher{s}
}

func MessageKeyboardContains(s string) MessageKeyboardContainsMatcher {
	return MessageKeyboardContainsMatcher{s}
}

func (m MessageTextContainsMatcher) Matches(x interface{}) bool {
	msg, ok := x.(tgbotapi.MessageConfig)
	if !ok {
		return false
	}

	return strings.Contains(msg.Text, m.s)
}

func (m MessageKeyboardContainsMatcher) Matches(x interface{}) bool {
	msg, ok := x.(tgbotapi.MessageConfig)
	if !ok {
		return false
	}

	markup := msg.ReplyMarkup
	keyboard, ok := markup.(tgbotapi.InlineKeyboardMarkup)
	if !ok {
		return false
	}

	for _, bb := range keyboard.InlineKeyboard {
		for _, b := range bb {
			if strings.Contains(b.Text, m.s) {
				return true
			}
		}
	}

	return false
}

func (m MessageTextContainsMatcher) String() string {
	return fmt.Sprintf("contains %v (%T)", m.s, m.s)
}

func (m MessageKeyboardContainsMatcher) String() string {
	return fmt.Sprintf("contains %v (%T)", m.s, m.s)
}
