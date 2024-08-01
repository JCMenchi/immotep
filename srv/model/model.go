package model

import (
	"fmt"
	"strings"
	"time"

	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
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
	Code     string  `gorm:"primaryKey" json:"code"`
	Name     string  `json:"nom"`
	Contour  string  `json:"contour"`
	AvgPrice float64 `json:"avg_price"`
	City     []City  `gorm:"foreignKey:CodeRegion;references:Code"`
}

type Department struct {
	Code     string  `gorm:"primaryKey" json:"code"`
	Name     string  `json:"nom"`
	Contour  string  `json:"contour"`
	AvgPrice float64 `json:"avg_price"`
	City     []City  `gorm:"foreignKey:CodeDepartment;references:Code"`
}

type City struct {
	Code           string `gorm:"primaryKey" json:"code"`
	Name           string `json:"nom"`
	NameUpper      string
	ZipCode        int
	Population     int      `json:"population"`
	Contour        string   `json:"contour"`
	CodeDepartment string   `json:"codeDepartement"`
	CodeRegion     string   `json:"codeRegion"`
	AvgPrice       float64  `json:"avg_price"`
	CodesPostaux   []string `gorm:"-" json:"codesPostaux"`
}

func ConnectToDB(dsn string) *gorm.DB {

	if strings.HasPrefix(dsn, "postgres:") {
		db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}),
			&gorm.Config{CreateBatchSize: 1000, Logger: logger.Default.LogMode(logger.Error)})

		if err != nil {
			log.Errorf("ConnectToDB error: %v\n", err)
			return nil
		}

		err = db.AutoMigrate(&Transaction{}, &Region{}, &Department{}, &City{})
		if err != nil {
			log.Errorf("AutoMigrate DB error: %v\n", err.Error())
			return nil
		}

		return db
	} else if strings.HasPrefix(dsn, "file:") {
		sl := sqlite.Open(dsn)
		db, err := gorm.Open(sl,
			&gorm.Config{CreateBatchSize: 100000, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Error)})

		if err != nil {
			log.Errorf("ConnectToDB error: %v\n", err)
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

func GetPOI(db *gorm.DB, limit, zip int, dep string, after string) []TransactionPOI {
	if db == nil {
		return nil
	}

	var pois []TransactionPOI

	whereClause := "lat > 0"
	if zip > 0 {
		whereClause = fmt.Sprintf("%v AND zip_code = %v", whereClause, zip)
	}
	if dep != "" {
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
		log.Errorf("GetPOI err: %v\n", result.Error)
		return nil
	}

	return pois
}

type BoundedTransactionInfo struct {
	Trans       []TransactionPOI `json:"transactions"`
	AvgPrice    float64          `json:"avgprice"`
	AvgPriceSQM float64          `json:"avgprice_sqm"`
}

func GetPOIFromBounds(db *gorm.DB, NElat, NELong, SWlat, SWLong float64, limit int, dep string, after string) *BoundedTransactionInfo {

	var info BoundedTransactionInfo

	whereClause := fmt.Sprintf("lat < %v AND lat > %v AND long < %v AND long > %v", NElat, SWlat, NELong, SWLong)

	if dep != "" {
		whereClause = fmt.Sprintf("%v AND department_code = '%v'", whereClause, dep)
	}

	if after != "" {
		whereClause = fmt.Sprintf("%v AND date > \"%v\"", whereClause, after)
	}

	if limit <= 0 {
		limit = 100
	} else if limit > 500 {
		limit = 500
	}

	result := db.Where(whereClause).Limit(limit).Find(&info.Trans)

	if result.Error != nil {
		log.Errorf("GetPOIFromBounds err: %v\n", result.Error)
		return nil
	}

	rows, err := db.Debug().Select("AVG(transactions.price) as avg_price, AVG(transactions.price_psqm) as avg_price_psqm").
		Where("lat < ? AND lat > ? AND long < ? AND long > ?", NElat, SWlat, NELong, SWLong).
		Table("transactions").
		Rows()

	if err != nil {
		log.Errorf("GetPOIFromBounds err: %v\n", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var avg_price, avg_price_psqm float64

		rows.Scan(&avg_price, &avg_price_psqm)

		info.AvgPrice = avg_price
		info.AvgPriceSQM = avg_price_psqm
	}

	return &info
}

/*
*

	SELECT tr.city as name,  avg(price_psqm) as ps, cities.contour as geojson FROM transactions as tr
	LEFT JOIN cities ON tr.city_code = cities.code WHERE tr.department_code = 29
	group by tr.city_code;
*/
type CityInfo struct {
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	ZipCode     int              `json:"zip"`
	AvgPriceSQM float64          `json:"avgprice"`
	Contour     *geojson.Feature `json:"contour"`
}

func GetCityDetails(db *gorm.DB, dep string) []CityInfo {
	var cities []City

	result := db.Where("code_department = ?", dep).Find(&cities)

	if result.Error != nil {
		log.Errorf("GetRegionDetails err: %v\n", result.Error)
		return nil
	}

	var cityinfos []CityInfo = make([]CityInfo, 0, len(cities))

	for _, c := range cities {
		var info CityInfo
		info.Name = c.Name
		info.Code = c.Code
		info.ZipCode = c.ZipCode
		info.AvgPriceSQM = c.AvgPrice

		feat, err := geojson.UnmarshalFeature([]byte(c.Contour))
		if err != nil {
			log.Errorf("GetCityDetails UnmarshalGeometry err: %v\n", err)
		} else {
			info.Contour = feat
			info.Contour.SetProperty("avgprice", c.AvgPrice)
			info.Contour.SetProperty("city", c.Code)
			info.Contour.SetProperty("population", c.Population)
			cityinfos = append(cityinfos, info)
		}
	}

	return cityinfos
}

type RegionInfo struct {
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	AvgPriceSQM float64          `json:"avgprice"`
	Contour     *geojson.Feature `json:"contour"`
}

func GetRegionDetails(db *gorm.DB) []RegionInfo {

	var regs []Region

	result := db.Find(&regs)

	if result.Error != nil {
		log.Errorf("GetRegionDetails err: %v\n", result.Error)
		return nil
	}

	var reginfos []RegionInfo = make([]RegionInfo, 0, len(regs))

	for _, r := range regs {
		var rinfo RegionInfo
		rinfo.Name = r.Name
		rinfo.Code = r.Code
		//rinfo.AvgPriceSQM = r.AvgPrice

		feat, err := geojson.UnmarshalFeature([]byte(r.Contour))
		if err != nil {
			log.Errorf("GetRegionDetails err: %v\n", err)
		} else {
			rinfo.Contour = feat
			rinfo.Contour.SetProperty("avgprice", rinfo.AvgPriceSQM)
			rinfo.Contour.SetProperty("name", rinfo.Name)
		}

		reginfos = append(reginfos, rinfo)
	}

	return reginfos
}

type DepartmentInfo struct {
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	AvgPriceSQM float64          `json:"avgprice"`
	Contour     *geojson.Feature `json:"contour"`
}

func GetDepartmentDetails(db *gorm.DB) []DepartmentInfo {

	var deps []Department

	result := db.Find(&deps)

	if result.Error != nil {
		log.Errorf("GetDepartmentDetails err: %v\n", result.Error)
		return nil
	}

	var depinfos []DepartmentInfo = make([]DepartmentInfo, 0, len(deps))

	for _, d := range deps {
		var dinfo DepartmentInfo
		dinfo.Name = d.Name
		dinfo.Code = d.Code
		dinfo.AvgPriceSQM = d.AvgPrice

		feat, err := geojson.UnmarshalFeature([]byte(d.Contour))
		if err != nil {
			log.Errorf("GetDepartmentDetails err: %v\n", err)
		} else {
			dinfo.Contour = feat
			dinfo.Contour.SetProperty("avgprice", dinfo.AvgPriceSQM)
			dinfo.Contour.SetProperty("name", dinfo.Name)
		}

		depinfos = append(depinfos, dinfo)
	}

	return depinfos
}
