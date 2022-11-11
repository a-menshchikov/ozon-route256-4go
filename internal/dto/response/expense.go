package response

import (
	"encoding/json"
	"time"
)

type AddExpense struct {
	Ready        bool
	LimitReached bool
	Success      bool
}

type GetReport struct {
	From     time.Time
	Currency string
	Ready    bool
	Data     map[string]int64
	Success  bool
}

func (r GetReport) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

func (r *GetReport) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}
