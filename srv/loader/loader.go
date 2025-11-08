// Package loader contains helpers to import raw property transaction data,
// load geographic configuration (regions, departments, cities) and geocoding
// support used by the immotep application.
//
// Responsibilities:
// - Parse raw transaction CSVs and populate the transactions table.
// - Read and import region/department/city geojson & JSON resources.
// - Provide utilities to resolve zipcode from city codes.
// - Support batching and progress reporting for large datasets.
package loader

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cheggaaa/pb/v3"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"jc.org/immotep/model"
)

var COLUMNS_NAME = []string{
	// 0               1
	"Code service CH", "Reference document", "1 Articles CGI", "2 Articles CGI", "3 Articles CGI",
	"4 Articles CGI", "5 Articles CGI", "No disposition", "Date mutation", "Nature mutation",
	// 10              11
	"Valeur fonciere", "No voie", "B/T/Q", "Type de voie", "Code voie",
	"Voie", "Code postal", "Commune", "Code departement", "Code commune",
	// 20                 21
	"Prefixe de section", "Section", "No plan", "No Volume", "1er lot",
	"Surface Carrez du 1er lot", "2eme lot", "Surface Carrez du 2eme lot", "3eme lot", "Surface Carrez du 3eme lot",
	// 30       31
	"4eme lot", "Surface Carrez du 4eme lot", "5eme lot", "Surface Carrez du 5eme lot", "Nombre de lots",
	"Code type local", "Type local", "Identifiant local", "Surface reelle bati", "Nombre pieces principales",
	// 40             41                         42
	"Nature culture", "Nature culture speciale", "Surface terrain"}

// Column index constants used to extract fields from the raw pipe-separated
// CSV rows produced by the French land registry dataset.
var DATE_COL = 8
var TYPE_VENTE_COL = 9
var PRICE_COL = 10
var STREET_NUMBER_COL = 11
var STREET_BIS_COL = 12
var STREET_TYPE_COL = 13
var STREET_COL = 15
var ZIP_COL = 16
var CITY_COL = 17
var DEP_COL = 18
var CITY_CODE_COL = 19
var SECTION_CADASTRE_COL = 21
var CADASTRE_COL = 22
var TYPE_BIEN_COL = 36
var HOUSE_AREA_COL = 38
var NB_ROOM_COL = 39
var TYPE_CULTURE_COL = 40
var FULL_AREA_COL = 42

/*
ReadZipcodeMap reads a semicolon-separated CSV mapping of official zipcode
records and returns a map mapping city name variants to integer zip codes.

Input file layout expected (semicolon separated):

	Code_commune_INSEE;Nom_de_la_commune;Code_postal;Libellé_d_acheminement;Ligne_5

Returns:
  - map[string]int: mapping of city name and alternative city name (spaces -> '-') to zip code.

Notes:
  - Any file I/O or parse error returns an empty map.
*/
func ReadZipcodeMap(filename string) map[string]int {
	var zipCodeMap map[string]int = make(map[string]int)

	// open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return zipCodeMap
	}
	defer f.Close()

	// parse CSV
	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.LazyQuotes = true

	// skip header
	_, err = reader.Read()
	if err != nil {
		log.Errorf("ReadZipcodeMap cannot read Header: %v\n", err)
		return zipCodeMap
	}

	for {
		row, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		city := row[1]

		zip, err := strconv.Atoi(strings.TrimLeft(row[2], "0"))
		if err == nil {
			zipCodeMap[city] = zip
			// add alternate name with - instead of space
			zipCodeMap[strings.ReplaceAll(city, " ", "-")] = zip
		}

	}

	return zipCodeMap
}

/*
lineCounter returns the number of lines in a file specified by filename.

It reads the file in chunks and counts newline bytes. Returns an error if the
file cannot be opened or an IO error occurs during reading.
*/
func lineCounter(filename string) (int, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("lineCounter error: %v\n", err)
		return -1, err
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

var badData [][]string = make([][]string, 0, 10)

/*
LoadRawData imports raw transaction rows from a pipe-separated CSV file into the
transactions table.

Parameters:
  - dsn: database connection string used to open the DB
  - filename: path to the raw CSV file

Behavior:
  - Counts lines to show progress.
  - Iterates rows, filters for "Maison" sales with price and rooms.
  - Removes obvious duplicates and batches inserts to the DB.
  - Tracks and logs errors and statistics.
*/
func LoadRawData(dsn string, filename string) {

	nbline, _ := lineCounter(filename)

	// open CSV file
	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("LoadRawData error: %v\n", err)
		return
	}
	defer f.Close()

	// parse CSV
	reader := csv.NewReader(f)
	reader.Comma = '|'
	reader.LazyQuotes = true

	// skip Header
	_, err = reader.Read()
	if err != nil {
		log.Errorf("LoadRawData cannot read Header: %v\n", err)
		return
	}

	db := model.ConnectToDB(dsn)

	// init counter
	var nbData int64 = 0
	nbHouse := 0
	nbHouseWithError := 0
	nbDuplicate := 0

	var previousRow []string

	batchSize := 500
	transBatch := make([]*model.Transaction, 0, batchSize)

	bar := pb.Default.Start(nbline)

	for {
		row, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		nbData++

		bar.Increment()

		if row[TYPE_BIEN_COL] == "Maison" && row[TYPE_VENTE_COL] == "Vente" && row[PRICE_COL] != "" && row[NB_ROOM_COL] != "" {
			ok := checkNotDuplicate(previousRow, row)
			if ok {
				item := createTransaction(dsn, row)
				if item != nil {
					nbHouse++
					transBatch = append(transBatch, item)

					if len(transBatch) == batchSize {
						result := db.Create(&transBatch)
						if result.Error != nil {
							log.Errorf("Error: %v\n", result.Error)
						}
						transBatch = make([]*model.Transaction, 0, batchSize)
						//log.Infof("%v %v/%v\n", time.Now().Format("15:04:05"), nbHouse, nbData)
					}
				} else {
					nbHouseWithError++
				}
			} else {
				nbDuplicate++
			}

			previousRow = row
		}

		if err != nil {
			log.Errorf("Error line %v: %v %v\n", nbData, row, err)
		}
	}

	if len(transBatch) > 0 {
		result := db.Create(&transBatch)
		if result.Error != nil {
			log.Errorf("Error: %v\n", result.Error)
		}
		//log.Infof("%v %v/%v\n", time.Now().Format("15:04:05"), nbHouse, nbData)
	}

	bar.Add(int(bar.Total() - bar.Current()))
	bar.Finish()
	log.Infof("File total rows: %v, data: %v, data with error: %v, duplicate: %v\n", nbData, nbHouse, nbHouseWithError, nbDuplicate)

	/*fmt.Printf("Errors:")
	for i, d := range badData {
		fmt.Printf("%v: %v\n", i, d)
	}*/
}

/*
checkNotDuplicate compares the current row with the previous row and returns
false if the rows look like duplicates based on a set of key columns.

Parameters:
  - previousRow: the previously processed CSV row (may be nil/empty)
  - row: current CSV row

Returns:
  - bool: true if the row is not a duplicate and should be processed.
*/
func checkNotDuplicate(previousRow, row []string) bool {
	if len(previousRow) != len(row) {
		return true
	}

	if previousRow[DATE_COL] == row[DATE_COL] &&
		previousRow[PRICE_COL] == row[PRICE_COL] &&
		previousRow[CITY_COL] == row[CITY_COL] &&
		previousRow[SECTION_CADASTRE_COL] == row[SECTION_CADASTRE_COL] &&
		previousRow[CADASTRE_COL] == row[CADASTRE_COL] {

		log.Tracef("DUPLICATE: %v\n           %v\n\n", row, previousRow)
		return false
	}

	return true
}

/*
createTransaction builds a model.Transaction from a parsed CSV row.

Parameters:
  - dsn: database connection string (used for zipcode lookup if ZIP absent)
  - row: slice of string fields representing a CSV row

Returns:
  - *model.Transaction: populated transaction or nil if required fields are invalid.

Behavior:
  - Extracts address parts, numeric conversions for rooms, area, price.
  - Computes price per sqm and constructs cadastre & city code fields.
  - If critical data is missing or conversion fails, the row is appended to badData and nil is returned.
*/
func createTransaction(dsn string, row []string) *model.Transaction {
	hasError := false

	item := model.Transaction{}

	item.Address = fmt.Sprintf("%v %v %v %v", row[STREET_NUMBER_COL], row[STREET_BIS_COL], row[STREET_TYPE_COL], row[STREET_COL])
	item.City = row[CITY_COL]

	i, err := strconv.Atoi(row[NB_ROOM_COL])
	if err != nil {
		log.Errorf("Cannot convert NB_ROOM_COL %v: %v\n", row, err)
		item.NbRoom = 0
		//hasError = true
	} else {
		item.NbRoom = i
	}

	item.DepartmentCode = row[DEP_COL]

	depcode := item.DepartmentCode
	if len(depcode) == 1 {
		depcode = "0" + depcode
	} else if len(depcode) > 2 {
		// only metropolitan dep
		return nil
	}

	item.CityCode = fmt.Sprintf("%v%v%v", depcode, strings.Repeat("0", 3-len(row[CITY_CODE_COL])), row[CITY_CODE_COL])

	if row[ZIP_COL] != "" {
		item.ZipCode, _ = strconv.Atoi(row[ZIP_COL])
	} else {
		item.ZipCode = getZipCodeFromCityCode(dsn, item.CityCode, item.City)
		if item.ZipCode == -1 {
			log.Errorf("No zip: (%v)  %v\n", row[ZIP_COL], row)
		}
	}

	v, err := strconv.ParseFloat(strings.Replace(row[PRICE_COL], ",", ".", 1), 64)
	if err != nil {
		// no interested when no price
		log.Errorf("No price: (%v)  %v\n", row[PRICE_COL], row)
		hasError = true
	} else {
		item.Price = v
	}

	i, err = strconv.Atoi(row[HOUSE_AREA_COL])
	if err != nil {
		log.Debugf("Cannot convert HOUSE_AREA_COL %v: %v\n", row, err)
		hasError = true
	} else {
		item.Area = i
	}

	if row[FULL_AREA_COL] != "" {
		i, err = strconv.Atoi(row[FULL_AREA_COL])
		if err != nil {
			log.Errorf("Cannot convert FULL_AREA_COL %v: %v\n", row, err)
		} else {
			item.FullArea = i
		}
	}

	item.PricePSQM = item.Price / float64(item.Area)

	item.Cadastre = row[CITY_CODE_COL] + row[SECTION_CADASTRE_COL] + row[CADASTRE_COL]

	t, err := time.Parse("02/01/2006", row[DATE_COL])
	if err != nil {
		log.Errorf("Cannot convert DATE_COL %v: %v\n", row, err)
		hasError = true
	}
	item.Date = t

	if hasError {
		badData = append(badData, row)
		return nil
	}

	return &item
}

/*
LoadRegion imports regions from a GeoJSON file into the regions table.

Parameters:
  - dsn: DB connection string
  - filename: path to regions geojson file

Behavior:
  - Skips import if regions table already contains rows.
  - Parses features, extracts 'nom' and 'code' properties and stores the
    whole feature JSON in the contour column.
*/
func LoadRegion(dsn string, filename string) error {
	// check if region already loaded
	db := model.ConnectToDB(dsn)

	var count int64
	db.Table("regions").Count(&count)
	if count > 0 {
		log.Infof("LoadRegion: region already loaded.\n")
		return nil
	}

	// Open our jsonFile
	jsonFile, err := os.Open(filename)

	if err != nil {
		log.Errorf("LoadRegion cannot open %v: %v\n", filename, err)
		return err
	}
	defer jsonFile.Close()
	log.Infof("Load region from: %v...\n", filename)

	byteValue, _ := io.ReadAll(jsonFile)

	var regionsgeo geojson.FeatureCollection

	err = json.Unmarshal(byteValue, &regionsgeo)
	if err != nil {
		log.Errorf("LoadRegion cannot decode JSON file %v: %v\n", filename, err)
		return err
	}

	nb := len(regionsgeo.Features)
	regions := make([]model.Region, 0, nb)
	for _, feature := range regionsgeo.Features {
		var r model.Region
		r.Name, err = feature.PropertyString("nom")
		if err != nil {
			log.Errorf("LoadRegion cannot read property nom: %v\n", err)
		}
		r.Code, err = feature.PropertyString("code")
		if err != nil {
			log.Errorf("LoadRegion cannot read property code: %v\n", err)
		}
		data, err := json.Marshal(feature)
		if err != nil {
			log.Errorf("LoadRegion cannot marshall contour: %v\n", err)
		} else {
			r.Contour = string(data)
		}

		if r.Name != "" {
			regions = append(regions, r)
		}
	}

	result := db.Create(&regions)

	if result.Error != nil {
		log.Errorf("LoadRegion Error: %v\n", result.Error)
	}

	log.Infof("...region loaded.\n")

	return nil
}

/*
LoadDepartment imports department polygons from a GeoJSON file into the
departments table.

Parameters:
  - dsn: DB connection string
  - filename: path to departments geojson file

Behavior:
  - Skips import if departments table already contains rows.
  - Only imports metropolitan departments (code length < 3).
  - Stores the feature JSON in the contour column and persists rows in batches.
*/
func LoadDepartment(dsn string, filename string) error {
	// check if department already loaded
	db := model.ConnectToDB(dsn)
	var count int64
	db.Table("departments").Count(&count)
	if count > 0 {
		log.Infof("LoadDepartment: department already loaded.\n")
		return nil
	}

	// Open our jsonFile
	jsonFile, err := os.Open(filename)

	if err != nil {
		log.Errorf("LoadDepartment cannot open %v: %v\n", filename, err)
		return err
	}
	defer jsonFile.Close()
	log.Infof("Load department from: %v...\n", filename)

	byteValue, _ := io.ReadAll(jsonFile)

	var departmentsgeo geojson.FeatureCollection
	err = json.Unmarshal(byteValue, &departmentsgeo)
	if err != nil {
		log.Errorf("LoadDepartment cannot decode JSON file %v: %v\n", filename, err)
		return err
	}

	nb := len(departmentsgeo.Features)
	departments := make([]model.Department, 0, nb)
	for _, feature := range departmentsgeo.Features {
		var d model.Department
		d.Name, err = feature.PropertyString("nom")
		if err != nil {
			log.Errorf("LoadDepartment cannot read property nom: %v\n", err)
		}
		d.Code, err = feature.PropertyString("code")
		if err != nil {
			log.Errorf("LoadDepartment cannot read property code: %v\n", err)
		}

		// only metropolitan dep
		if err == nil && d.Name != "" && len(d.Code) < 3 {
			data, err := json.Marshal(feature)
			if err != nil {
				log.Errorf("LoadDepartment cannot marshall contour: %v\n", err)
			} else {
				d.Contour = string(data)
			}

			departments = append(departments, d)
		}
	}

	result := db.CreateInBatches(&departments, 10)
	if result.Error != nil {
		log.Errorf("Error: %v\n", result.Error)
	}

	log.Infof("...department loaded.\n")

	return nil
}

/*
LoadCity imports city metadata from a JSON file and associates GeoJSON contours
from a companion geojson file.

Parameters:
  - dsn: DB connection string
  - filename: path to the cities JSON (list of City structs)
  - geofilename: path to the cities GeoJSON (feature collection with contours)

Behavior:
  - Skips import if cities table already contains rows.
  - Normalizes city names (uppercase, strip accents) and populates zipcode.
  - Persists city batches and updates the PostGIS geometry column from stored contour JSON.
*/
func LoadCity(dsn string, filename string, geofilename string) error {
	// check if city already loaded
	db := model.ConnectToDB(dsn)
	var count int64
	db.Table("cities").Count(&count)
	if count > 0 {
		log.Infof("LoadCity: city already loaded.\n")
		return nil
	}

	// Open our jsonFile
	jsonFile, err := os.Open(filename)

	if err != nil {
		log.Errorf("LoadCity cannot open %v: %v\n", filename, err)
		return err
	}
	defer jsonFile.Close()

	// Open our geojsonFile
	geojsonFile, err := os.Open(geofilename)

	if err != nil {
		log.Errorf("LoadCity cannot open geo json %v: %v\n", geofilename, err)
		return err
	}
	defer geojsonFile.Close()

	// load geojson data
	byteValue, _ := io.ReadAll(geojsonFile)

	var communesgeo geojson.FeatureCollection
	err = json.Unmarshal(byteValue, &communesgeo)
	if err != nil {
		log.Errorf("LoadCity cannot decode GEOJSON file %v: %v\n", geofilename, err)
		return err
	}

	// load data
	log.Infof("Load city from: %v...\n", filename)

	byteValue, _ = io.ReadAll(jsonFile)

	var cities []model.City

	err = json.Unmarshal(byteValue, &cities)
	if err != nil {
		log.Errorf("LoadCity cannot decode JSON file %v: %v\n", filename, err)
		return err
	}

	total := len(cities)
	bar := pb.Default.Start(total)

	batchSize := 200
	cityBatch := make([]model.City, 0, batchSize)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	for _, city := range cities {
		bar.Increment()
		if len(city.CodesPostaux) > 0 {
			city.ZipCode, _ = strconv.Atoi(city.CodesPostaux[0])
		}

		city.NameUpper, _, _ = transform.String(t, strings.ToUpper(city.Name))

		if city.Code != "" && city.CodeDepartment != "" && len(city.CodeDepartment) < 3 { // only metropolitan dep
			city.Contour, err = getCityContour(city.Code, communesgeo)
			if err != nil {
				log.Errorf("LoadCity cannot get contour for %v: %v\n", city.Name, err)
			} else {
				cityBatch = append(cityBatch, city)
				if len(cityBatch) == batchSize {
					result := db.Omit("Geom").Create(&cityBatch)
					if result.Error != nil {
						log.Errorf("Error: %v\n", result.Error)
					}
					cityBatch = make([]model.City, 0, batchSize)
				}
			}
		}
	}

	if len(cityBatch) > 0 {
		result := db.Omit("Geom").CreateInBatches(&cityBatch, 50)
		if result.Error != nil {
			log.Errorf("Error: %v\n", result.Error)
		}
	}

	bar.Finish()

	// Update Geometry
	log.Infof("Update city postgis column...\n")
	text := "WITH csubquery AS (SELECT code, ST_GeomFromGeoJSON(contour::json->>'geometry') as imp FROM cities) UPDATE cities SET geom=csubquery.imp FROM csubquery WHERE cities.code=csubquery.code;"
	res := db.Exec(text)
	if res.Error != nil {
		log.Errorf("Update Geometry error: %v\n", res.Error)
		return nil
	}

	log.Infof("...city loaded.\n")

	return nil
}

// CityInfo is a lightweight structure used when extracting city contour data.
type CityInfo struct {
	Name    string           `json:"nom"`
	Code    string           `json:"code"`
	Contour geojson.Geometry `json:"contour"`
}

/*
getCityContour searches the provided geojson FeatureCollection for a feature
matching the given cityCode and returns the serialized feature JSON (contour).

Parameters:
  - cityCode: INSEE city code to match
  - communesgeo: geojson.FeatureCollection containing city features

Returns:
  - string: JSON representation of the matched feature
  - error: if no matching contour is found

Behavior:
  - When a match is found, the function removes the matched feature from the
    input slice to reduce subsequent search cost.
*/
func getCityContour(cityCode string, communesgeo geojson.FeatureCollection) (string, error) {

	for idx, feature := range communesgeo.Features {
		code, err := feature.PropertyString("code")
		if err != nil {
			log.Errorf("GetCityContour cannot read property code: %v\n", err)
		}

		// only metropolitan dep
		if err == nil && code == cityCode {
			data, errm := json.Marshal(feature)
			if errm != nil {
				log.Errorf("GetCityContour cannot read property code: %v\n", err)
			} else {
				// delete element to reduce array
				communesgeo.Features[idx] = communesgeo.Features[len(communesgeo.Features)-1] // Copy last element to index i.
				communesgeo.Features = communesgeo.Features[:len(communesgeo.Features)-1]     // Truncate slice.
				return string(data), nil
			}
		}
	}

	return "", errors.New("no contour found")
}

/*
getZipCodeFromCityCode attempts to resolve a zipcode given a city INSEE code
and an optional city name.

Parameters:
  - dsn: DB connection string
  - codeCity: INSEE city code (used to query cities table)
  - uppername: city name (fallback lookup by normalized name)

Returns:
  - int: zipcode if found, otherwise -1
*/
func getZipCodeFromCityCode(dsn string, codeCity string, uppername string) int {
	db := model.ConnectToDB(dsn)
	if db == nil {
		log.Errorf("getZipCodeFromCityCode err: cannot connect to DB: %v\n", dsn)
		return -1
	}

	// build query
	rows, err := db.Select("zip_code").
		Where("code = ?", codeCity).
		Table("cities").
		Rows()

	if err != nil {
		log.Errorf("getZipCodeFromCityCode err: %v\n", err)
		return -1
	}
	defer rows.Close()

	for rows.Next() {
		var zip int

		rows.Scan(&zip)

		return zip
	}

	rows2, err := db.Select("zip_code").
		Where("name_upper = ?", uppername).
		Table("cities").
		Rows()

	if err != nil {
		log.Errorf("getZipCodeFromCityCode err: %v\n", err)
		return -1
	}
	defer rows2.Close()

	for rows2.Next() {
		var zip int

		rows2.Scan(&zip)

		return zip
	}

	log.Errorf("getZipCodeFromCityCode no zip for: %v\n", codeCity)
	return -1
}
