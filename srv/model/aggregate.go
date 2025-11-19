// Package model provides data models and aggregation routines for the immotep
// application. This file defines yearly aggregate types and functions that
// compute average price-per-square-meter and year-over-year increase for
// cities, departments and regions.
//
// Aggregation strategy:
//   - Use DB SQL to compute average price_psqm grouped by year and geographic
//     unit (city_code, department_code, code_region).
//   - Compute a simple relative increase compared to the previous year for the
//     same geographic code.
//   - Persist results into tables: city_yearly_aggs, department_yearly_aggs,
//     region_yearly_aggs.
//
// Notes:
//   - Aggregation reads from the transactions and geo tables (cities, regions,
//     departments) via SQL JOINs.
//   - Functions use GORM to manage schema and perform inserts in batches to keep
//     memory usage reasonable.
package model

import (
	"fmt"

	"github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CityYearlyAgg stores yearly aggregated statistics for a city.
// Primary key is (Code, Year).
type CityYearlyAgg struct {
	Code     string  `gorm:"primaryKey" json:"code"`
	Year     int     `gorm:"primaryKey" json:"year"`
	Name     string  `json:"nom"`
	AvgPrice float64 `json:"avg_price"`
	Increase float64 `json:"increase"`
}

// DepartmentYearlyAgg stores yearly aggregated statistics for a department.
// Primary key is (Code, Year).
type DepartmentYearlyAgg struct {
	Code     string  `gorm:"primaryKey" json:"code"`
	Year     int     `gorm:"primaryKey" json:"year"`
	Name     string  `json:"nom"`
	AvgPrice float64 `json:"avg_price"`
	Increase float64 `json:"increase"`
}

// RegionYearlyAgg stores yearly aggregated statistics for a region.
// Primary key is (Code, Year).
type RegionYearlyAgg struct {
	Code     string  `gorm:"primaryKey" json:"code"`
	Year     int     `gorm:"primaryKey" json:"year"`
	Name     string  `json:"nom"`
	AvgPrice float64 `json:"avg_price"`
	Increase float64 `json:"increase"`
}

// AggregateData orchestrates the full aggregation process.
//
// It:
// - Ensures aggregate tables exist (AutoMigrate).
// - Clears any existing aggregate rows.
// - Runs per-entity aggregation routines for cities, departments and regions.
func AggregateData(dsn string) {
	db := ConnectToDB(dsn)

	db.AutoMigrate(&CityYearlyAgg{})
	db.AutoMigrate(&DepartmentYearlyAgg{})
	db.AutoMigrate(&RegionYearlyAgg{})

	cleanAggregate(db)
	log.Infof("Aggregate Data for Cities...\n")
	aggregateCities(db)
	log.Infof("Aggregate Data for Departments...\n")
	aggregateDepartments(db)
	log.Infof("Aggregate Data for Regions...\n")
	aggregateRegions(db)
	log.Infof("All computation done.\n")
}

// cleanAggregate truncates the aggregate tables to remove any previous results.
//
// This ensures a fresh computation when AggregateData is called.
func cleanAggregate(db *gorm.DB) {
	db.Exec("TRUNCATE city_yearly_aggs;")
	db.Exec("TRUNCATE region_yearly_aggs;")
	db.Exec("TRUNCATE department_yearly_aggs;")
}

// aggregateCities computes yearly average price per sqm for each city and
// inserts the results into the city_yearly_aggs table.
//
// Behavior:
//   - Uses a SQL query joining transactions and cities, grouped by year and city.
//   - Computes a simple year-over-year relative increase using the previous
//     row's average for the same city code (as rows are ordered by code,year).
//   - Inserts results in batches and shows a progress bar.
func aggregateCities(db *gorm.DB) {
	colList := fmt.Sprintf("%s as year, transactions.city_code as code, MIN(cities.name) as name, AVG(transactions.price_psqm) as avgPricePSQM",
		func() string {
			if db.Dialector.Name() == "sqlite" {
				log.Debugf("Using SQLITE year extract syntax.\n")
				return SQLITE_QUERY_YEAR_EXTRACT
			} else {
				log.Debugf("Using POSTGRES year extract syntax.\n")
				return POSTGRES_QUERY_YEAR_EXTRACT
			}
		}())

	rows, err := db.Select(colList).
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

	prevAverage := 0.0
	prevCode := ""

	for rows.Next() {
		var code string
		var name string
		var avgPricePSQM float64
		var year int
		increase := 0.0

		rows.Scan(&year, &code, &name, &avgPricePSQM)

		if code == prevCode {
			increase = (avgPricePSQM - prevAverage) / prevAverage
		}
		prevCode = code
		prevAverage = avgPricePSQM

		city2update = append(city2update, map[string]interface{}{"year": year, "code": code, "name": name, "avg_price": avgPricePSQM, "increase": increase})

		log.Debugf("City (%v) year %v avg psqm: %.0f€\n", code, year, avgPricePSQM)
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

const SQLITE_QUERY_YEAR_EXTRACT = "strftime('%Y', transactions.date)"
const POSTGRES_QUERY_YEAR_EXTRACT = "EXTRACT(year FROM transactions.date)"

// aggregateDepartments computes yearly average price per sqm for each department
// and inserts the results into department_yearly_aggs.
//
// Implementation mirrors aggregateCities but joins on departments and uses
// transactions.department_code as the grouping key.
func aggregateDepartments(db *gorm.DB) {

	colList := fmt.Sprintf("%s as year, transactions.department_code as code, MIN(departments.name) as name, AVG(transactions.price_psqm) as avgPricePSQM",
		func() string {
			if db.Dialector.Name() == "sqlite" {
				log.Debugf("Using SQLITE year extract syntax.\n")
				return SQLITE_QUERY_YEAR_EXTRACT
			} else {
				log.Debugf("Using POSTGRES year extract syntax.\n")
				return POSTGRES_QUERY_YEAR_EXTRACT
			}
		}())

	rows, err := db.Select(colList).
		Table("transactions").
		Joins("LEFT JOIN departments on departments.code = transactions.department_code").
		Group("year").Group("transactions.department_code").
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "code"}, Desc: false},
			{Column: clause.Column{Name: "year"}, Desc: false},
		}}).
		Rows()

	if err != nil {
		log.Errorf("aggregateDepartments err: %v\n", err)
		return
	}
	defer rows.Close()

	batchSize := 200
	var dep2update = make([]map[string]interface{}, 0, batchSize)

	prevAverage := 0.0
	prevCode := ""

	for rows.Next() {
		var code string
		var name string
		var avgPricePSQM float64
		var year int
		increase := 0.0

		rows.Scan(&year, &code, &name, &avgPricePSQM)

		if code == prevCode {
			increase = (avgPricePSQM - prevAverage) / prevAverage
		}
		prevCode = code
		prevAverage = avgPricePSQM

		dep2update = append(dep2update, map[string]interface{}{"year": year, "code": code, "name": name, "avg_price": avgPricePSQM, "increase": increase})

		log.Debugf("Dep (%v) year %v avg psqm: %.0f€\n", code, year, avgPricePSQM)
	}

	if len(dep2update) <= 0 {
		log.Infof("Nothing to aggregate for departments.\n")
		return
	}
	bar := pb.Default.Start(len(dep2update))

	for {

		l := len(dep2update)
		b := l - 1
		if l > batchSize {
			b = batchSize
		}

		batch := dep2update[0:b]

		updresult := db.Table("department_yearly_aggs").Create(&batch)

		if updresult.Error != nil {
			log.Errorf("Error aggregateDepartments update: %.0f\n", updresult.Error)
		}

		bar.Add(b)
		if l-1 == b {
			break
		}
		dep2update = dep2update[b+1:]
	}

	bar.Add(int(bar.Total() - bar.Current()))
	bar.Finish()
}

// aggregateRegions computes yearly average price per sqm for regions and writes
// results into region_yearly_aggs.
//
// It joins transactions -> cities -> regions to obtain the region code and name.
func aggregateRegions(db *gorm.DB) {

	colList := fmt.Sprintf("%s as year, cities.code_region as code, MIN(regions.name) as name, AVG(transactions.price_psqm) as avgPricePSQM",
		func() string {
			if db.Dialector.Name() == "sqlite" {
				log.Debugf("Using SQLITE year extract syntax.\n")
				return SQLITE_QUERY_YEAR_EXTRACT
			} else {
				log.Debugf("Using POSTGRES year extract syntax.\n")
				return POSTGRES_QUERY_YEAR_EXTRACT
			}
		}())

	rows, err := db.Select(colList).
		Table("transactions").
		Joins("LEFT JOIN cities on cities.code = transactions.city_code").
		Joins("LEFT JOIN regions on cities.code_region = regions.code").
		Group("year").Group("cities.code_region").
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "code"}, Desc: false},
			{Column: clause.Column{Name: "year"}, Desc: false},
		}}).
		Rows()

	if err != nil {
		log.Errorf("aggregateRegions err: %v\n", err)
		return
	}
	defer rows.Close()

	batchSize := 200
	var region2update = make([]map[string]interface{}, 0, batchSize)

	prevAverage := 0.0
	prevCode := ""

	for rows.Next() {
		var code string
		var name string
		var avgPricePSQM float64
		var year int
		increase := 0.0

		rows.Scan(&year, &code, &name, &avgPricePSQM)

		if code == prevCode {
			increase = (avgPricePSQM - prevAverage) / prevAverage
		}
		prevCode = code
		prevAverage = avgPricePSQM

		region2update = append(region2update, map[string]interface{}{"year": year, "code": code, "name": name, "avg_price": avgPricePSQM, "increase": increase})

		log.Debugf("Dep (%v) year %v avg psqm: %.0f€\n", code, year, avgPricePSQM)
	}

	if len(region2update) <= 0 {
		log.Infof("Nothing to aggregate for region.\n")
		return
	}
	bar := pb.Default.Start(len(region2update))

	for {

		l := len(region2update)
		b := l - 1
		if l > batchSize {
			b = batchSize
		}

		batch := region2update[0:b]

		updresult := db.Table("region_yearly_aggs").Create(&batch)

		if updresult.Error != nil {
			log.Errorf("Error aggregateRegions update: %.0f\n", updresult.Error)
		}

		bar.Add(b)
		if l-1 == b {
			break
		}
		region2update = region2update[b+1:]
	}

	bar.Add(int(bar.Total() - bar.Current()))
	bar.Finish()
}
