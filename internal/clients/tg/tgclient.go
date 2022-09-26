package tg

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
)

type TokenGetter interface {
	Token() string
}

type Client struct {
	client *tgbotapi.BotAPI
}

func New(tokenGetter TokenGetter) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(tokenGetter.Token())
	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) SendMessage(userID int64, text string) error {
	message := tgbotapi.NewMessage(userID, text)
	message.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := c.client.Send(message)
	if err != nil {
		return errors.Wrap(err, "client.Send")
	}

	return nil
}

func (c *Client) ListenUpdates(ctx context.Context, bot *model.Bot) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Println("listening for messages")

	updates := c.client.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return

		case update := <-updates:
			if update.Message == nil {
				continue
			}

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			err := bot.HandleMessage(dto.Message{
				Text:   update.Message.Text,
				UserID: update.Message.From.ID,
			})
			if err != nil {
				log.Println("error processing message: ", err)
			}
		}
	}
}
