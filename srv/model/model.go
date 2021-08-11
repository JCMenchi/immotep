package model

import (
	"fmt"
	"strings"
	"time"

	geojson "github.com/paulmach/go.geojson"
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
	CityCode       string
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

type Region struct {
	Code       string       `gorm:"primaryKey" json:"code"`
	Name       string       `json:"nom"`
	Department []Department `gorm:"foreignKey:CodeRegion;references:Code"`
}

type Department struct {
	Code       string `gorm:"primaryKey" json:"code"`
	Name       string `json:"nom"`
	CodeRegion string `json:"codeRegion"`
	City       []City `gorm:"foreignKey:CodeDepartment;references:Code"`
}

type City struct {
	Code           string `gorm:"primaryKey" json:"code"`
	Name           string `json:"nom"`
	ZipCode        int
	Population     int      `json:"population"`
	Contour        string   `json:"contour"`
	CodeDepartment string   `json:"codeDepartement"`
	CodesPostaux   []string `gorm:"-" json:"codesPostaux"`
}

func ConnectToDB(dsn string) *gorm.DB {

	if strings.HasPrefix(dsn, "postgres:") {
		db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}),
			&gorm.Config{CreateBatchSize: 1000, Logger: logger.Default.LogMode(logger.Info)})

		if err != nil {
			fmt.Printf("ConnectToDB error: %v\n", err)
			return nil
		}

		err = db.AutoMigrate(&Transaction{}, &Region{}, &Department{}, &City{})
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

		db.AutoMigrate(&Transaction{}, &Region{}, &Department{}, &City{})

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
	} else if limit > 500 {
		limit = 500
	}

	result := db.Where(whereClause).Limit(limit).Find(&pois)

	if result.Error != nil {
		fmt.Printf("err: %v\n", result.Error)
		return nil
	}

	return pois
}

/**
  SELECT tr.city as name,  avg(price_psqm) as ps, cities.contour as geojson FROM transactions as tr
  LEFT JOIN cities ON tr.city_code = cities.code WHERE tr.department_code = 29
  group by tr.city_code;

*/
type CityInfo struct {
	Name        string          `json:"name"`
	AvgPriceSQM float64         `json:"avgprice"`
	Contour     geojson.Feature `json:"contour"`
}

func GetCityDetails(db *gorm.DB, limit, dep int) []CityInfo {

	if limit <= 0 {
		limit = 100
	}

	var cityinfos []CityInfo = make([]CityInfo, 0, limit)

	rows, err := db.Debug().Select("transactions.city as name, AVG(transactions.price_psqm) as avg_price_psqm, cities.contour as geojson, cities.code as citycode, cities.population as population").
		Joins("LEFT JOIN cities ON cities.code = transactions.city_code").
		Where("department_code = ?", dep).
		Table("transactions").
		Group("city_code").
		Limit(limit).Rows()

	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var name, geojsonString, code string
		var avg_price_psqm float64
		var population int
		rows.Scan(&name, &avg_price_psqm, &geojsonString, &code, &population)

		var info CityInfo
		info.Name = name
		info.AvgPriceSQM = avg_price_psqm

		geom, err := geojson.UnmarshalGeometry([]byte(geojsonString))
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			info.Contour.Geometry = geom
			info.Contour.SetProperty("avgprice", avg_price_psqm)
			info.Contour.SetProperty("city", code)
			info.Contour.SetProperty("population", population)
			cityinfos = append(cityinfos, info)
		}
	}

	return cityinfos
}
