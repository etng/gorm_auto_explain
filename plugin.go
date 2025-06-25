package gorm_auto_explain

import (
	"encoding/json"
	"github.com/etng/gorm_auto_explain/internal/constants"
	"github.com/etng/gorm_auto_explain/internal/helpers"
	"gorm.io/gorm"
	"log"
	"regexp"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

type Plugin struct {
	durationThreshold time.Duration
	onExplainCbList   []OnExplainCb
}

func InitPlugin(durationThreshold time.Duration) *Plugin {
	return &Plugin{durationThreshold: durationThreshold}
}

func (p *Plugin) OnExplain(cb OnExplainCb) *Plugin {
	p.onExplainCbList = append(p.onExplainCbList, cb)
	return p
}

func (p *Plugin) Name() string {
	return "explainPlugin"
}

func (p *Plugin) Initialize(db *gorm.DB) error {
	explainFunc := func(db *gorm.DB) {
		if disabled.Load() {
			if constants.Debug {
				log.Println("jae85ekjgu skipping auto explain for disabled")
			}
			return
		}
		var duration time.Duration
		if iv, ok := db.Statement.Get(constants.ExplainKey); ok {
			duration = time.Since(iv.(time.Time))
		}
		if p.durationThreshold != 0 && duration < p.durationThreshold {
			if constants.Debug {
				log.Println("jae85ekjgu skipping auto explain for duration threshold", duration)
			}
			return
		}
		if constants.Debug {
			log.Printf("xh2xjtbyvx captured SQL: %s, Vars: %v %s", db.Statement.SQL.String(), db.Statement.Vars, duration)
		}
		if db.Error != nil || db.Statement.SQL.String() == "" {
			log.Printf("86rjbjfjjj skipping auto explain for sql error: error=%v, SQL empty=%v", db.Error, db.Statement.SQL.String() == "")
			return
		}
		if strings.HasPrefix(db.Statement.SQL.String(), constants.ExplainClause) {
			if constants.Debug {
				log.Println("rkendgns6c skipping auto explain for already explain")
			}
			return
		}
		for _, pattern := range []*regexp.Regexp{
			patternFromInformationSchema,
			patternSelectDb,
		} {
			if pattern.MatchString(db.Statement.SQL.String()) {
				if constants.Debug {
					log.Println("anjw4f2mj6 skipping auto explain for internal sql")
				}
				return
			}
		}
		if constants.Debug {
			log.Println("5dy6j7szw7 auto explain call stack", string(debug.Stack()))

		}
		var explainResult string
		// do not use this kind, environment not clean
		//err := db.Raw(ExplainClause + db.Statement.SQL.String()).Scan(&explainResult).Error
		err := helpers.ResetDb(db).Raw(constants.ExplainClause+db.Statement.SQL.String(), db.Statement.Vars...).Scan(&explainResult).Error
		if err != nil {
			log.Printf("5dy6j7szw7 auto explain failed: %v", err)
			return
		}
		if constants.Debug {
			log.Println("2526vnkhdq auto explain json result:", explainResult)
		}
		var explainData helpers.ExplainRawData

		if err = json.Unmarshal([]byte(explainResult), &explainData); err != nil {
			log.Printf("2vej7gp6gh auto explain json parse failed: %v", err)
			return
		}

		result := Result{
			Duration:  duration,
			Raw:       explainResult,
			Message:   explainData.QueryBlock.Message,
			Query:     db.Statement.SQL.String(),
			QueryCost: helpers.ParseFloat64Default(explainData.QueryBlock.CostInfo.QueryCost, 0),
		}

		if constants.Debug {
			log.Printf("6us4pg3haa auto explain result: %+v", result)
		}
		for _, cb := range p.onExplainCbList {
			cb(result)
		}
	}
	explainPrepareFunc := func(db *gorm.DB) {
		db.Statement.Set(constants.ExplainKey, time.Now())
	}
	return helpers.Until(
		func() error {
			return db.Callback().Query().Before("gorm:query").Register(constants.CallbackPrefix+"before_query", explainPrepareFunc)
		}, func() error {
			return db.Callback().Raw().Before("gorm:raw").Register(constants.CallbackPrefix+"before_raw", explainPrepareFunc)
		}, func() error {
			return db.Callback().Row().Before("gorm:row").Register(constants.CallbackPrefix+"before_row", explainPrepareFunc)
		},

		func() error {
			return db.Callback().Query().After("gorm:query").Register(constants.CallbackPrefix+"after_query", explainFunc)
		}, func() error {
			return db.Callback().Raw().After("gorm:raw").Register(constants.CallbackPrefix+"after_raw", explainFunc)
		}, func() error {
			return db.Callback().Row().After("gorm:row").Register(constants.CallbackPrefix+"after_row", explainFunc)
		},
	)
}

var disabled atomic.Bool

func Toggle(b bool) {
	disabled.Store(!b)
}

var (
	patternFromInformationSchema = regexp.MustCompile(`(?i)from\s+information_schema\.`)
	patternSelectDb              = regexp.MustCompile(`(?i)select\s+database\(\s*\)`)
)
