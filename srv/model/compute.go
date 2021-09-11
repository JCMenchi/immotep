package model

import (
	"github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ComputeRegions(db *gorm.DB) {
	rows, err := db.Select("regions.name as name, regions.code as code, AVG(transactions.price_psqm) as avg_price_psqm").
		Joins("LEFT JOIN departments ON departments.code = transactions.department_code").
		Joins("LEFT JOIN regions ON regions.code = departments.code_region").
		Table("transactions").
		Group("regions.code").
		Rows()

	if err != nil {
		log.Errorf("ComputeRegions err: %v\n", err)
		return
	}
	defer rows.Close()

	var reginfos []RegionInfo = make([]RegionInfo, 0, 100)

	for rows.Next() {
		var code, name string
		var avg_price_psqm float64

		rows.Scan(&name, &code, &avg_price_psqm)
		reginfos = append(reginfos, RegionInfo{Code: code, AvgPriceSQM: avg_price_psqm})
		log.Debugf("Region avg psqm %v: %.0f€\n", name, avg_price_psqm)
	}

	if len(reginfos) <= 0 {
		log.Infof("Nothing to compute for regions.\n")
		return
	}

	for _, info := range reginfos {

		updresult := db.Model(Region{}).Where("code = ?", info.Code).Updates(map[string]interface{}{"avg_price": info.AvgPriceSQM})
		if updresult.Error != nil {
			log.Errorf("Error ComputeRegions update: %v\n", updresult.Error)
		}
	}
}

func ComputeDepartments(db *gorm.DB) {

	rows, err := db.Select("departments.code as code, AVG(transactions.price_psqm) as avg_price_psqm").
		Joins("LEFT JOIN departments ON departments.code = transactions.department_code").
		Table("transactions").
		Group("departments.code").
		Rows()

	if err != nil {
		log.Errorf("ComputeDepartments err: %v\n", err)
		return
	}
	defer rows.Close()

	var depinfos []DepartmentInfo = make([]DepartmentInfo, 0, 100)

	for rows.Next() {
		var code string
		var avg_price_psqm float64

		rows.Scan(&code, &avg_price_psqm)
		depinfos = append(depinfos, DepartmentInfo{Code: code, AvgPriceSQM: avg_price_psqm})
		log.Debugf("Department avg psqm %v: %.0f€\n", code, avg_price_psqm)
	}

	if len(depinfos) <= 0 {
		log.Infof("Nothing to compute for departments.\n")
		return
	}

	for _, info := range depinfos {

		updresult := db.Model(Department{}).Where("code = ?", info.Code).Updates(map[string]interface{}{"avg_price": info.AvgPriceSQM})
		if updresult.Error != nil {
			log.Errorf("Error ComputeDepartments update: %v\n", updresult.Error)
		}
	}
}

func ComputeCities(db *gorm.DB) {

	rows, err := db.Select("transactions.city_code as code, AVG(transactions.price_psqm) as avg_price_psqm").
		Table("transactions").
		Group("city_code").
		Rows()

	if err != nil {
		log.Errorf("ComputeCities err: %v\n", err)
		return
	}
	defer rows.Close()

	batchSize := 1000
	var city2update = make([]map[string]interface{}, 0, batchSize)

	for rows.Next() {
		var code string
		var avg_price_psqm float64

		rows.Scan(&code, &avg_price_psqm)

		city2update = append(city2update, map[string]interface{}{"code": code, "avg_price": avg_price_psqm})

		log.Debugf("City (%v) avg psqm: %.0f€\n", code, avg_price_psqm)
	}

	if len(city2update) <= 0 {
		log.Infof("Nothing to compute for cities.\n")
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

		updresult := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoUpdates: clause.AssignmentColumns([]string{"avg_price"}),
		}).Table("cities").Create(&batch)

		if updresult.Error != nil {
			log.Errorf("Error ComputeCities update: %.0f\n", updresult.Error)
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

func ComputeStat(dsn string) {
	db := ConnectToDB(dsn)
	log.Infof("Compute Stat for Regions...\n")
	ComputeRegions(db)
	log.Infof("Compute Stat for Departments...\n")
	ComputeDepartments(db)
	log.Infof("Compute Stat for Cities...\n")
	ComputeCities(db)
	log.Infof("All Stat computed.\n")
}
