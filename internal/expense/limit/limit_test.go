package limit

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
	simpleError = errors.New("error")
)

func Test_limiter_Get(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseLimitStorage
	}
	type args struct {
		category string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.LimitItem
		wantErr bool
	}{
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Get(testUser, "taxi").Return(types.LimitItem{}, false, simpleError)
					return m
				},
			},
			args: args{
				category: "taxi",
			},
			want:    types.LimitItem{},
			wantErr: true,
		},
		{
			name: "limit not found",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Get(testUser, "").Return(types.LimitItem{}, false, nil)
					return m
				},
			},
			args: args{
				category: "",
			},
			want:    types.LimitItem{},
			wantErr: false,
		},
		{
			name: "found",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
						Total:    1000000,
						Remains:  150000,
						Currency: "USD",
					}, true, nil)
					return m
				},
			},
			args: args{
				category: "coffee",
			},
			want: types.LimitItem{
				Total:    1000000,
				Remains:  150000,
				Currency: "USD",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(tt.fields.storage())
			item, err := l.Get(testUser, tt.args.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(item, tt.want) {
				t.Errorf("Get() item = %v, want %v", item, tt.want)
			}
		})
	}
}

func Test_limiter_Set(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseLimitStorage
	}
	type args struct {
		limit    int64
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
			name: "negative limit",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					return nil
				},
			},
			args: args{
				limit:    -10000,
				currency: "RUB",
				category: "",
			},
			wantErr: true,
		},
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Set(testUser, int64(1000000), "RUB", "coffee").Return(simpleError)
					return m
				},
			},
			args: args{
				limit:    1000000,
				currency: "RUB",
				category: "coffee",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Set(testUser, int64(1200000), "RUB", "coffee").Return(nil)
					return m
				},
			},
			args: args{
				limit:    1200000,
				currency: "RUB",
				category: "coffee",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(tt.fields.storage())
			if err := l.Set(testUser, tt.args.limit, tt.args.currency, tt.args.category); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_limiter_Decrease(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseLimitStorage
	}
	type args struct {
		value    int64
		category string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Decrease(testUser, int64(100000), "taxi").Return(false, simpleError)
					return m
				},
			},
			args: args{
				value:    100000,
				category: "taxi",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Decrease(testUser, int64(200000), "coffee").Return(true, nil)
					return m
				},
			},
			args: args{
				value:    200000,
				category: "coffee",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(tt.fields.storage())
			ok, err := l.Decrease(testUser, tt.args.value, tt.args.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ok != tt.want {
				t.Errorf("Decrease() ok = %v, want %v", ok, tt.want)
			}
		})
	}
}

func Test_limiter_Unset(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseLimitStorage
	}
	type args struct {
		category string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Unset(testUser, "").Return(simpleError)
					return m
				},
			},
			args: args{
				category: "",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().Unset(testUser, "taxi").Return(nil)
					return m
				},
			},
			args: args{
				category: "taxi",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(tt.fields.storage())
			if err := l.Unset(testUser, tt.args.category); (err != nil) != tt.wantErr {
				t.Errorf("Unset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_limiter_List(t *testing.T) {
	ctrl := gomock.NewController(t)

	type fields struct {
		storage func() storage.ExpenseLimitStorage
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]types.LimitItem
		wantErr bool
	}{
		{
			name: "error",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().List(testUser).Return(nil, false, simpleError)
					return m
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "not found",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().List(testUser).Return(nil, false, nil)
					return m
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success",
			fields: fields{
				storage: func() storage.ExpenseLimitStorage {
					m := mocks.NewMockExpenseLimitStorage(ctrl)
					m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
						"taxi": {
							Total:    15000000,
							Remains:  6000000,
							Currency: "RUB",
						},
						"": {
							Total:    1000000,
							Remains:  600000,
							Currency: "EUR",
						},
					}, true, nil)
					return m
				},
			},
			want: map[string]types.LimitItem{
				"taxi": {
					Total:    15000000,
					Remains:  6000000,
					Currency: "RUB",
				},
				"": {
					Total:    1000000,
					Remains:  600000,
					Currency: "EUR",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(tt.fields.storage())
			list, err := l.List(testUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(list, tt.want) {
				t.Errorf("List() list = %v, want %v", list, tt.want)
			}
		})
	}
}
