package rates

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	rmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/currency/rates"
	smocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

var (
	today       = utils.TruncateToDate(time.Now())
	yesterday   = today.Add(-24 * time.Hour)
	simpleError = errors.New("error")
	defaultCfg  = config.CurrencyConfig{
		Base: "USD",
		Available: []config.Currency{
			{
				Code: "USD",
				Flag: "$",
			},
			{
				Code: "EUR",
				Flag: "¢",
			},
		},
		RefreshInterval: 100 * time.Millisecond,
	}
)

func Test_rater_Run(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		cfg     config.CurrencyConfig
		storage func() storage.CurrencyRatesStorage
		gateway func() gateway
	}
	type args struct {
		ctx func() context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "refresh once",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Add("USD", today, int64(500000)).Return(nil)
					m.EXPECT().Add("EUR", today, int64(550000)).Return(nil)
					return m
				},
				gateway: func() gateway {
					m := rmocks.NewMockgateway(ctrl)
					var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
					m.EXPECT().FetchRates(gomock.AssignableToTypeOf(ctx)).Return(map[string]int64{
						"USD": 500000,
						"EUR": 550000,
					}, today, nil)
					return m
				},
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						<-time.After(time.Millisecond)
						cancel()
					}()
					return ctx
				},
			},
		},
		{
			name: "refresh twice with errors",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Add("USD", yesterday, int64(410000)).Return(simpleError)
					m.EXPECT().Add("EUR", yesterday, int64(460000)).Return(nil)
					return m
				},
				gateway: func() gateway {
					m := rmocks.NewMockgateway(ctrl)
					var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
					m.EXPECT().FetchRates(gomock.AssignableToTypeOf(ctx)).Return(nil, time.Time{}, simpleError)
					m.EXPECT().FetchRates(gomock.AssignableToTypeOf(ctx)).Return(map[string]int64{
						"USD": 410000,
						"EUR": 460000,
					}, yesterday, nil)
					return m
				},
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						<-time.After(150 * time.Millisecond)
						cancel()
					}()
					return ctx
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRater(tt.fields.cfg, tt.fields.storage(), tt.fields.gateway())
			r.Run(tt.args.ctx())
			if ready := r.Ready(); !ready {
				t.Error("Rater not ready after Run")
			}
		})
	}
}

func Test_rater_Ready(t *testing.T) {
	type fields struct {
		cfg     config.CurrencyConfig
		storage func() storage.CurrencyRatesStorage
		gateway func() gateway
	}
	type args struct {
		ready bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "true",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					return nil
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				ready: true,
			},
			want: true,
		},
		{
			name: "false",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					return nil
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				ready: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRater(tt.fields.cfg, tt.fields.storage(), tt.fields.gateway())
			r.ready = tt.args.ready
			if ready := r.Ready(); ready != tt.want {
				t.Errorf("Ready() = %v, want %v", ready, tt.want)
			}
		})
	}
}

func Test_rater_Exchange(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		cfg     config.CurrencyConfig
		storage func() storage.CurrencyRatesStorage
		gateway func() gateway
	}
	type args struct {
		value int64
		from  string
		to    string
		date  time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "equal currencies",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					return nil
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				value: 10000,
				from:  "EUR",
				to:    "EUR",
				date:  today,
			},
			want:    10000,
			wantErr: false,
		},
		{
			name: "RUB to USD base error",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Get("RUB", today).Return(int64(0), false, simpleError)
					return m
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				value: 1000000,
				from:  "RUB",
				to:    "USD",
				date:  today,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "RUB to USD base no rates",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Get("RUB", today).Return(int64(0), false, nil)
					return m
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				value: 1000000,
				from:  "RUB",
				to:    "USD",
				date:  today,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "RUB to EUR target error",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Get("RUB", today).Return(int64(200), true, nil)
					m.EXPECT().Get("EUR", today).Return(int64(0), false, simpleError)
					return m
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				value: 1000000,
				from:  "RUB",
				to:    "EUR",
				date:  today,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "RUB to EUR target no rates",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Get("RUB", today).Return(int64(200), true, nil)
					m.EXPECT().Get("EUR", today).Return(int64(0), false, nil)
					return m
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				value: 1000000,
				from:  "RUB",
				to:    "EUR",
				date:  today,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "RUB to USD success",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyRatesStorage {
					m := smocks.NewMockCurrencyRatesStorage(ctrl)
					m.EXPECT().Get("RUB", today).Return(int64(200), true, nil)
					return m
				},
				gateway: func() gateway {
					return nil
				},
			},
			args: args{
				value: 1500000,
				from:  "RUB",
				to:    "USD",
				date:  today,
			},
			want:    30000,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRater(tt.fields.cfg, tt.fields.storage(), tt.fields.gateway())
			value, err := r.Exchange(tt.args.value, tt.args.from, tt.args.to, tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exchange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if value != tt.want {
				t.Errorf("Exchange() value = %v, want %v", value, tt.want)
			}
		})
	}
}
