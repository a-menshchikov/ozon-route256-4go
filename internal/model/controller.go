package model

import (
	"context"
	"log"
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

type Controller interface {
	ListCurrencies(request.ListCurrencies) response.ListCurrencies
	SetCurrency(request.SetCurrency) response.SetCurrency

	ListLimits(request.ListLimits) response.ListLimits
	SetLimit(request.SetLimit) response.SetLimit

	AddExpense(request.AddExpense) response.AddExpense

	GetReport(request.GetReport) response.GetReport
}

type expenser interface {
	Add(user *types.User, date time.Time, amount int64, currency, category string) error
	Report(user *types.User, from time.Time) (map[string][]types.ExpenseItem, error)
}

type limiter interface {
	Get(user *types.User, category string) (types.LimitItem, error)
	Set(user *types.User, limit int64, currency, category string) error
	Decrease(user *types.User, value int64, category string) (bool, error)
	Unset(user *types.User, category string) error
	List(user *types.User) (map[string]types.LimitItem, error)
}

type rater interface {
	Run(ctx context.Context)
	Ready() bool
	Exchange(value int64, from, to string, date time.Time) (int64, error)
}

type currencyManager interface {
	Get(user *types.User) (string, error)
	Set(user *types.User, currency string) error
	ListCurrenciesCodesWithFlags() []string
}

type controller struct {
	expenser        expenser
	limiter         limiter
	currencyManager currencyManager
	rater           rater
}

func NewController(e expenser, l limiter, c currencyManager, r rater) *controller {
	return &controller{
		expenser:        e,
		limiter:         l,
		currencyManager: c,
		rater:           r,
	}
}

func (c *controller) ListCurrencies(req request.ListCurrencies) (resp response.ListCurrencies) {
	currency, ok := c.resolveUserCurrency(req.User)
	if ok {
		resp.Current = currency
		resp.List = c.currencyManager.ListCurrenciesCodesWithFlags()
	}

	return
}

func (c *controller) SetCurrency(req request.SetCurrency) (resp response.SetCurrency) {
	if err := c.currencyManager.Set(req.User, req.Code); err != nil {
		logHandleError("cannot set user currency:", err, req)
		return
	}

	resp = true
	return
}

func (c *controller) ListLimits(req request.ListLimits) (resp response.ListLimits) {
	resp.Ready = c.rater.Ready()
	if !resp.Ready {
		return
	}

	currency, ok := c.resolveUserCurrency(req.User)
	if !ok {
		return
	}

	resp.CurrentCurrency = currency

	limits, err := c.limiter.List(req.User)
	if err != nil {
		logHandleError("cannot get user limits:", err, req)
		return
	}

	list := make(map[string]response.LimitItem)
	today := utils.TruncateToDate(time.Now())

	for category := range limits {
		origin := limits[category]
		item := response.LimitItem{Origin: origin}

		if item.Total, err = c.rater.Exchange(origin.Total, origin.Currency, currency, today); err != nil {
			logHandleError("cannot exchange currency:", err, origin.Currency, currency)
			return
		}

		if item.Remains, err = c.rater.Exchange(origin.Remains, origin.Currency, currency, today); err != nil {
			logHandleError("cannot exchange currency:", err, origin.Currency, currency)
			return
		}

		list[category] = item
	}

	resp.List = list
	resp.Success = true

	return
}

func (c *controller) SetLimit(req request.SetLimit) response.SetLimit {
	currency, ok := c.resolveUserCurrency(req.User)
	if !ok {
		return false
	}

	if req.Value == 0 {
		err := c.limiter.Unset(req.User, req.Category)
		if err != nil {
			logHandleError("cannot unset user limit:", err, req)
			return false
		}
	} else {
		err := c.limiter.Set(req.User, req.Value, currency, req.Category)
		if err != nil {
			logHandleError("cannot set user limit:", err, req)
			return false
		}
	}

	return true
}

func (c *controller) AddExpense(req request.AddExpense) (resp response.AddExpense) {
	resp.Ready = c.rater.Ready()
	if !resp.Ready {
		return
	}

	currency, ok := c.resolveUserCurrency(req.User)
	if !ok {
		return
	}

	if err := c.expenser.Add(req.User, req.Date, req.Amount, currency, req.Category); err != nil {
		logHandleError("cannot add expense:", err, req)
		return
	}

	resp.Success = true

	limit, err := c.limiter.Get(req.User, req.Category)
	if err != nil {
		logHandleError("cannot get user limit:", err, req)
		return
	}

	if limit.Total == 0 {
		return
	}

	limitRetention, err := c.rater.Exchange(req.Amount, currency, limit.Currency, utils.TruncateToDate(req.Date))
	if err != nil {
		logHandleError("cannot exchange currency:", err, currency, limit.Currency)
		return
	}

	reached, err := c.limiter.Decrease(req.User, limitRetention, req.Category)
	if err != nil {
		logHandleError("cannot decrease limit:", err, currency, limit)
		return
	}

	resp.LimitReached = reached
	return
}

func (c *controller) GetReport(req request.GetReport) response.GetReport {
	resp := response.GetReport{
		From:  req.From,
		Ready: c.rater.Ready(),
	}

	if !resp.Ready {
		return resp
	}

	currency, ok := c.resolveUserCurrency(req.User)
	if !ok {
		return resp
	}

	resp.Currency = currency

	report, err := c.expenser.Report(req.User, req.From)
	if err != nil {
		logHandleError("cannot get expenses list:", err, req)
		return resp
	}

	data := make(map[string]int64)

	for category := range report {
		for _, item := range report[category] {
			amount, err := c.rater.Exchange(item.Amount, item.Currency, resp.Currency, item.Date)
			if err != nil {
				logHandleError("cannot exchange currency", err, item.Currency, resp.Currency)
				return resp
			}

			data[category] += amount
		}
	}

	resp.Data = data
	resp.Success = true

	return resp
}

func (c *controller) resolveUserCurrency(user *types.User) (string, bool) {
	currency, err := c.currencyManager.Get(user)
	if err != nil {
		logHandleError("cannot get user currency:", err)
		return "", false
	}

	return currency, true
}

func logHandleError(prefix string, err error, rest ...interface{}) {
	log.Println(prefix, err, rest)
}
