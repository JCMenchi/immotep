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
ReadZipcodeMap gets data from official zipcode list.
Returns a map of zipcode by city name as integer.

Data store in CSV file with following layout:

#Code_commune_INSEE;Nom_de_la_commune;Code_postal;Libellé_d_acheminement;Ligne_5
01001;L ABERGEMENT CLEMENCIAT;01400;L ABERGEMENT CLEMENCIAT;
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
TODO: merge duplicate surface in main item
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

SELECT city, cadastre, date, MAX(price), MAX(price)/MAX(area) as ppsqm, COUNT(date) as Total
FROM transactions
GROUP BY city, cadastre, date
HAVING Total > 2
ORDER BY ppsqm;

SELECT zip_code, AVG(price_psqm) as ppsqm
FROM transactions
WHERE department_code = 29
GROUP BY zip_code
ORDER BY zip_code;

HAVING Total > 2
ORDER BY ppsqm;


SELECT t1.department_code, perdep, errperdep, errperdep/perdep
FROM
(select department_code,count(*) as perdep
from transactions group by department_code) t1
LEFT JOIN
(select department_code,count(*) as errperdep
from transactions where nb_room = 0 group by department_code) t2
ON (t1.department_code = t2.department_code);

Feature property
			"properties": {
                "code": "75",
                "nom": "Nouvelle-Aquitaine"
            }
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
				"properties": {
	                "reg_code": "84",
	                "dep_is_ctu": "Non",
	                "dep_status": "urbain",
	                "dep_name_upper": "LOIRE",
	                "reg_name": "Auvergne-Rh\u00f4ne-Alpes",
	                "geo_point_2d": [
	                    45.7279998676,
	                    4.16481278582
	                ],
	                "dep_current_code": "42",
	                "dep_name_lower": "loire",
	                "dep_code": "42",
	                "dep_type": "d\u00e9partement",
	                "year": "2021",
	                "dep_area_code": "FXX",
	                "dep_siren_code": "224200014",
	                "dep_name": "Loire",
	                "viewport": "{\"type\": \"Polygon\", \"coordinates\": [[[3.688420154, 45.231033918], [3.688420154, 46.276565491], [4.760377824, 46.276565491], [4.760377824, 45.231033918], [3.688420154, 45.231033918]]]}"
	            }
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
					result := db.Create(&cityBatch)
					if result.Error != nil {
						log.Errorf("Error: %v\n", result.Error)
					}
					cityBatch = make([]model.City, 0, batchSize)
				}
			}
		}
	}

	if len(cityBatch) > 0 {
		result := db.CreateInBatches(&cityBatch, 50)
		if result.Error != nil {
			log.Errorf("Error: %v\n", result.Error)
		}
	}

	bar.Finish()
	log.Infof("...city loaded.\n")

	return nil
}

type CityInfo struct {
	Name    string           `json:"nom"`
	Code    string           `json:"code"`
	Contour geojson.Geometry `json:"contour"`
}

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

func getZipCodeFromCityCode(dsn string, codeCity string, name string) int {
	db := model.ConnectToDB(dsn)

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
		Where("name_upper = ?", name).
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
