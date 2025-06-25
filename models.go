package gorm_auto_explain

import (
	"time"
)

type Result struct {
	Duration            time.Duration `json:"duration"`
	Message             string        `json:"message"`
	QueryCost           float64       `json:"query_cost"`
	UsingTemporaryTable bool          `json:"using_temporary_table"`
	Raw                 string        `json:"raw"`
	Query               string        `json:"query"`
}
type OnExplainCb func(result Result)
