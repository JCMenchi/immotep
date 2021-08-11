package loader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	geojson "github.com/paulmach/go.geojson"
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
	Get data from official zipcode list.
	CSV file with following layotu

Code_commune_INSEE;Nom_commune;Code_postal;Ligne_5;Libellé_d_acheminement;coordonnees_gps
02552;NEUVILLETTE;02390;;NEUVILLETTE;49.8554002239,3.46831298648
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
		fmt.Printf("ReadZipcodeMap cannot read Header: %v\n", err)
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
			// add alternate name with -
			zipCodeMap[strings.ReplaceAll(city, " ", "-")] = zip
		}

	}

	return zipCodeMap
}

var badData [][]string = make([][]string, 0, 10)

func LoadRawData(dsn string, filename string, zipCodeMap map[string]int) {
	// open CSV file
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("LoadRawData error: %v\n", err)
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
		fmt.Printf("LoadRawData cannot read Header: %v\n", err)
		return
	}

	db := model.ConnectToDB(dsn)
	doStore := true

	// init counter
	nbData := 0
	nbHouse := 0
	nbHouseWithError := 0
	nbDuplicate := 0

	var previousRow []string

	for {
		row, err := reader.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		nbData++

		if row[TYPE_BIEN_COL] == "Maison" && row[TYPE_VENTE_COL] == "Vente" && row[PRICE_COL] != "" && row[NB_ROOM_COL] != "" {
			ok := checkNotDuplicate(previousRow, row)
			if ok {
				item := createTransaction(row, zipCodeMap)
				if item != nil {
					nbHouse++

					if doStore {
						result := db.Create(item)
						if result.Error != nil {
							fmt.Printf("Error: %v\n", result.Error)
						}
					}

					if nbHouse%50000 == 0 {
						fmt.Printf("%v %v/%v\n", time.Now().Format("15:04:05"), nbHouse, nbData)
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
			fmt.Printf("Error line %v: %v\n", nbData, row)
			fmt.Println(err)
		}
	}
	fmt.Printf("File total rows: %v, data: %v, data with error: %v, duplicate: %v\n\n", nbData, nbHouse, nbHouseWithError, nbDuplicate)

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

		//fmt.Printf("DUPLICATE: %v\n           %v\n\n", row, previousRow)
		return false
	}

	return true
}

func createTransaction(row []string, zipCodeMap map[string]int) *model.Transaction {
	hasError := false

	item := model.Transaction{}

	item.Address = fmt.Sprintf("%v %v %v %v", row[STREET_NUMBER_COL], row[STREET_BIS_COL], row[STREET_TYPE_COL], row[STREET_COL])
	item.City = row[CITY_COL]

	i, err := strconv.Atoi(row[ZIP_COL])
	if err != nil {
		// search in map
		_, ok := zipCodeMap[item.City]
		if ok {
			item.ZipCode = zipCodeMap[item.City]
		} else {
			fmt.Printf("Cannot convert ZIP_CODE %v: %v\n", row, err)
			item.ZipCode = 0
			hasError = true
		}
	} else {
		item.ZipCode = i
		zipCodeMap[item.City] = item.ZipCode
	}

	i, err = strconv.Atoi(row[NB_ROOM_COL])
	if err != nil {
		fmt.Printf("Cannot convert NB_ROOM_COL %v: %v\n", row, err)
		item.NbRoom = 0
		//hasError = true
	} else {
		item.NbRoom = i
	}

	item.DepartmentCode = row[DEP_COL]

	depcode := item.DepartmentCode
	if len(depcode) == 1 {
		depcode = "0" + depcode
	}
	item.CityCode = fmt.Sprintf("%v%v%v", depcode, strings.Repeat("0", 3-len(row[CITY_CODE_COL])), row[CITY_CODE_COL])

	v, err := strconv.ParseFloat(strings.Replace(row[PRICE_COL], ",", ".", 1), 64)
	if err != nil {
		// no interested when no price
		fmt.Printf("No price: (%v)  %v\n", row[PRICE_COL], row)
		hasError = true
	} else {
		item.Price = v
	}

	i, err = strconv.Atoi(row[HOUSE_AREA_COL])
	if err != nil {
		// fmt.Printf("Cannot convert HOUSE_AREA_COL %v: %v\n", row, err)
		hasError = true
	} else {
		item.Area = i
	}

	if row[FULL_AREA_COL] != "" {
		i, err = strconv.Atoi(row[FULL_AREA_COL])
		if err != nil {
			fmt.Printf("Cannot convert FULL_AREA_COL %v: %v\n", row, err)
		} else {
			item.FullArea = i
		}
	}

	item.PricePSQM = item.Price / float64(item.Area)

	item.Cadastre = row[CITY_CODE_COL] + row[SECTION_CADASTRE_COL] + row[CADASTRE_COL]

	t, err := time.Parse("02/01/2006", row[DATE_COL])
	if err != nil {
		fmt.Printf("Cannot convert DATE_COL %v: %v\n", row, err)
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

*/

func LoadRegion(dsn string, filename string) error {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)

	if err != nil {
		fmt.Printf("LoadRegion cannot open %v: %v\n", filename, err)
		return err
	}
	defer jsonFile.Close()
	fmt.Printf("Load region from: %v\n", filename)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var regions []model.Region

	err = json.Unmarshal(byteValue, &regions)
	if err != nil {
		fmt.Printf("LoadRegion cannot decode JSON file %v: %v\n", filename, err)
		return err
	}

	db := model.ConnectToDB(dsn)
	result := db.Create(&regions)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	return nil
}

func LoadDepartment(dsn string, filename string) error {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)

	if err != nil {
		fmt.Printf("LoadDepartment cannot open %v: %v\n", filename, err)
		return err
	}
	defer jsonFile.Close()
	fmt.Printf("Load department from: %v\n", filename)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var departments []model.Department

	err = json.Unmarshal(byteValue, &departments)
	if err != nil {
		fmt.Printf("LoadDepartment cannot decode JSON file %v: %v\n", filename, err)
		return err
	}

	db := model.ConnectToDB(dsn)
	result := db.CreateInBatches(&departments, 30)
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	return nil
}

func LoadCity(dsn string, filename string) error {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)

	if err != nil {
		fmt.Printf("LoadCity cannot open %v: %v\n", filename, err)
		return err
	}
	defer jsonFile.Close()
	fmt.Printf("Load city from: %v\n", filename)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var cities []model.City

	err = json.Unmarshal(byteValue, &cities)
	if err != nil {
		fmt.Printf("LoadCity cannot decode JSON file %v: %v\n", filename, err)
		return err
	}

	db := model.ConnectToDB(dsn)
	cityBatch := make([]model.City, 0, 100)
	total := len(cities)
	for idx, city := range cities {
		if len(city.CodesPostaux) > 0 {
			city.ZipCode, _ = strconv.Atoi(city.CodesPostaux[0])
		}
		err = GetCityContour(&city)
		if err != nil {
			fmt.Printf("LoadCity cannot get contour for %v: %v\n", city.Name, err)
		} else {
			cityBatch = append(cityBatch, city)
			if len(cityBatch) == 100 {
				fmt.Printf("FLush %v/%v cities\n", idx+1, total)
				result := db.Create(&cityBatch)
				if result.Error != nil {
					fmt.Printf("Error: %v\n", result.Error)
				}
				cityBatch = make([]model.City, 0, 100)
			}
		}
	}

	result := db.Create(&cityBatch)
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	return nil
}

type CityInfo struct {
	Name    string           `json:"nom"`
	Code    string           `json:"code"`
	Contour geojson.Geometry `json:"contour"`
}

/**
curl 'https://geo.api.gouv.fr/communes/{codecommune}?fields=code,nom,contour'
*/
var geoAPIBaseURL string = "https://geo.api.gouv.fr/communes/"

func GetCityContour(city *model.City) error {

	query := geoAPIBaseURL + city.Code + "?fields=code,nom,contour"

	response, err := http.Get(query)
	if err != nil {
		fmt.Printf("GetCityContour error in HTTP GET: %v\n", err)
		return err
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("GetCityContour error reading body: %v\n", err)
		return err
	}

	var info CityInfo
	err = json.Unmarshal(responseData, &info)
	if err != nil {
		fmt.Printf("GetCityContour unmarshalling error: %v\n %v\n", err, string(responseData))
		return err
	}

	data, err := json.Marshal(info.Contour)
	if err != nil {
		fmt.Printf("GetCityContour cannot marshall contour: %v\n", err)
		return err
	}
	city.Contour = string(data)

	return nil
}
