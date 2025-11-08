package model

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"
)

// helper to open a temporary sqlite DB and return db + dsn
func openTestDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.db")
	dsn := fmt.Sprintf("file:%s", path)
	db := ConnectToDB(dsn)
	if db == nil {
		t.Fatalf("ConnectToDB returned nil for dsn %s", dsn)
	}
	// ensure yearly agg tables exist for some tests
	db.AutoMigrate(&CityYearlyAgg{}, &DepartmentYearlyAgg{}, &RegionYearlyAgg{})
	return db, dsn
}

// seed a minimal dataset: region, department, city and two transactions in two years
func seedMinimal(db *gorm.DB, t *testing.T) {
	t.Helper()

	feat := `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]},"properties":{}}`

	r := Region{Code: "R1", Name: "Reg1", Contour: feat}
	d := Department{Code: "D1", Name: "Dep1", Contour: feat}
	c := City{Code: "C1", Name: "City1", NameUpper: "CITY1", ZipCode: 75000, Population: 1000, Contour: feat, CodeDepartment: "D1", CodeRegion: "R1"}

	if err := db.Create(&r).Error; err != nil {
		t.Fatalf("create region: %v", err)
	}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("create department: %v", err)
	}
	if err := db.Omit("Geom").Create(&c).Error; err != nil {
		t.Fatalf("create city: %v", err)
	}

	// two transactions: 2020 and 2021
	tr1 := Transaction{
		Date:           time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		Address:        "1 rue A",
		ZipCode:        75000,
		City:           "City1",
		CityCode:       "C1",
		DepartmentCode: "D1",
		Price:          100000,
		Area:           50,
		PricePSQM:      100000.0 / 50.0, // 2000
		Lat:            0.5,
		Long:           0.5,
	}
	tr2 := Transaction{
		Date:           time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
		Address:        "2 rue B",
		ZipCode:        75000,
		City:           "City1",
		CityCode:       "C1",
		DepartmentCode: "D1",
		Price:          110000,
		Area:           50,
		PricePSQM:      110000.0 / 50.0, // 2200
		Lat:            0.6,
		Long:           0.6,
	}

	if err := db.Create(&tr1).Error; err != nil {
		t.Fatalf("create tr1: %v", err)
	}
	if err := db.Create(&tr2).Error; err != nil {
		t.Fatalf("create tr2: %v", err)
	}
}

func TestConnectToDB_SQLite(t *testing.T) {
	db, _ := openTestDB(t)
	if db == nil {
		t.Fatal("expected db not nil")
	}
	// check tables exist
	if !db.Migrator().HasTable(&Transaction{}) {
		t.Fatal("transactions table missing")
	}
	if !db.Migrator().HasTable(&City{}) {
		t.Fatal("cities table missing")
	}
	if !db.Migrator().HasTable(&Region{}) {
		t.Fatal("regions table missing")
	}
	if !db.Migrator().HasTable(&Department{}) {
		t.Fatal("departments table missing")
	}
}

func TestComputeRegionsDepartmentsCities(t *testing.T) {
	db, _ := openTestDB(t)
	seedMinimal(db, t)

	// run computations
	ComputeRegions(db)
	ComputeDepartments(db)
	ComputeCities(db)

	// verify region avg price
	var reg Region
	if err := db.First(&reg, "code = ?", "R1").Error; err != nil {
		t.Fatalf("read region: %v", err)
	}
	// average of 2000 and 2200 = 2100
	if reg.AvgPrice != 2100.0 {
		t.Fatalf("region avg expect 2100 got %v", reg.AvgPrice)
	}

	// verify department avg price
	var dep Department
	if err := db.First(&dep, "code = ?", "D1").Error; err != nil {
		t.Fatalf("read department: %v", err)
	}
	if dep.AvgPrice != 2100.0 {
		t.Fatalf("department avg expect 2100 got %v", dep.AvgPrice)
	}

	// verify city avg price (upsert)
	var c City
	if err := db.First(&c, "code = ?", "C1").Error; err != nil {
		t.Fatalf("read city: %v", err)
	}
	if c.AvgPrice != 2100.0 {
		t.Fatalf("city avg expect 2100 got %v", c.AvgPrice)
	}
}

func TestComputeEmptyDB(t *testing.T) {

	db, dsn := openTestDB(t)

	// run computations
	ComputeStat(dsn)

	// verify region
	var nbRegion int64
	if err := db.Table("regions").Count(&nbRegion).Error; err != nil {
		t.Fatalf("read region: %v", err)
	}

	if nbRegion != 0 {
		t.Fatalf("region not empty. Count %v", nbRegion)
	}

	// verify department
	var nbDep int64
	if err := db.Table("departments").Count(&nbDep).Error; err != nil {
		t.Fatalf("read department: %v", err)
	}
	if nbDep != 0 {
		t.Fatalf("department not empty. Count %v", nbDep)
	}

	// verify city avg price (upsert)
	var nbCity int64
	if err := db.Table("cities").Count(&nbCity).Error; err != nil {
		t.Fatalf("read city: %v", err)
	}
	if nbCity != 0 {
		t.Fatalf("city not empty. Count %v", nbCity)
	}
}

func TestGetPOIAndGetPOIFromBounds(t *testing.T) {
	db, _ := openTestDB(t)
	seedMinimal(db, t)

	// Test GetPOI: limit and zip
	pois := GetPOI(db, 10, 75000, "")
	if len(pois) == 0 {
		t.Fatalf("expected pois > 0")
	}

	// Test GetPOIFromBounds: coords covering (0.4..0.7)
	info := GetPOIFromBounds(db, 1.0, 1.0, 0.0, 0.0, 10, "", 0)
	if info == nil {
		t.Fatalf("GetPOIFromBounds returned nil")
	}
	if info.AvgPriceSQM == 0 {
		t.Fatalf("expected non-zero AvgPriceSQM")
	}
}

func TestGetCityRegionDepartmentDetailsAndStats(t *testing.T) {
	db, _ := openTestDB(t)
	seedMinimal(db, t)

	// prepare yearly aggs to test getCityStat/getRegionStat/getDepartmentStat
	db.Create(&CityYearlyAgg{Code: "C1", Year: 2020, Name: "City1", AvgPrice: 2000, Increase: 0.0})
	db.Create(&CityYearlyAgg{Code: "C1", Year: 2021, Name: "City1", AvgPrice: 2200, Increase: 0.1})

	db.Create(&DepartmentYearlyAgg{Code: "D1", Year: 2020, Name: "Dep1", AvgPrice: 2000, Increase: 0.0})
	db.Create(&RegionYearlyAgg{Code: "R1", Year: 2021, Name: "Reg1", AvgPrice: 2200, Increase: 0.1})

	// City details
	cities := GetCityDetails(db, "")
	if len(cities) == 0 {
		t.Fatalf("GetCityDetails returned none")
	}
	// getCityStat (unexported) should return formatted map
	cstat := getCityStat(db, "C1")
	if len(cstat) == 0 {
		t.Fatalf("getCityStat empty")
	}
	if _, ok := cstat[2021]; !ok {
		t.Fatalf("getCityStat missing 2021")
	}

	// Region details & stat
	regs := GetRegionDetails(db)
	if len(regs) == 0 {
		t.Fatalf("GetRegionDetails returned none")
	}
	rstat := getRegionStat(db, "R1")
	if len(rstat) == 0 {
		t.Fatalf("getRegionStat empty")
	}

	// Department details & stat
	deps := GetDepartmentDetails(db)
	if len(deps) == 0 {
		t.Fatalf("GetDepartmentDetails returned none")
	}
	dstat := getDepartmentStat(db, "D1")
	if len(dstat) == 0 {
		t.Fatalf("getDepartmentStat empty")
	}

	// quick sanity checks
	if cities[0].Code != "C1" {
		t.Fatalf("unexpected city code %v", cities[0].Code)
	}
	if regs[0].Code != "R1" {
		t.Fatalf("unexpected region code %v", regs[0].Code)
	}
	if deps[0].Code != "D1" {
		t.Fatalf("unexpected department code %v", deps[0].Code)
	}
}

func TestAggregateDataCreatesYearlyAggs(t *testing.T) {
	_, dsn := openTestDB(t)
	// AggregateData will call ConnectToDB and AutoMigrate yearly agg tables
	// seed via a separate connection
	db := ConnectToDB(dsn)
	if db == nil {
		t.Fatal("connect2 returned nil")
	}
	seedMinimal(db, t)

	AggregateData(dsn)

	// after aggregation, check city_yearly_aggs contains entries for C1
	var stat []CityYearlyAgg
	if err := db.Where("code = ?", "C1").Find(&stat).Error; err != nil {
		t.Fatalf("query city_yearly_aggs: %v", err)
	}

	if len(stat) != 0 {
		t.Fatalf("expected city_yearly_aggs rows for C1")
	}
}
