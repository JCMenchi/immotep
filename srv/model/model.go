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
			&gorm.Config{CreateBatchSize: 1000, Logger: logger.Default.LogMode(logger.Info)})

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

type TransactionPOI struct {
	TrId    uint64    `gorm:"primaryKey" json:"id"`
	Date    time.Time `json:"date"`
	Address string    `json:"address"`
	City    string    `json:"city"`
	Price   float64   `json:"price"`
	Area    int       `json:"area"`
	Lat     float64   `json:"lat"`
	Long    float64   `json:"long"`
}

func (TransactionPOI) TableName() string {
	return "transactions"
}

func GetPOI(db *gorm.DB, limit, zip, dep int, after string) []TransactionPOI {
	if db == nil {
		return nil
	}

	var pois []TransactionPOI

	whereClause := "lat > 0"
	if zip > 0 {
		whereClause = fmt.Sprintf("%v AND zip_code = %v", whereClause, zip)
	}
	if dep > 0 {
		whereClause = fmt.Sprintf("%v AND department_code = %v", whereClause, dep)
	}

	if after != "" {
		whereClause = fmt.Sprintf("%v AND date > \"%v\"", whereClause, after)
	}

	if limit <= 0 {
		limit = 100
	}

	result := db.Where(whereClause).Limit(limit).Find(&pois)

	if result.Error != nil {
		fmt.Printf("err: %v\n", result.Error)
		return nil
	}

	return pois
}

func GetPOIFromBounds(db *gorm.DB, NElat, NELong, SWlat, SWLong float64, limit, dep int, after string) []TransactionPOI {
	var pois []TransactionPOI

	whereClause := fmt.Sprintf("lat < %v AND lat > %v AND long < %v AND long > %v", NElat, SWlat, NELong, SWLong)

	if dep > 0 {
		whereClause = fmt.Sprintf("%v AND department_code = %v", whereClause, dep)
	}

	if after != "" {
		whereClause = fmt.Sprintf("%v AND date > \"%v\"", whereClause, after)
	}

	if limit <= 0 {
		limit = 100
	}

	result := db.Where(whereClause).Limit(limit).Find(&pois)

	if result.Error != nil {
		fmt.Printf("err: %v\n", result.Error)
		return nil
	}

	return pois
}
