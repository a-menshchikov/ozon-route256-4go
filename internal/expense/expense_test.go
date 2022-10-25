package expense

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
	today       = utils.TruncateToDate(time.Now())
	tomorrow    = today.Add(24 * time.Hour)
	simpleError = errors.New("error")
)

func Test_expenser_Add(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseStorage
	}
	type args struct {
		date     time.Time
		amount   int64
		currency string
		category string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "negative amount",
			fields: fields{
				storage: func() storage.ExpenseStorage {
					return nil
				},
			},
			args: args{
				date:     today,
				amount:   -10000,
				currency: "RUB",
				category: "",
			},
			wantErr: true,
		},
		{
			name: "feature expense",
			fields: fields{
				storage: func() storage.ExpenseStorage {
					return nil
				},
			},
			args: args{
				date:     tomorrow,
				amount:   10000,
				currency: "RUB",
				category: "",
			},
			wantErr: true,
		},
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseStorage {
					m := mocks.NewMockExpenseStorage(ctrl)
					m.EXPECT().Add(testUser, types.ExpenseItem{
						Date:     today,
						Amount:   20000,
						Currency: "RUB",
					}, "taxi").Return(simpleError)
					return m
				},
			},
			args: args{
				date:     today,
				amount:   20000,
				currency: "RUB",
				category: "taxi",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				storage: func() storage.ExpenseStorage {
					m := mocks.NewMockExpenseStorage(ctrl)
					m.EXPECT().Add(testUser, types.ExpenseItem{
						Date:     today,
						Amount:   150000,
						Currency: "RUB",
					}, "coffee").Return(nil)
					return m
				},
			},
			args: args{
				date:     today,
				amount:   150000,
				currency: "RUB",
				category: "coffee",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExpenser(tt.fields.storage())
			if err := e.Add(testUser, tt.args.date, tt.args.amount, tt.args.currency, tt.args.category); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_expenser_Report(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseStorage
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string][]types.ExpenseItem
		wantErr bool
	}{
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseStorage {
					m := mocks.NewMockExpenseStorage(ctrl)
					m.EXPECT().List(testUser, today).Return(nil, simpleError)
					return m
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				storage: func() storage.ExpenseStorage {
					m := mocks.NewMockExpenseStorage(ctrl)
					m.EXPECT().List(testUser, today).Return(map[string][]types.ExpenseItem{
						"taxi": {
							{
								Date:     today,
								Amount:   100000,
								Currency: "USD",
							},
							{
								Date:     today,
								Amount:   120000,
								Currency: "EUR",
							},
						},
						"coffee": {
							{
								Date:     today,
								Amount:   1200000,
								Currency: "RUB",
							},
						},
					}, nil)
					return m
				},
			},
			want: map[string][]types.ExpenseItem{
				"taxi": {
					{
						Date:     today,
						Amount:   100000,
						Currency: "USD",
					},
					{
						Date:     today,
						Amount:   120000,
						Currency: "EUR",
					},
				},
				"coffee": {
					{
						Date:     today,
						Amount:   1200000,
						Currency: "RUB",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExpenser(tt.fields.storage())
			data, err := e.Report(testUser, today)
			if (err != nil) != tt.wantErr {
				t.Errorf("Report() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(data, tt.want) {
				t.Errorf("Report() data = %v, want %v", data, tt.want)
			}
		})
	}
}
