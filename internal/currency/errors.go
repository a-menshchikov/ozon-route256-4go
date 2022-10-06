package currency

import (
	"github.com/pkg/errors"
)

var (
	ErrUnknownCurrency = errors.New("указана неизвестная валюта")
)
