package exchanger

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency"
	"golang.org/x/text/encoding/charmap"
)

const (
	_ratesAPI     = "https://www.cbr.ru/scripts/XML_daily.asp"
	_ratesTimeout = 3 * time.Second
)

type Exchanger struct {
	mu sync.RWMutex

	ready  bool
	config config.Currency
	rates  map[string]int64 // 4 decimal digits

	client *http.Client
}

func NewCbrExchanger(cfg config.Currency) *Exchanger {
	rates := make(map[string]int64)
	for _, currConfig := range cfg.Available {
		rates[currConfig.Code] = 0
	}

	return &Exchanger{
		config: cfg,
		rates:  rates,
		client: http.DefaultClient,
	}
}

func (e *Exchanger) ExchangeFromBase(value int64, curr string) (int64, error) {
	if curr == e.config.Base {
		return value, nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if rate, ok := e.rates[curr]; ok {
		return value * 10000 / rate, nil
	}

	return 0, currency.ErrUnknownCurrency
}

func (e *Exchanger) Ready() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.ready
}

func (e *Exchanger) ExchangeToBase(value int64, curr string) (int64, error) {
	if curr == e.config.Base {
		return value, nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if rate, ok := e.rates[curr]; ok {
		return value * rate / 10000, nil
	}

	return 0, currency.ErrUnknownCurrency
}

func (e *Exchanger) ListCurrencies() []string {
	list := make([]string, len(e.config.Available))

	for k, curr := range e.config.Available {
		list[k] = curr.Code + " " + curr.Flag
	}

	return list
}

func (e *Exchanger) Run(ctx context.Context) {
	e.refreshRates(ctx)
	ticker := time.NewTicker(e.config.RefreshInterval)

	select {
	case <-ctx.Done():
		ticker.Stop()
		return

	case <-ticker.C:
		e.refreshRates(ctx)
	}
}

func (e *Exchanger) refreshRates(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	list, err := fetchCurrentRates(ctx, e.client)
	if err != nil {
		log.Println("rates refresh failed:", err.Error())
		return
	}

	e.ready = false
	for _, curr := range list.Currencies {
		if _, ok := e.rates[curr.CharCode]; ok {
			e.rates[curr.CharCode] = int64(curr.Value*10000) / int64(curr.Nominal)
		}
	}
	e.ready = true
}

func fetchCurrentRates(ctx context.Context, client *http.Client) (*CurrencyList, error) {
	ctx, cancel := context.WithTimeout(ctx, _ratesTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", _ratesAPI, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create request for fetch rates")
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "fetch rates request failed")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "fetch rates response body read failed")
	}

	body, err = charmap.Windows1251.NewDecoder().Bytes(body)
	if err != nil {
		return nil, errors.Wrap(err, "fetch rates response decode failed (win-1251)")
	}

	body = bytes.Replace(body, []byte(` encoding="windows-1251"`), []byte(""), -1)
	body = bytes.Replace(body, []byte(","), []byte("."), -1)
	decoder := xml.NewDecoder(bytes.NewReader(body))

	var list CurrencyList
	if err := decoder.Decode(&list); err != nil {
		return nil, errors.Wrap(err, "fetch rates response decode failed")
	}

	return &list, nil
}
