package cbr

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"golang.org/x/text/encoding/charmap"
)

const (
	_ratesUrl       = "https://www.cbr.ru/scripts/XML_daily.asp"
	_defaultTimeout = 3 * time.Second
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type gateway struct {
	client  httpClient
	url     string
	timeout time.Duration
}

func NewGateway(client httpClient) *gateway {
	return &gateway{
		client:  client,
		url:     _ratesUrl,
		timeout: _defaultTimeout,
	}
}

func (g *gateway) FetchRates(ctx context.Context) (map[string]int64, time.Time, error) {
	list, err := fetchCurrentRates(ctx, g.client, g.url, g.timeout)
	if err != nil {
		return nil, time.Time{}, errors.Wrap(err, "cannot fetch rates")
	}

	date, err := time.Parse("02.01.2006", list.Date)
	if err != nil {
		return nil, time.Time{}, errors.Wrap(err, "cannot parse rates date")
	}
	date = utils.TruncateToDate(date)

	rates := make(map[string]int64)
	for _, curr := range list.Currencies {
		rates[curr.CharCode] = int64(curr.Value*10000) / int64(curr.Nominal)
	}

	return rates, date, nil
}

func fetchCurrentRates(ctx context.Context, client httpClient, url string, timeout time.Duration) (*currencyList, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create request for fetch rates")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fetch rates request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

	var list currencyList
	if err := decoder.Decode(&list); err != nil {
		return nil, errors.Wrap(err, "fetch rates response decode failed")
	}

	return &list, nil
}
