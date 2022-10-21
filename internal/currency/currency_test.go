package currency

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
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
	}
)

func Test_manager_Get(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		cfg     config.CurrencyConfig
		storage func() storage.CurrencyStorage
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "error",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					m := mocks.NewMockCurrencyStorage(ctrl)
					m.EXPECT().Get(testUser).Return("", false, simpleError)
					return m
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					m := mocks.NewMockCurrencyStorage(ctrl)
					m.EXPECT().Get(testUser).Return("EUR", true, nil)
					return m
				},
			},
			want:    "EUR",
			wantErr: false,
		},
		{
			name: "default",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					m := mocks.NewMockCurrencyStorage(ctrl)
					m.EXPECT().Get(testUser).Return("", false, nil)
					return m
				},
			},
			want:    "USD",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.fields.cfg, tt.fields.storage())
			currency, err := m.Get(testUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if currency != tt.want {
				t.Errorf("Get() currency = %v, want %v", currency, tt.want)
			}
		})
	}
}

func Test_manager_Set(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		cfg     config.CurrencyConfig
		storage func() storage.CurrencyStorage
	}
	type args struct {
		curr string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "unknown",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					return nil
				},
			},
			args: args{
				curr: "RUB",
			},
			wantErr: true,
		},
		{
			name: "error",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					m := mocks.NewMockCurrencyStorage(ctrl)
					m.EXPECT().Set(testUser, "EUR").Return(simpleError)
					return m
				},
			},
			args: args{
				curr: "EUR",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					m := mocks.NewMockCurrencyStorage(ctrl)
					m.EXPECT().Set(testUser, "USD").Return(nil)
					return m
				},
			},
			args: args{
				curr: "USD",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.fields.cfg, tt.fields.storage())
			if err := m.Set(testUser, tt.args.curr); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_ListCurrenciesCodesWithFlags(t *testing.T) {
	type fields struct {
		cfg     config.CurrencyConfig
		storage func() storage.CurrencyStorage
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "list",
			fields: fields{
				cfg: defaultCfg,
				storage: func() storage.CurrencyStorage {
					return nil
				},
			},
			want: []string{
				"USD $",
				"EUR ¢",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.fields.cfg, tt.fields.storage())
			if list := m.ListCurrenciesCodesWithFlags(); !reflect.DeepEqual(list, tt.want) {
				t.Errorf("ListCurrenciesCodesWithFlags() = %v, want %v", list, tt.want)
			}
		})
	}
}
