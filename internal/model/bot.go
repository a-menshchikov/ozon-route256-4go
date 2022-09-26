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

const (
	addHelp = `Чтобы добавить запись о расходах, отправь команду:
` + "```" + `
/add [дата] <сумма> <категория>
` + "```" + `
Дата может быть указана в формате *dd\.mm\.yyyy* \(день\.месяц\.год\)\.
Чтобы задать сегодняшнее число, можно использовать знак *@* в качестве даты, или не указывать дату совсем\.
Кроме того, в качестве даты можно использовать строку вида *\-Nd*, где N — количество "дней назад" \(1 можно не указывать\)\.
Например, *\-2d* значит "2 дня назад"\.

Сумма указывается в формате *XX\[\.yy\]*: целого или дробного числа с одним или двумя знаками после запятой \(вместо которой можно использовать точку\)\.`
	reportHelp = `Для просмотра расходов по категориям выполните одну из команд \(w — расходы за неделю, m — за месяц, y — за год\):
` + "```" + `
/report \[N\]w
/report \[N\]m
/report \[N\]y
` + "```" + `
Если задать положительное число N, будут выведены расходы за N последних недель/месяцев/лет\.`
)

var (
	commandRx = regexp.MustCompile(`^/(\w+)\b(.*)$`)
	addRx     = regexp.MustCompile(`^(|@|-\d+d|\d{2}\.\d{2}\.\d{4})\s*(\d+(?:[.,]\d+)?) (.+)$`)
	reportRx  = regexp.MustCompile(`^(\d*)([wmy])$`)
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
	m := commandRx.FindStringSubmatch(msg.Text)

	command, args := m[1], m[2]
	args = strings.TrimSpace(args)

	var response string

	switch command {
	case "start":
		response = b.start(msg.UserID)

	case "add":
		response = b.addExpense(args, msg.UserID)

	case "report":
		response = b.report(args, msg.UserID)

	default:
		response = "Извини, я не знаю такой команды\\. 🙁\n\n" + addHelp + "\n\n\n" + reportHelp
	}

	return b.sender.SendMessage(msg.UserID, response)
}

func (b *Bot) start(userID int64) string {
	b.storage.Init(userID)

	return "Привет\\! 👋\n\n" + addHelp + "\n\n\n" + reportHelp
}

func (b *Bot) addExpense(args string, userID int64) string {
	m := addRx.FindStringSubmatch(args)
	if len(m) == 0 {
		return addHelp
	}

	date, amount, category, err := parseAddArgs(m[1:])
	if err == nil {
		err = b.storage.Add(userID, date, amount, category)
	}

	if err != nil {
		return "Не удалось добавить расход\\.\nОшибка: " + err.Error() + "\\.\n\n\nДля справки:\n" + addHelp
	}

	return "Готово\\!"
}

func parseAddArgs(args []string) (date time.Time, amount int64, category string, err error) {
	if date, err = parseDate(args[0]); err != nil {
		return
	}

	floatAmount, err := strconv.ParseFloat(strings.ReplaceAll(args[1], ",", "."), 64)
	if err != nil {
		err = errors.New("не удалось определить сумму")
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
			return time.Time{}, errors.New("не удалось определить дату")
		}

		return time.Now().Add(-time.Duration(rate) * 24 * time.Hour), nil
	}

	date, err := time.Parse("02.01.2006", input)
	if err != nil {
		return time.Time{}, errors.New("не удалось определить дату")
	}

	return date, nil
}

func (b *Bot) report(args string, userID int64) string {
	m := reportRx.FindStringSubmatch(args)
	if len(m) == 0 {
		return reportHelp
	}

	from, err := parseReportArgs(m[1:])
	if err != nil {
		return "Не удалось сформировать отчёт\\.\nОшибка: " + err.Error() + "\\.\n\n\nДля справки:\n" + reportHelp
	}

	data := b.storage.List(userID, from)
	if len(data) == 0 {
		return "Вы ещё не добавили ни одного расхода\\."
	}

	categories := make([]string, 0, len(data))
	for category := range data {
		categories = append(categories, category)
	}

	sort.Strings(categories)

	response := fmt.Sprintf("Ваши расходы с %s:\n", from.Format("02\\.01\\.2006"))

	for _, category := range categories {
		response += strings.ReplaceAll(fmt.Sprintf("%s: %.2f\n", category, float64(data[category])/100), ".", "\\.")
	}

	return response
}

func parseReportArgs(args []string) (time.Time, error) {
	hours := 1

	if args[0] != "" {
		rate, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return time.Time{}, errors.New("не удалось определить срок формирования отчёта")
		}
		hours *= int(rate)
	}

	switch args[1] {
	case "w":
		hours *= 24 * 7
	case "m":
		hours *= 24 * 30
	case "y":
		hours *= 24 * 365
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%dh", hours))
	if err != nil {
		return time.Time{}, errors.New("не удалось определить срок формирования отчёта")
	}

	return time.Now().Truncate(24 * time.Hour).Add(-duration), nil
}
