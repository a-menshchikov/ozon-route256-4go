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

var (
	addRx    = regexp.MustCompile(`^(|@|-\d+d|\d{2}\.\d{2}\.\d{4})\s*(\d+(?:[.,]\d+)?) (.+)$`)
	reportRx = regexp.MustCompile(`^(\d*)([wmy]?)$`)

	errWrongExpenseDate    = errors.New("не удалось определить дату")
	errWrongExpenseAmount  = errors.New("не удалось определить сумму")
	errWrongReportDuration = errors.New("не удалось определить срок формирования отчёта")
)

type MessageSender interface {
	SendMessage(userID int64, text string) error
}

type ExpenseStorage interface {
	Init(userID int64)
	Add(userID int64, date time.Time, amount int64, category string) error
	List(userID int64, from time.Time) map[string]int64
}

type Bot struct {
	sender  MessageSender
	storage ExpenseStorage
}

func NewBot(sender MessageSender, storage ExpenseStorage) *Bot {
	return &Bot{
		sender:  sender,
		storage: storage,
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

	default:
		response = sorryMessage
	}

	return b.sender.SendMessage(msg.UserID, response)
}

func (b *Bot) start(userID int64) string {
	b.storage.Init(userID)

	return helloMessage
}

func (b *Bot) addExpense(args string, userID int64) string {
	m := addRx.FindStringSubmatch(args)
	if len(m) == 0 {
		return "Не удалось определить расход.\n\n" + addHelpMessage
	}

	date, amount, category, err := parseAddArgs(m[1:])
	if err == nil {
		err = b.storage.Add(userID, date, amount, category)
	}

	if err != nil {
		return errorMessage(err, "Не удалось добавить расход.", addHelpMessage)
	}

	return "Готово!"
}

func parseAddArgs(args []string) (date time.Time, amount int64, category string, err error) {
	if date, err = parseDate(args[0]); err != nil {
		return
	}

	floatAmount, err := strconv.ParseFloat(strings.ReplaceAll(args[1], ",", "."), 64)
	if err != nil {
		err = errWrongExpenseAmount
	}

	amount = int64(floatAmount * 100)

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
	m := reportRx.FindStringSubmatch(args)
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

	response := fmt.Sprintf("Расходы с %s:\n", from.Format("02.01.2006"))
	for _, category := range categories {
		response += fmt.Sprintf("%s: %.2f\n", category, float64(data[category])/100)
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
