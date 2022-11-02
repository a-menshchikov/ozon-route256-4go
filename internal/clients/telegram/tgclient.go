package telegram

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

const (
	_updateTimeout = 60
	_buttonsPerRow = 4
)

var (
	_addRx    = regexp.MustCompile(`^(|@|-\d+d|\d{2}\.\d{2}\.\d{4})\s*(\d+(?:[.,]\d+)?) (.+)$`)
	_reportRx = regexp.MustCompile(`^(?:(\d+)([wmy]))?$`)

	errWrongExpenseDate    = errors.New("не удалось определить дату")
	errWrongExpenseAmount  = errors.New("не удалось определить сумму")
	errWrongReportDuration = errors.New("не удалось определить срок формирования отчёта")
)

type api interface {
	GetUpdatesChan(tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type client struct {
	api        api
	storage    storage.TelegramUserStorage
	controller model.Controller
}

func NewClient(token string, s storage.TelegramUserStorage) (*client, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &client{
		api:     api,
		storage: s,
	}, nil
}

func (c *client) RegisterController(handler model.Controller) {
	c.controller = handler
}

func (c *client) ListenUpdates(ctx context.Context) error {
	if c.controller == nil {
		return errors.New("register controller first")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = _updateTimeout

	log.Println("listening for messages")

	updates := c.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return nil

		case update := <-updates:
			if update.Message != nil {
				c.handleMessage(update.Message)
			} else if update.CallbackQuery != nil {
				c.handleCallback(update.CallbackQuery)
			}
		}
	}
}

func (c *client) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] message: %s", message.From.UserName, message.Text)

	user, err := c.resolveUser(message.From)
	if err != nil {
		log.Println("cannot resolve user:", err)
		c.sendMessage(message.From.ID, emergencyMessage)
		return
	}

	command := message.Command()
	text := unknownCommandMessage

	if command == "currency" {
		text, keyboard := c.handleCurrency(user)
		c.sendMessageWithInlineKeyboard(message.From.ID, text, keyboard)
		return
	}

	handler, ok := map[string]func(*types.User, string) string{
		"start":  func(*types.User, string) string { return helloMessage },
		"limit":  c.handleLimit,
		"add":    c.handleAdd,
		"report": c.handleReport,
	}[command]

	if ok {
		text = handler(user, strings.TrimSpace(message.CommandArguments()))
	}

	c.sendMessage(message.From.ID, text)
}

func (c *client) handleCallback(callbackQuery *tgbotapi.CallbackQuery) {
	log.Printf("[%s] callback: %s", callbackQuery.From.UserName, callbackQuery.Data)

	user, err := c.resolveUser(callbackQuery.From)
	if err != nil {
		log.Println("cannot resolve user:", err)
		c.sendMessage(callbackQuery.From.ID, emergencyMessage)
		return
	}

	text, ok := c.handleCurrencyCallback(user, callbackQuery.Data)
	if ok {
		callback := tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data)
		if _, err := c.api.Request(callback); err != nil {
			log.Println("error processing callback: ", err)
		}
	}

	c.sendMessage(callbackQuery.From.ID, text)
}

func (c *client) handleCurrency(user *types.User) (string, [][][]string) {
	resp := c.controller.ListCurrencies(request.ListCurrencies{
		User: user,
	})

	return currencyCurrentMessage + resp.Current + "\n\n" + currencyChooseMessage, prepareCurrenciesKeyboard(resp.List)
}

func (c *client) handleCurrencyCallback(user *types.User, currency string) (string, bool) {
	if c.controller.SetCurrency(request.SetCurrency{
		User: user,
		Code: currency,
	}) {
		return doneMessage, true
	}

	return errorMessage(nil, "Не удалось сменить текущую валюту.", currencyHelpMessage), false
}

func (c *client) handleLimit(user *types.User, args string) string {
	if args == "" {
		return renderLimits(c.controller.ListLimits(request.ListLimits{
			User: user,
		}))
	}

	limitStr, category, ok := strings.Cut(args, " ")
	if !ok {
		limitStr, category = args, ""
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err == nil && c.controller.SetLimit(request.SetLimit{
		User:     user,
		Value:    limit * 10000,
		Category: strings.TrimSpace(category),
	}) {
		return doneMessage
	}

	return errorMessage(nil, "Не удалось задать лимит.", limitsHelpMessage)
}

func renderLimits(resp response.ListLimits) string {
	switch {
	case !resp.Ready:
		return currencyLaterMessage

	case !resp.Success:
		return emergencyMessage

	case len(resp.List) == 0:
		return limitsEmptyMessage
	}

	baseItem, baseOk := resp.List[""]
	if baseOk && len(resp.List) == 1 {
		return "Общий лимит (осталось/всего):\n• " + renderLimitRow(baseItem, resp.CurrentCurrency)
	}

	categories := make([]string, 0, len(resp.List))
	for category := range resp.List {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	text := "Твои лимиты (осталось/всего):"
	for category, item := range resp.List {
		if category == "" {
			continue
		}
		text += "\n• " + category + ": " + renderLimitRow(item, resp.CurrentCurrency)
	}

	if baseOk {
		text += "\n• остальные расходы: " + renderLimitRow(baseItem, resp.CurrentCurrency)
	}

	return text
}

func renderLimitRow(item response.LimitItem, currency string) (row string) {
	if item.Remains == 0 {
		row = fmt.Sprintf("<b>%.2f</b>/%.2f %s", float64(item.Remains)/10000, float64(item.Total)/10000, currency)
	} else {
		row = fmt.Sprintf("%.2f/%.2f %s", float64(item.Remains)/10000, float64(item.Total)/10000, currency)
	}

	if item.Origin.Currency != currency {
		row += fmt.Sprintf(" (%.2f/%.2f %s)", float64(item.Origin.Remains)/10000, float64(item.Origin.Total)/10000, item.Origin.Currency)
	}

	return
}

func (c *client) handleAdd(user *types.User, args string) string {
	m := _addRx.FindStringSubmatch(args)
	if len(m) == 0 {
		return errorMessage(nil, "Не удалось добавить расход.", addHelpMessage)
	}

	date, amount, category, err := parseAddArgs(m[1:])
	if err == nil {
		resp := c.controller.AddExpense(request.AddExpense{
			User:     user,
			Date:     date,
			Amount:   amount,
			Category: category,
		})

		switch {
		case !resp.Ready:
			return currencyLaterMessage

		case !resp.Success:
			return emergencyMessage

		case resp.LimitReached:
			return doneMessage + "\n\n" + limitReached

		default:
			return doneMessage
		}
	}

	return errorMessage(err, "Не удалось добавить расход.", addHelpMessage)
}

func parseAddArgs(args []string) (date time.Time, amount int64, category string, err error) {
	if date, err = parseDate(args[0]); err != nil {
		return time.Time{}, 0, "", errors.New("дата указана неверно")
	}

	date = utils.TruncateToDate(date)

	floatAmount, err := strconv.ParseFloat(strings.ReplaceAll(args[1], ",", "."), 64)
	if err != nil {
		err = errWrongExpenseAmount
	}

	amount = int64(floatAmount * 10000)

	category = strings.TrimSpace(args[2])

	return
}

func parseDate(input string) (time.Time, error) {
	if input == "" || input == "@" {
		return time.Now(), nil
	}

	if input[0] == '-' {
		rate, err := strconv.ParseUint(input[1:len(input)-1], 10, 64)
		if err != nil {
			return time.Time{}, errWrongExpenseDate
		}

		return time.Now().Add(-time.Duration(rate) * 24 * time.Hour), nil
	}

	date, err := time.Parse("02.01.2006", input)
	if err != nil {
		return time.Time{}, errWrongExpenseDate
	}

	return date, nil
}

func (c *client) handleReport(user *types.User, args string) string {
	var (
		from time.Time
		err  error
	)

	if args == "" {
		from = utils.TruncateToDate(time.Now()).Add(-7 * 24 * time.Hour)
	} else if m := _reportRx.FindStringSubmatch(args); len(m) == 0 {
		return reportHelpMessage
	} else if from, err = parseReportArgs(m[1:]); err != nil {
		return errorMessage(err, "Не удалось сформировать отчёт.", reportHelpMessage)
	}

	resp := c.controller.GetReport(request.GetReport{
		User: user,
		From: from,
	})

	switch {
	case !resp.Ready:
		return currencyLaterMessage

	case !resp.Success:
		return emergencyMessage

	case len(resp.Data) == 0:
		return reportNoExpenses
	}

	categories := make([]string, 0, len(resp.Data))
	for category := range resp.Data {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	text := fmt.Sprintf("Расходы с %s (валюта — %s):\n", resp.From.Local().Format("02.01.2006"), resp.Currency)
	for _, category := range categories {
		text += fmt.Sprintf("%s: %.2f\n", category, float64(resp.Data[category])/10000)
	}

	return text
}

func parseReportArgs(args []string) (time.Time, error) {
	hours, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return time.Time{}, errWrongReportDuration
	}

	switch args[1] {
	case "w":
		hours *= 24 * 7
	case "m":
		hours *= 24 * 30
	case "y":
		hours *= 24 * 365
	}

	return utils.TruncateToDate(time.Now()).Add(-(time.Duration(hours) * time.Hour)), nil
}

func (c *client) sendMessage(chatID int64, text string) {
	message := tgbotapi.NewMessage(chatID, text)
	message.ParseMode = tgbotapi.ModeHTML

	_, err := c.api.Send(message)
	if err != nil {
		log.Println("cannot send telegram message:", err)
	}
}

func (c *client) sendMessageWithInlineKeyboard(chatID int64, text string, rowsData [][][]string) {
	message := tgbotapi.NewMessage(chatID, text)
	message.ParseMode = tgbotapi.ModeHTML

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, rowData := range rowsData {
		var row []tgbotapi.InlineKeyboardButton
		for j, button := range rowData {
			if len(button) != 2 {
				log.Println(fmt.Errorf("invalid keyboard button (row %d, button %d)", i, j))
			}

			row = append(row, tgbotapi.NewInlineKeyboardButtonData(button[0], button[1]))
		}
		rows = append(rows, row)

	}

	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	_, err := c.api.Send(message)
	if err != nil {
		log.Println("cannot send telegram message (with inline keyboard):", err)
	}
}

func prepareCurrenciesKeyboard(currencies []string) [][][]string {
	var buttons [][]string
	for _, currency := range currencies {
		code, flag, _ := strings.Cut(currency, " ")
		buttons = append(buttons, []string{flag + " " + code, code})
	}

	var keyboard [][][]string
	for _buttonsPerRow < len(buttons) {
		buttons, keyboard = buttons[_buttonsPerRow:], append(keyboard, buttons[:_buttonsPerRow:_buttonsPerRow])
	}
	keyboard = append(keyboard, buttons)

	return keyboard
}

func (c *client) resolveUser(tgUser *tgbotapi.User) (*types.User, error) {
	if user, err := c.storage.FetchByID(tgUser.ID); err == nil {
		return user, nil
	}

	return c.storage.Add(tgUser.ID)
}