/*
*

Get all year with data
select distinct(EXTRACT(year FROM date)) from transactions;

Select EXTRACT(year FROM transactions.date) as year, transactions.city_code as code, MIN(cities.name) as name, AVG(transactions.price_psqm) as avg_price_psqm
from transactions
GROUP BY year, city_code;
*/
package model

import (
	"github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CityYearlyAgg struct {
	Code     string  `gorm:"primaryKey" json:"code"`
	Year     int     `gorm:"primaryKey" json:"year"`
	Name     string  `json:"nom"`
	AvgPrice float64 `json:"avg_price"`
	Increase float64 `json:"increase"`
}

func AggregateData(dsn string) {
	db := ConnectToDB(dsn)

	db.AutoMigrate(&CityYearlyAgg{})

	cleanAggregate(db)
	log.Infof("Aggregate Data for Cities...\n")
	aggregateCities(db)
	log.Infof("All computation done.\n")
}

func cleanAggregate(db *gorm.DB) {
	db.Exec("TRUNCATE city_yearly_aggs;")
}

func aggregateCities(db *gorm.DB) {
	rows, err := db.Select("EXTRACT(year FROM transactions.date) as year, transactions.city_code as code, MIN(cities.name) as name, AVG(transactions.price_psqm) as avg_price_psqm").
		Table("transactions").
		Joins("LEFT JOIN cities on cities.code = transactions.city_code").
		Group("year").Group("transactions.city_code").
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "code"}, Desc: false},
			{Column: clause.Column{Name: "year"}, Desc: false},
		}}).
		Rows()

	if err != nil {
		log.Errorf("aggregateCities err: %v\n", err)
		return
	}
	defer rows.Close()

	batchSize := 200
	var city2update = make([]map[string]interface{}, 0, batchSize)

	prev_avg := 0.0
	prev_code := ""

	for rows.Next() {
		var code string
		var name string
		var avg_price_psqm float64
		var year int
		increase := 0.0

		rows.Scan(&year, &code, &name, &avg_price_psqm)

		if code == prev_code {
			increase = (avg_price_psqm - prev_avg) / prev_avg
		}
		prev_code = code
		prev_avg = avg_price_psqm

		city2update = append(city2update, map[string]interface{}{"year": year, "code": code, "name": name, "avg_price": avg_price_psqm, "increase": increase})

		log.Debugf("City (%v) year %v avg psqm: %.0f€\n", code, year, avg_price_psqm)
	}

	if len(city2update) <= 0 {
		log.Infof("Nothing to aggregate for cities.\n")
		return
	}
	bar := pb.Default.Start(len(city2update))

	for {

		l := len(city2update)
		b := l - 1
		if l > batchSize {
			b = batchSize
		}

		batch := city2update[0:b]

		updresult := db.Table("city_yearly_aggs").Create(&batch)

		if updresult.Error != nil {
			log.Errorf("Error aggregateCities update: %.0f\n", updresult.Error)
		}

		bar.Add(b)
		if l-1 == b {
			break
		}
		city2update = city2update[b+1:]
	}

	bar.Add(int(bar.Total() - bar.Current()))
	bar.Finish()
}
