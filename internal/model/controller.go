package model

import (
	"context"
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
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
	Run(ctx context.Context) error
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
	logger          *zap.Logger
}

func NewController(e expenser, l limiter, c currencyManager, r rater, logger *zap.Logger) *controller {
	return &controller{
		expenser:        e,
		limiter:         l,
		currencyManager: c,
		rater:           r,
		logger:          logger,
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
		c.logger.Error("cannot set user currency", zap.Error(err), zap.Object("request", req))
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
		c.logger.Error("cannot get user limits", zap.Error(err), zap.Object("request", req))
		return
	}

	list := make(map[string]response.LimitItem)
	today := utils.TruncateToDate(time.Now())

	for category := range limits {
		origin := limits[category]
		item := response.LimitItem{Origin: origin}

		if item.Total, err = c.rater.Exchange(origin.Total, origin.Currency, currency, today); err != nil {
			c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", origin.Currency), zap.String("to", currency))
			return
		}

		if item.Remains, err = c.rater.Exchange(origin.Remains, origin.Currency, currency, today); err != nil {
			c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", origin.Currency), zap.String("to", currency))
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
			c.logger.Error("cannot unset user limit", zap.Error(err), zap.Object("request", req))
			return false
		}
	} else {
		err := c.limiter.Set(req.User, req.Value, currency, req.Category)
		if err != nil {
			c.logger.Error("cannot set user limit", zap.Error(err), zap.Object("request", req))
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
		c.logger.Error("cannot add expense", zap.Error(err), zap.Object("request", req))
		return
	}

	resp.Success = true

	limit, err := c.limiter.Get(req.User, req.Category)
	if err != nil {
		c.logger.Error("cannot get user limit", zap.Error(err), zap.Object("request", req))
		return
	}

	if limit.Total == 0 {
		return
	}

	limitRetention, err := c.rater.Exchange(req.Amount, currency, limit.Currency, utils.TruncateToDate(req.Date))
	if err != nil {
		c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", currency), zap.String("to", limit.Currency))
		return
	}

	reached, err := c.limiter.Decrease(req.User, limitRetention, req.Category)
	if err != nil {
		c.logger.Error("cannot decrease limit", zap.Error(err), zap.String("currency", currency), zap.Object("limit", limit))
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
		c.logger.Error("cannot get expenses list", zap.Error(err), zap.Object("request", req))
		return resp
	}

	data := make(map[string]int64)

	for category := range report {
		for _, item := range report[category] {
			amount, err := c.rater.Exchange(item.Amount, item.Currency, resp.Currency, item.Date)
			if err != nil {
				c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", item.Currency), zap.String("to", resp.Currency))
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
		c.logger.Error("cannot get user currency", zap.Error(err), zap.Int64("user", int64(*user)))
		return "", false
	}

	return currency, true
}
