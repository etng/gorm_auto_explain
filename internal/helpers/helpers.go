package helpers

import (
	"gorm.io/gorm"
	"strconv"
)

func ParseFloat64Default(s string, dflt float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return dflt
}

func Until(cbList ...func() error) (err error) {
	for _, cb := range cbList {
		err = cb()
		if err != nil {
			return
		}
	}
	return
}
func ResetDb(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true})
}

type ExplainRawData struct {
	QueryBlock struct {
		Message  string `json:"message"`
		CostInfo struct {
			QueryCost string `json:"query_cost"`
		} `json:"cost_info"`
	} `json:"query_block"`
}
