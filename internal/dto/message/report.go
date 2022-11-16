package message

import (
	"time"
)

type Report struct {
	From     time.Time
	Currency string
}
