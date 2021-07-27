package model

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
	BUG: in gorm CreateIndexAfterCreateTable is forced to true
	and autoIncrement is ignored if field is declared as primaryKey
	but if autoIncrement is set it becomes a primaryKey automagically
*/
type Transaction struct {
	TrId           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Date           time.Time
	Address        string `json:"address"`
	ZipCode        int
	City           string
	DepartmentCode string
	Price          float64
	PricePSQM      float64
	Area           int
	FullArea       int
	NbRoom         int
	Cadastre       string
	TypeCulture    string
	Lat            float64
	Long           float64
}

func ConnectToDB(dsn string) *gorm.DB {

	if strings.HasPrefix(dsn, "postgres:") {
		db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}),
			&gorm.Config{CreateBatchSize: 1000, Logger: logger.Default.LogMode(logger.Warn)})

		if err != nil {
			fmt.Printf("ConnectToDB error: %v\n", err)
			return nil
		}

		err = db.AutoMigrate(&Transaction{})
		if err != nil {
			fmt.Printf("AutoMigrate DB error: %v\n", err.Error())
			return nil
		}

		return db
	} else if strings.HasPrefix(dsn, "file:") {
		sl := sqlite.Open(dsn)
		db, err := gorm.Open(sl,
			&gorm.Config{CreateBatchSize: 100000, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Warn)})

		if err != nil {
			fmt.Printf("ConnectToDB error: %v\n", err)
			return nil
		}

		db.AutoMigrate(&Transaction{})

		return db
	}

	return nil
}
