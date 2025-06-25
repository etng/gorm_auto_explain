package main

import (
	"encoding/json"
	"fmt"
	autoExplain "github.com/etng/gorm_auto_explain"
	dsn "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	db := initDb(dsn.Config{
		User:                 "root",
		Passwd:               "toor",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "example",
		Params:               nil,
		Collation:            "utf8mb4_general_ci",
		AllowNativePasswords: true,
		ParseTime:            true,
	})
	autoExplain.Toggle(false)
	Must(db.AutoMigrate(&User{}))
	autoExplain.Toggle(true)
	db.Use(autoExplain.InitPlugin(0).OnExplain(func(result autoExplain.Result) {
		if result.Message != "" {
			log.Println("notice:", result.Message)
		}
		if result.QueryCost > 100 {
			log.Printf("heavy sql you should try to optimize found, score=%.2f, duration=%s", result.QueryCost, result.Duration)
		}
		log.Printf("explain detail: query: %s\ncost: %.2f\n duration:%s\n%s", result.Query, result.QueryCost, result.Duration, result.Raw)
	}))
	for i := 0; i < 10000; i++ {
		user := &User{
			Username: "user" + strconv.Itoa(i),
			Password: "password" + strconv.Itoa(i),
			JoinedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		Must(UpdateOnConflict(
			db,
			[]string{"username", "password", "joined_at"},
			"username",
		).Create(user).Error)
	}
	log.Println("try query")
	func() {
		rows, err := db.Model(&User{}).
			Where("username = ?", "admin").
			Where("joined_at > ?", time.Now().Add(-time.Hour*24*365*2)).
			Rows()
		Must(err)
		defer rows.Close()
		for rows.Next() {
			var user User
			Must(rows.Scan(&user))
			log.Println(user.String())
		}
	}()
	func() {
		var results []struct {
			TheDate string
			Count   int
		}
		Must(db.Model(&User{}).
			Select("DATE_FORMAT(joined_at, '%Y-%m-%d 00:00:00') as the_date, COUNT(*) as count").
			Where("joined_at >= ?", time.Now().AddDate(0, 0, -3)).
			Group("DATE_FORMAT(joined_at, '%Y-%m-%d 00:00:00')").
			Order("the_date DESC").
			Scan(&results).Error)
		for _, result := range results {
			fmt.Println(strings.TrimSuffix(result.TheDate, " 00:00:00"), result.Count)
		}
	}()
}
func initDb(dsnConfig dsn.Config) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSNConfig:                 &dsnConfig,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: logger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
		TranslateError: true,
	})
	Must(err)
	sqlDB, err := db.DB()
	Must(err)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
	JoinedAt time.Time
}

func (user User) String() string {
	b, e := json.MarshalIndent(user, "", "  ")
	Must(e)
	return string(b)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
func UpdateOnConflict(db *gorm.DB, assignColumns []string, keyColumns ...string) *gorm.DB {
	if len(keyColumns) == 0 {
		keyColumns = append(keyColumns, "id")
	}
	return db.Clauses(clause.OnConflict{
		Columns: func() (columns []clause.Column) {
			for _, column := range keyColumns {
				columns = append(columns, clause.Column{Name: column})
			}
			return
		}(),
		DoUpdates: clause.AssignmentColumns(assignColumns),
	})
}
