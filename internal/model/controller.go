package model

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
)

type Controller interface {
	ListCurrencies(ctx context.Context, req request.ListCurrencies) response.ListCurrencies
	SetCurrency(ctx context.Context, req request.SetCurrency) response.SetCurrency

	ListLimits(ctx context.Context, req request.ListLimits) response.ListLimits
	SetLimit(ctx context.Context, req request.SetLimit) response.SetLimit

	AddExpense(ctx context.Context, req request.AddExpense) response.AddExpense

	GetReport(ctx context.Context, req request.GetReport) response.GetReport
}

type expenser interface {
	Add(ctx context.Context, user *types.User, date time.Time, amount int64, currency, category string) error
	Report(ctx context.Context, user *types.User, from time.Time) (map[string][]types.ExpenseItem, error)
}

type limiter interface {
	Get(ctx context.Context, user *types.User, category string) (types.LimitItem, error)
	Set(ctx context.Context, user *types.User, limit int64, currency, category string) error
	Decrease(ctx context.Context, user *types.User, value int64, category string) (bool, error)
	Unset(ctx context.Context, user *types.User, category string) error
	List(ctx context.Context, user *types.User) (map[string]types.LimitItem, error)
}

type rater interface {
	Run(ctx context.Context) error
	Ready() bool
	Exchange(ctx context.Context, value int64, from, to string, date time.Time) (int64, error)
}

type currencyManager interface {
	Get(ctx context.Context, user *types.User) (string, error)
	Set(ctx context.Context, user *types.User, currency string) error
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

func (c *controller) ListCurrencies(ctx context.Context, req request.ListCurrencies) (resp response.ListCurrencies) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "controller.ListCurrencies")
	defer span.Finish()

	currency, ok := c.resolveUserCurrency(ctx, req.User)
	if ok {
		resp.Current = currency
		resp.List = c.currencyManager.ListCurrenciesCodesWithFlags()
	}

	return
}

func (c *controller) SetCurrency(ctx context.Context, req request.SetCurrency) (resp response.SetCurrency) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "controller.SetCurrency")
	defer span.Finish()

	if err := c.currencyManager.Set(ctx, req.User, req.Code); err != nil {
		c.logger.Error("cannot set user currency", zap.Error(err), zap.Object("request", req))
		return
	}

	resp = true
	return
}

func (c *controller) ListLimits(ctx context.Context, req request.ListLimits) (resp response.ListLimits) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "controller.ListLimits")
	defer span.Finish()

	resp.Ready = c.rater.Ready()
	if !resp.Ready {
		return
	}

	currency, ok := c.resolveUserCurrency(ctx, req.User)
	if !ok {
		return
	}

	resp.CurrentCurrency = currency

	limits, err := c.limiter.List(ctx, req.User)
	if err != nil {
		c.logger.Error("cannot get user limits", zap.Error(err), zap.Object("request", req))
		return
	}

	list := make(map[string]response.LimitItem)
	today := utils.TruncateToDate(time.Now())

	for category := range limits {
		origin := limits[category]
		item := response.LimitItem{Origin: origin}

		if item.Total, err = c.rater.Exchange(ctx, origin.Total, origin.Currency, currency, today); err != nil {
			c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", origin.Currency), zap.String("to", currency))
			return
		}

		if item.Remains, err = c.rater.Exchange(ctx, origin.Remains, origin.Currency, currency, today); err != nil {
			c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", origin.Currency), zap.String("to", currency))
			return
		}

		list[category] = item
	}

	resp.List = list
	resp.Success = true

	return
}

func (c *controller) SetLimit(ctx context.Context, req request.SetLimit) response.SetLimit {
	span, ctx := opentracing.StartSpanFromContext(ctx, "controller.SetLimit")
	defer span.Finish()

	currency, ok := c.resolveUserCurrency(ctx, req.User)
	if !ok {
		return false
	}

	if req.Value == 0 {
		err := c.limiter.Unset(ctx, req.User, req.Category)
		if err != nil {
			c.logger.Error("cannot unset user limit", zap.Error(err), zap.Object("request", req))
			return false
		}
	} else {
		err := c.limiter.Set(ctx, req.User, req.Value, currency, req.Category)
		if err != nil {
			c.logger.Error("cannot set user limit", zap.Error(err), zap.Object("request", req))
			return false
		}
	}

	return true
}

func (c *controller) AddExpense(ctx context.Context, req request.AddExpense) (resp response.AddExpense) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "controller.AddExpense")
	defer span.Finish()

	resp.Ready = c.rater.Ready()
	if !resp.Ready {
		return
	}

	currency, ok := c.resolveUserCurrency(ctx, req.User)
	if !ok {
		return
	}

	if err := c.expenser.Add(ctx, req.User, req.Date, req.Amount, currency, req.Category); err != nil {
		c.logger.Error("cannot add expense", zap.Error(err), zap.Object("request", req))
		return
	}

	resp.Success = true

	limit, err := c.limiter.Get(ctx, req.User, req.Category)
	if err != nil {
		c.logger.Error("cannot get user limit", zap.Error(err), zap.Object("request", req))
		return
	}

	if limit.Total == 0 {
		return
	}

	limitRetention, err := c.rater.Exchange(ctx, req.Amount, currency, limit.Currency, utils.TruncateToDate(req.Date))
	if err != nil {
		c.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", currency), zap.String("to", limit.Currency))
		return
	}

	reached, err := c.limiter.Decrease(ctx, req.User, limitRetention, req.Category)
	if err != nil {
		c.logger.Error("cannot decrease limit", zap.Error(err), zap.String("currency", currency), zap.Object("limit", limit))
		return
	}

	resp.LimitReached = reached
	return
}

func (c *controller) GetReport(ctx context.Context, req request.GetReport) response.GetReport {
	span, ctx := opentracing.StartSpanFromContext(ctx, "controller.GetReport")
	defer span.Finish()

	resp := response.GetReport{
		From:  req.From,
		Ready: c.rater.Ready(),
	}

	if !resp.Ready {
		return resp
	}

	currency, ok := c.resolveUserCurrency(ctx, req.User)
	if !ok {
		return resp
	}

	resp.Currency = currency

	report, err := c.expenser.Report(ctx, req.User, req.From)
	if err != nil {
		c.logger.Error("cannot get expenses list", zap.Error(err), zap.Object("request", req))
		return resp
	}

	data := make(map[string]int64)

	for category := range report {
		for _, item := range report[category] {
			amount, err := c.rater.Exchange(ctx, item.Amount, item.Currency, resp.Currency, item.Date)
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

func (c *controller) resolveUserCurrency(ctx context.Context, user *types.User) (string, bool) {
	currency, err := c.currencyManager.Get(ctx, user)
	if err != nil {
		c.logger.Error("cannot get user currency", zap.Error(err), zap.Int64("user", int64(*user)))
		return "", false
	}

	return currency, true
}
