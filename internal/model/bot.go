package model

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto"
)

const _buttonsPerRow = 4

var (
	_addRx    = regexp.MustCompile(`^(|@|-\d+d|\d{2}\.\d{2}\.\d{4})\s*(\d+(?:[.,]\d+)?) (.+)$`)
	_reportRx = regexp.MustCompile(`^(\d*)([wmy]?)$`)

	errWrongExpenseDate    = errors.New("не удалось определить дату")
	errWrongExpenseAmount  = errors.New("не удалось определить сумму")
	errWrongReportDuration = errors.New("не удалось определить срок формирования отчёта")
)

type MessageSender interface {
	SendMessage(userID int64, text string) error
	SendMessageWithInlineKeyboard(userID int64, text string, rows [][][]string) error
}

type ExpenseStorage interface {
	Init(userID int64)
	Add(userID int64, date time.Time, amount int64, category string) error
	List(userID int64, from time.Time) map[string]int64
}

type Exchanger interface {
	Ready() bool
	ExchangeFromBase(value int64, currency string) (int64, error)
	ExchangeToBase(value int64, currency string) (int64, error)
	ListCurrencies() []string
}

type CurrencyKeeper interface {
	Set(userID int64, currency string) error
	Get(userID int64) string
}

type Bot struct {
	sender         MessageSender
	storage        ExpenseStorage
	exchanger      Exchanger
	currencyKeeper CurrencyKeeper
}

func NewBot(sender MessageSender, storage ExpenseStorage, exchanger Exchanger, currencyKeeper CurrencyKeeper) *Bot {
	return &Bot{
		sender:         sender,
		storage:        storage,
		exchanger:      exchanger,
		currencyKeeper: currencyKeeper,
	}
}

func (b *Bot) HandleMessage(msg dto.Message) error {
	command, args, _ := strings.Cut(msg.Text, " ")
	args = strings.TrimSpace(args)

	var response string

	switch command {
	case "/start":
		response = b.start(msg.UserID)

	case "/add":
		response = b.addExpense(args, msg.UserID)

	case "/report":
		response = b.report(args, msg.UserID)

	case "/currency":
		if args == "" {
			return b.sender.SendMessageWithInlineKeyboard(
				msg.UserID,
				currencyCurrentMessage+b.currencyKeeper.Get(msg.UserID)+"\n\n"+currencyChooseMessage,
				prepareCurrenciesKeyboard(b.exchanger.ListCurrencies()),
			)
		}

		response = b.changeCurrency(msg.UserID, args)

	default:
		response = sorryMessage
	}

	return b.sender.SendMessage(msg.UserID, response)
}

func (b *Bot) HandleCallbackQuery(query dto.CallbackQuery) error {
	response := b.changeCurrency(query.UserID, query.Data)

	return b.sender.SendMessage(query.UserID, response)
}

func (b *Bot) start(userID int64) string {
	b.storage.Init(userID)

	return helloMessage
}

func (b *Bot) addExpense(args string, userID int64) string {
	m := _addRx.FindStringSubmatch(args)
	if len(m) == 0 {
		return "Не удалось определить расход.\n\n" + addHelpMessage
	}

	date, amount, category, err := parseAddArgs(m[1:])

	if err == nil {
		currency := b.currencyKeeper.Get(userID)
		amount, err = b.exchanger.ExchangeToBase(amount, currency)
	}

	if err == nil {
		err = b.storage.Add(userID, date, amount, category)
	}

	if err != nil {
		return errorMessage(err, "Не удалось добавить расход.", addHelpMessage)
	}

	return doneMessage
}

func parseAddArgs(args []string) (date time.Time, amount int64, category string, err error) {
	if date, err = parseDate(args[0]); err != nil {
		return
	}

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

func (b *Bot) report(args string, userID int64) string {
	m := _reportRx.FindStringSubmatch(args)
	if len(m) == 0 {
		return reportHelpMessage
	}

	from, err := parseReportArgs(m[1:])
	if err != nil {
		return errorMessage(err, "Не удалось сформировать отчёт.", reportHelpMessage)
	}

	data := b.storage.List(userID, from)
	if len(data) == 0 {
		return "Вы ещё не добавили ни одного расхода."
	}

	categories := make([]string, 0, len(data))
	for category := range data {
		categories = append(categories, category)
	}

	sort.Strings(categories)

	currency := b.currencyKeeper.Get(userID)
	response := fmt.Sprintf("Расходы с %s (валюта — %s):\n", from.Format("02.01.2006"), currency)
	for _, category := range categories {
		amount, err := b.exchanger.ExchangeFromBase(data[category], currency)
		if err != nil {
			return "Ошибка при формировании отчёта: " + err.Error()
		}

		response += fmt.Sprintf("%s: %.2f\n", category, float64(amount)/10000)
	}

	return response
}

func parseReportArgs(args []string) (time.Time, error) {
	hours := 1

	if args[0] != "" {
		rate, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return time.Time{}, errWrongReportDuration
		}
		hours *= int(rate)
	}

	switch args[1] {
	case "":
		hours = 24 * 7
	case "w":
		hours *= 24 * 7
	case "m":
		hours *= 24 * 30
	case "y":
		hours *= 24 * 365
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%dh", hours))
	if err != nil {
		return time.Time{}, errWrongReportDuration
	}

	return time.Now().Truncate(24 * time.Hour).Add(-duration), nil
}

func (b *Bot) changeCurrency(userID int64, currency string) string {
	if !b.exchanger.Ready() {
		return currencyLaterMessage
	}

	if err := b.currencyKeeper.Set(userID, currency); err != nil {
		return errorMessage(err, "Не удалось сменить текущую валюту.", currencyHelpMessage)
	}

	return doneMessage
}

func prepareCurrenciesKeyboard(currencies []string) [][][]string {
	var buttons [][]string
	for _, currency := range currencies {
		name, flag, ok := strings.Cut(currency, " ")
		if ok {
			buttons = append(buttons, []string{flag + " " + name, name})
		}
	}

	var keyboard [][][]string
	for _buttonsPerRow < len(buttons) {
		buttons, keyboard = buttons[_buttonsPerRow:], append(keyboard, buttons[:_buttonsPerRow:_buttonsPerRow])
	}
	keyboard = append(keyboard, buttons)

	return keyboard
}
