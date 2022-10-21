package cbr

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/currency/rates/cbr"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"golang.org/x/text/encoding/charmap"
)

const (
	brokenXml = `<?xml version="1.0" encoding="windows-1251"?><ValCurs Date="21.10.2022"><Valute ID="R01235"><NumCode>840</NumCode><CharCode>USD</CharCode><Nominal>1</Nominal><Name>Доллар США</Name>`
	validXml  = `<?xml version="1.0" encoding="windows-1251"?><ValCurs Date="21.10.2022" name="Foreign Currency Market"><Valute ID="R01235"><NumCode>840</NumCode><CharCode>USD</CharCode><Nominal>1</Nominal><Name>Доллар США</Name><Value>61,1958</Value></Valute><Valute ID="R01239"><NumCode>978</NumCode><CharCode>EUR</CharCode><Nominal>1</Nominal><Name>Евро</Name><Value>59,8378</Value></Valute><Valute ID="R01375"><NumCode>156</NumCode><CharCode>CNY</CharCode><Nominal>10</Nominal><Name>Китайских юаней</Name><Value>83,7320</Value></Valute></ValCurs>`
)

var (
	simpleError = errors.New("error")
)

func Test_gateway_FetchRates(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name        string
		client      func() httpClient
		initializer func(g *gateway)
		wantRates   map[string]int64
		wantDate    time.Time
		wantErr     bool
	}{
		{
			name: "invalid url",
			client: func() httpClient {
				return nil
			},
			initializer: func(g *gateway) {
				g.url = string(rune(0x7f))
			},
			wantRates: nil,
			wantDate:  time.Time{},
			wantErr:   true,
		},
		{
			name: "client error",
			client: func() httpClient {
				m := mocks.NewMockhttpClient(ctrl)
				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(nil, simpleError)
				return m
			},
			wantRates: nil,
			wantDate:  time.Time{},
			wantErr:   true,
		},
		{
			name: "decode error",
			client: func() httpClient {
				m := mocks.NewMockhttpClient(ctrl)

				xml, _ := charmap.Windows1251.NewEncoder().String(brokenXml)
				w := httptest.NewRecorder()
				_, _ = io.WriteString(w, xml)

				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(w.Result(), nil)
				return m
			},
			wantRates: nil,
			wantDate:  time.Time{},
			wantErr:   true,
		},
		{
			name: "invalid date",
			client: func() httpClient {
				m := mocks.NewMockhttpClient(ctrl)

				modifiedXml := strings.Replace(validXml, `Date="21.10.2022"`, `Date="10.21.2022"`, 1)
				xml, _ := charmap.Windows1251.NewEncoder().String(modifiedXml)
				w := httptest.NewRecorder()
				_, _ = io.WriteString(w, xml)

				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(w.Result(), nil)
				return m
			},
			wantRates: nil,
			wantDate:  time.Time{},
			wantErr:   true,
		},
		{
			name: "success",
			client: func() httpClient {
				m := mocks.NewMockhttpClient(ctrl)

				xml, _ := charmap.Windows1251.NewEncoder().String(validXml)
				w := httptest.NewRecorder()
				_, _ = io.WriteString(w, xml)

				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(w.Result(), nil)
				return m
			},
			wantRates: map[string]int64{
				"USD": 611958,
				"EUR": 598378,
				"CNY": 83732,
			},
			wantDate: utils.TruncateToDate(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGateway(tt.client())
			if tt.initializer != nil {
				tt.initializer(g)
			}
			rates, date, err := g.FetchRates(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchRates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(rates, tt.wantRates) {
				t.Errorf("FetchRates() rates = %v, want %v", rates, tt.wantRates)
			}
			if !reflect.DeepEqual(date, tt.wantDate) {
				t.Errorf("FetchRates() date = %v, want %v", date, tt.wantDate)
			}
		})
	}
}
