package tg

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto"
)

type RequestHandler interface {
	HandleMessage(message dto.Message) error
	HandleCallbackQuery(query dto.CallbackQuery) error
}

type Client struct {
	client *tgbotapi.BotAPI
}

func New(token string) (*Client, error) {
	c, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &Client{
		client: c,
	}, nil
}

func (c *Client) SendMessage(userID int64, text string) error {
	message := tgbotapi.NewMessage(userID, text)
	message.ParseMode = tgbotapi.ModeHTML

	_, err := c.client.Send(message)
	if err != nil {
		return errors.Wrap(err, "client.Send")
	}

	return nil
}

func (c *Client) SendMessageWithInlineKeyboard(userID int64, text string, rowsData [][][]string) error {
	message := tgbotapi.NewMessage(userID, text)
	message.ParseMode = tgbotapi.ModeHTML

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, rowData := range rowsData {
		var row []tgbotapi.InlineKeyboardButton
		for j, button := range rowData {
			if len(button) != 2 {
				return fmt.Errorf("invalid keyboard button (row %d, button %d)", i, j)
			}

			row = append(row, tgbotapi.NewInlineKeyboardButtonData(button[0], button[1]))
		}
		rows = append(rows, row)

	}

	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	_, err := c.client.Send(message)
	if err != nil {
		return errors.Wrap(err, "client.Send (with inline keyboard)")
	}

	return nil
}

func (c *Client) ListenUpdates(ctx context.Context, handler RequestHandler) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Println("listening for messages")

	updates := c.client.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return

		case update := <-updates:
			if update.Message != nil {
				c.handleMessage(update.Message, handler)
			} else if update.CallbackQuery != nil {
				callbackQuery := update.CallbackQuery
				c.handleCallback(callbackQuery, handler)
			}
		}
	}
}

func (c *Client) handleMessage(message *tgbotapi.Message, handler RequestHandler) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	err := handler.HandleMessage(dto.Message{
		UserID: message.From.ID,
		Text:   message.Text,
	})
	if err != nil {
		log.Println("error processing message: ", err)
	}
}

func (c *Client) handleCallback(callbackQuery *tgbotapi.CallbackQuery, handler RequestHandler) {
	log.Printf("[%s] callback: %s", callbackQuery.From.UserName, callbackQuery.Data)

	callback := tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data)
	if _, err := c.client.Request(callback); err != nil {
		log.Println("error processing callback: ", err)
	}

	err := handler.HandleCallbackQuery(dto.CallbackQuery{
		UserID: callbackQuery.From.ID,
		Data:   callbackQuery.Data,
	})
	if err != nil {
		log.Println("error processing callback query: ", err)
	}
}
