//go:build unit

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
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model/currency/cbr"
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

type mocksInitializer struct {
	client func(*mocks.MockhttpClient)
}

func setupGateway(t *testing.T, i mocksInitializer) *cbrGateway {
	ctrl := gomock.NewController(t)

	clientMock := mocks.NewMockhttpClient(ctrl)
	if i.client != nil {
		i.client(clientMock)
	}

	return NewCbrGateway(clientMock)
}

func Test_gateway_FetchRates(t *testing.T) {
	t.Parallel()

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		g := setupGateway(t, mocksInitializer{})
		g.url = string(rune(0x7f))

		// ACT
		rates, date, err := g.FetchRates(context.Background())

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, rates)
		assert.Empty(t, date)
	})

	t.Run("client error", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		g := setupGateway(t, mocksInitializer{
			client: func(m *mocks.MockhttpClient) {
				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(nil, simpleError)
			},
		})

		// ACT
		rates, date, err := g.FetchRates(context.Background())

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, rates)
		assert.Empty(t, date)
	})

	t.Run("decode error", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		g := setupGateway(t, mocksInitializer{
			client: func(m *mocks.MockhttpClient) {
				xml, _ := charmap.Windows1251.NewEncoder().String(brokenXml)
				w := httptest.NewRecorder()
				_, _ = io.WriteString(w, xml)

				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(w.Result(), nil)
			},
		})

		// ACT
		rates, date, err := g.FetchRates(context.Background())

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, rates)
		assert.Empty(t, date)
	})

	t.Run("invalid date", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		g := setupGateway(t, mocksInitializer{
			client: func(m *mocks.MockhttpClient) {
				modifiedXml := strings.Replace(validXml, `Date="21.10.2022"`, `Date="10.21.2022"`, 1)
				xml, _ := charmap.Windows1251.NewEncoder().String(modifiedXml)
				w := httptest.NewRecorder()
				_, _ = io.WriteString(w, xml)

				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(w.Result(), nil)
			},
		})

		// ACT
		rates, date, err := g.FetchRates(context.Background())

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, rates)
		assert.Empty(t, date)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		g := setupGateway(t, mocksInitializer{
			client: func(m *mocks.MockhttpClient) {
				xml, _ := charmap.Windows1251.NewEncoder().String(validXml)
				w := httptest.NewRecorder()
				_, _ = io.WriteString(w, xml)

				var req = reflect.TypeOf((**http.Request)(nil)).Elem()
				m.EXPECT().Do(gomock.AssignableToTypeOf(req)).Return(w.Result(), nil)
			},
		})

		// ACT
		rates, date, err := g.FetchRates(context.Background())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, map[string]int64{
			"USD": 611958,
			"EUR": 598378,
			"CNY": 83732,
		}, rates)
		assert.Equal(t, utils.TruncateToDate(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)), date)
	})
}
