package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
	today       = utils.TruncateToDate(time.Now())
	yesterday   = today.Add(-time.Hour * 24)
	simpleError = errors.New("error")
)

func Test_controller_ListCurrencies(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		currencyManager func() currencyManager
	}
	type args struct {
		req request.ListCurrencies
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp response.ListCurrencies
	}{
		{
			name: "no currency",
			fields: fields{
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("", simpleError)
					return m
				},
			},
			args: args{
				req: request.ListCurrencies{
					User: testUser,
				},
			},
			wantResp: response.ListCurrencies{},
		},
		{
			name: "success",
			fields: fields{
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					m.EXPECT().ListCurrenciesCodesWithFlags().Return([]string{"RUB", "USD"})
					return m
				},
			},
			args: args{
				req: request.ListCurrencies{
					User: testUser,
				},
			},
			wantResp: response.ListCurrencies{
				Current: "RUB",
				List:    []string{"RUB", "USD"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewController(nil, nil, tt.fields.currencyManager(), nil)
			if gotResp := a.ListCurrencies(tt.args.req); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("ListCurrencies() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func Test_controller_SetCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		currencyManager func() currencyManager
	}
	type args struct {
		req request.SetCurrency
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp response.SetCurrency
	}{
		{
			name: "failed",
			fields: fields{
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Set(testUser, "RUB").Return(simpleError)
					return m
				},
			},
			args: args{
				req: request.SetCurrency{
					User: testUser,
					Code: "RUB",
				},
			},
			wantResp: false,
		},
		{
			name: "success",
			fields: fields{
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Set(testUser, "EUR").Return(nil)
					return m
				},
			},
			args: args{
				req: request.SetCurrency{
					User: testUser,
					Code: "EUR",
				},
			},
			wantResp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewController(nil, nil, tt.fields.currencyManager(), nil)
			if gotResp := a.SetCurrency(tt.args.req); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("SetCurrency() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func Test_controller_ListLimits(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		limiter         func() limiter
		currencyManager func() currencyManager
		rater           func() rater
	}
	type args struct {
		req request.ListLimits
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp response.ListLimits
	}{
		{
			name: "not ready",
			fields: fields{
				limiter: func() limiter {
					return nil
				},
				currencyManager: func() currencyManager {
					return nil
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(false)
					return m
				},
			},
			args: args{
				req: request.ListLimits{
					User: testUser,
				},
			},
			wantResp: response.ListLimits{},
		},
		{
			name: "no currency",
			fields: fields{
				limiter: func() limiter {
					return nil
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("", simpleError)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.ListLimits{
					User: testUser,
				},
			},
			wantResp: response.ListLimits{
				Ready: true,
			},
		},
		{
			name: "no limits list",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().List(testUser).Return(nil, simpleError)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.ListLimits{
					User: testUser,
				},
			},
			wantResp: response.ListLimits{
				Ready:           true,
				CurrentCurrency: "RUB",
			},
		},
		{
			name: "cannot exchange total",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
						"": {
							Total:    1000000,
							Remains:  750000,
							Currency: "USD",
						},
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(1000000), "USD", "RUB", today).Return(int64(0), simpleError)
					return m
				},
			},
			args: args{
				req: request.ListLimits{
					User: testUser,
				},
			},
			wantResp: response.ListLimits{
				Ready:           true,
				CurrentCurrency: "RUB",
			},
		},
		{
			name: "cannot exchange remains",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
						"": {
							Total:    2000000,
							Remains:  1500000,
							Currency: "USD",
						},
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(2000000), "USD", "RUB", today).Return(int64(10000000), nil)
					m.EXPECT().Exchange(int64(1500000), "USD", "RUB", today).Return(int64(0), simpleError)
					return m
				},
			},
			args: args{
				req: request.ListLimits{
					User: testUser,
				},
			},
			wantResp: response.ListLimits{
				Ready:           true,
				CurrentCurrency: "RUB",
			},
		},
		{
			name: "success",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
						"taxi": {
							Total:    400000,
							Remains:  300000,
							Currency: "USD",
						},
						"": {
							Total:    30000000,
							Remains:  10000000,
							Currency: "RUB",
						},
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(400000), "USD", "RUB", today).Return(int64(2000000), nil)
					m.EXPECT().Exchange(int64(300000), "USD", "RUB", today).Return(int64(1500000), nil)
					m.EXPECT().Exchange(int64(30000000), "RUB", "RUB", today).Return(int64(30000000), nil)
					m.EXPECT().Exchange(int64(10000000), "RUB", "RUB", today).Return(int64(10000000), nil)
					return m
				},
			},
			args: args{
				req: request.ListLimits{
					User: testUser,
				},
			},
			wantResp: response.ListLimits{
				Ready:           true,
				CurrentCurrency: "RUB",
				List: map[string]response.LimitItem{
					"taxi": {
						Total:   2000000,
						Remains: 1500000,
						Origin: types.LimitItem{
							Total:    400000,
							Remains:  300000,
							Currency: "USD",
						},
					},
					"": {
						Total:   30000000,
						Remains: 10000000,
						Origin: types.LimitItem{
							Total:    30000000,
							Remains:  10000000,
							Currency: "RUB",
						},
					},
				},
				Success: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewController(nil, tt.fields.limiter(), tt.fields.currencyManager(), tt.fields.rater())
			if gotResp := a.ListLimits(tt.args.req); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("ListLimits() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func Test_controller_SetLimit(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		limiter         func() limiter
		currencyManager func() currencyManager
	}
	type args struct {
		req request.SetLimit
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp response.SetLimit
	}{
		{
			name: "no currency",
			fields: fields{
				limiter: func() limiter {
					return nil
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("", simpleError)
					return m
				},
			},
			args: args{
				req: request.SetLimit{
					User:     testUser,
					Value:    1000000,
					Category: "",
				},
			},
			wantResp: false,
		},
		{
			name: "cannot set",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Set(testUser, int64(1000000), "USD", "taxi").Return(simpleError)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
			},
			args: args{
				req: request.SetLimit{
					User:     testUser,
					Value:    1000000,
					Category: "taxi",
				},
			},
			wantResp: false,
		},
		{
			name: "successful set",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Set(testUser, int64(2000000), "USD", "taxi").Return(nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
			},
			args: args{
				req: request.SetLimit{
					User:     testUser,
					Value:    2000000,
					Category: "taxi",
				},
			},
			wantResp: true,
		},
		{
			name: "cannot unset",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Unset(testUser, "taxi").Return(simpleError)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
			},
			args: args{
				req: request.SetLimit{
					User:     testUser,
					Value:    0,
					Category: "taxi",
				},
			},
			wantResp: false,
		},
		{
			name: "successful unset",
			fields: fields{
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Unset(testUser, "taxi").Return(nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
			},
			args: args{
				req: request.SetLimit{
					User:     testUser,
					Value:    0,
					Category: "taxi",
				},
			},
			wantResp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewController(nil, tt.fields.limiter(), tt.fields.currencyManager(), nil)
			if gotResp := a.SetLimit(tt.args.req); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("SetLimit() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func Test_controller_AddExpense(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		expenser        func() expenser
		limiter         func() limiter
		currencyManager func() currencyManager
		rater           func() rater
	}
	type args struct {
		req request.AddExpense
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp response.AddExpense
	}{
		{
			name: "not ready",
			fields: fields{
				expenser: func() expenser {
					return nil
				},
				limiter: func() limiter {
					return nil
				},
				currencyManager: func() currencyManager {
					return nil
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(false)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   10000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{},
		},
		{
			name: "no currency",
			fields: fields{
				expenser: func() expenser {
					return nil
				},
				limiter: func() limiter {
					return nil
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("", simpleError)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   15000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready: true,
			},
		},
		{
			name: "cannot add",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(20000), "USD", "coffee").Return(simpleError)
					return m
				},
				limiter: func() limiter {
					return nil
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   20000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready: true,
			},
		},
		{
			name: "cannot get limits",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(25000), "USD", "coffee").Return(nil)
					return m
				},
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{}, simpleError)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   25000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready:   true,
				Success: true,
			},
		},
		{
			name: "no limit",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(30000), "USD", "coffee").Return(nil)
					return m
				},
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
						Total: 0,
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   30000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready:   true,
				Success: true,
			},
		},
		{
			name: "cannot exchange for limit",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(35000), "EUR", "coffee").Return(nil)
					return m
				},
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
						Total:    60000,
						Remains:  30000,
						Currency: "USD",
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("EUR", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(35000), "EUR", "USD", today).Return(int64(0), simpleError)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   35000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready:   true,
				Success: true,
			},
		},
		{
			name: "cannot decrease limit",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(2000000), "RUB", "coffee").Return(nil)
					return m
				},
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
						Total:    60000,
						Remains:  30000,
						Currency: "USD",
					}, nil)
					m.EXPECT().Decrease(testUser, int64(40000), "coffee").Return(false, simpleError)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(2000000), "RUB", "USD", today).Return(int64(40000), nil)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   2000000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready:   true,
				Success: true,
			},
		},
		{
			name: "limit has been reached",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(2000000), "RUB", "coffee").Return(nil)
					return m
				},
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
						Total:    60000,
						Remains:  30000,
						Currency: "USD",
					}, nil)
					m.EXPECT().Decrease(testUser, int64(40000), "coffee").Return(true, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(2000000), "RUB", "USD", today).Return(int64(40000), nil)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   2000000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready:        true,
				LimitReached: true,
				Success:      true,
			},
		},
		{
			name: "limit has not been reached",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Add(testUser, today, int64(1000000), "RUB", "coffee").Return(nil)
					return m
				},
				limiter: func() limiter {
					m := mocks.NewMocklimiter(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
						Total:    60000,
						Remains:  30000,
						Currency: "USD",
					}, nil)
					m.EXPECT().Decrease(testUser, int64(20000), "coffee").Return(true, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(1000000), "RUB", "USD", today).Return(int64(20000), nil)
					return m
				},
			},
			args: args{
				req: request.AddExpense{
					User:     testUser,
					Date:     today,
					Amount:   1000000,
					Category: "coffee",
				},
			},
			wantResp: response.AddExpense{
				Ready:        true,
				LimitReached: true,
				Success:      true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewController(tt.fields.expenser(), tt.fields.limiter(), tt.fields.currencyManager(), tt.fields.rater())
			if gotResp := a.AddExpense(tt.args.req); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("SetLimit() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func Test_controller_GetReport(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		expenser        func() expenser
		currencyManager func() currencyManager
		rater           func() rater
	}
	type args struct {
		req request.GetReport
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp response.GetReport
	}{
		{
			name: "not ready",
			fields: fields{
				expenser: func() expenser {
					return nil
				},
				currencyManager: func() currencyManager {
					return nil
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(false)
					return m
				},
			},
			args: args{
				req: request.GetReport{
					User: testUser,
					From: today,
				},
			},
			wantResp: response.GetReport{
				From:  today,
				Ready: false,
			},
		},
		{
			name: "no currency",
			fields: fields{
				expenser: func() expenser {
					return nil
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("", simpleError)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.GetReport{
					User: testUser,
					From: today,
				},
			},
			wantResp: response.GetReport{
				From:  today,
				Ready: true,
			},
		},
		{
			name: "no data",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Report(testUser, today).Return(nil, simpleError)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					return m
				},
			},
			args: args{
				req: request.GetReport{
					User: testUser,
					From: today,
				},
			},
			wantResp: response.GetReport{
				From:     today,
				Ready:    true,
				Currency: "RUB",
			},
		},
		{
			name: "cannot exchange currency",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Report(testUser, today).Return(map[string][]types.ExpenseItem{
						"taxi": {
							{
								Date:     today,
								Amount:   100000,
								Currency: "EUR",
							},
						},
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("RUB", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(100000), "EUR", "RUB", today).Return(int64(0), simpleError)
					return m
				},
			},
			args: args{
				req: request.GetReport{
					User: testUser,
					From: today,
				},
			},
			wantResp: response.GetReport{
				From:     today,
				Ready:    true,
				Currency: "RUB",
			},
		},
		{
			name: "success",
			fields: fields{
				expenser: func() expenser {
					m := mocks.NewMockexpenser(ctrl)
					m.EXPECT().Report(testUser, today).Return(map[string][]types.ExpenseItem{
						"coffee": {
							{
								Date:     today,
								Amount:   1000000,
								Currency: "RUB",
							},
						},
						"taxi": {
							{
								Date:     today,
								Amount:   100000,
								Currency: "USD",
							},
							{
								Date:     yesterday,
								Amount:   1500000,
								Currency: "RUB",
							},
						},
					}, nil)
					return m
				},
				currencyManager: func() currencyManager {
					m := mocks.NewMockcurrencyManager(ctrl)
					m.EXPECT().Get(testUser).Return("USD", nil)
					return m
				},
				rater: func() rater {
					m := mocks.NewMockrater(ctrl)
					m.EXPECT().Ready().Return(true)
					m.EXPECT().Exchange(int64(1000000), "RUB", "USD", today).Return(int64(20000), nil)
					m.EXPECT().Exchange(int64(100000), "USD", "USD", today).Return(int64(100000), nil)
					m.EXPECT().Exchange(int64(1500000), "RUB", "USD", yesterday).Return(int64(30000), nil)
					return m
				},
			},
			args: args{
				req: request.GetReport{
					User: testUser,
					From: today,
				},
			},
			wantResp: response.GetReport{
				From:     today,
				Ready:    true,
				Currency: "USD",
				Data: map[string]int64{
					"coffee": 20000,
					"taxi":   130000,
				},
				Success: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewController(tt.fields.expenser(), nil, tt.fields.currencyManager(), tt.fields.rater())
			if gotResp := a.GetReport(tt.args.req); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("GetReport() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
