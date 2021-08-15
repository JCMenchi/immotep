package loader

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"jc.org/immotep/model"
)

/*
map[
    attribution:BAN
    licence:ETALAB-2.0
    limit:5
    query:5156   PELUS SAINT-JEAN-SUR-REYSSOUZE
    type:FeatureCollection
    version:draft

    features:[
        map[
            geometry:map[coordinates:[5.060514 46.387785]
                            type:Point
                            ]
            properties:map[city:Saint-Jean-sur-Reyssouze
                            citycode:01364
                            context:01, Ain, Auvergne-Rhône-Alpes
                            id:01364_0420
                            importance:0.471
                            label:Route des Pelus 01560 Saint-Jean-sur-Reyssouze
                            name:Route des Pelus
                            postcode:01560
                            score:0.6136427061310782
                            type:street
                            x:858327.53
							y:6.58960389e+06
                            ]
            type:Feature
            ]
        ]
]

curl -X POST -F data=@path/to/file.csv -F columns=voie -F columns=ville https://api-adresse.data.gouv.fr/search/csv/

*/

type ShapeGeometry struct {
	Coordinates []float64 `json:"coordinates"`
	FeatureType string    `json:"type"`
}

type GeoFeature struct {
	Geometry    ShapeGeometry `json:"geometry"`
	Properties  interface{}   `json:"properties"`
	FeatureType string        `json:"type"`
}

type GeoCodeInfo struct {
	Attribution string       `json:"attribution"`
	Licence     string       `json:"licence"`
	Limit       int          `json:"limit"`
	Query       string       `json:"query"`
	FeatureType string       `json:"type"`
	Version     string       `json:"version"`
	Features    []GeoFeature `json:"features"`
}

var baseURL string = "https://api-adresse.data.gouv.fr/search/?q="

func GetLatLong(item *model.Transaction) *model.Transaction {

	city := strings.Join(strings.Fields(item.City), "-")
	city = strings.ReplaceAll(city, "-", "_")

	addr := strings.Join(strings.Fields(item.Address), "+")
	addr = addr + "+" + city

	response, err := http.Get(baseURL + addr)
	if err != nil {
		fmt.Printf("GetLatLong error in HTTP GET: %v\n", err)
		return nil
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("GetLatLong error reading body: %v\n", err)
		return nil
	}

	var info GeoCodeInfo
	err = json.Unmarshal(responseData, &info)
	if err != nil {
		fmt.Printf("GetLatLong unmarshalling error: %v\n %v\n", err, string(responseData))
		return nil
	}

	if len(info.Features) == 0 {
		fmt.Printf("GetLatLong no feature: %v\n", info)
		// FIXME: mark with 0
		item.Long = 0.0
		item.Lat = 0.0
		return item
	}

	geom := info.Features[0].Geometry

	fmt.Printf("Geometry (%v) %v %v\n", geom, item.Address, item.City)

	item.Long = geom.Coordinates[0]
	item.Lat = geom.Coordinates[1]

	return item
}

func GeocodeDB(dsn string, depcode string) {

	db := model.ConnectToDB(dsn)

	// build query
	query := db.Where("lat = 0")
	if depcode != "" {
		query = db.Where("lat = 0 AND department_code = ?", depcode)
	}

	// nb elt
	var count int64
	query.Table("transactions").Count(&count)
	nbprocessed := 0

	// batch size 2000
	batchSize := 5000
	var trans []model.Transaction
	result := query.FindInBatches(&trans, batchSize, func(tx *gorm.DB, batch int) error {

		nbError := 0

		// create CSV data in memory
		b := new(strings.Builder)
		b.WriteString("trid,Address,ZipCode,City\n")

		for _, item := range trans {
			b.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", item.TrId, item.Address, item.ZipCode, item.City))
		}

		// fetch data
		csvread, err := getGPSCoord(b.String())
		if err != nil {
			return nil
		}

		// skip header
		_, err = csvread.Read()
		if err != nil {
			return nil
		}

		// parse result
		var tr2update = make([]map[string]interface{}, 0, batchSize)

		for {
			row, err := csvread.Read()
			// Stop at EOF.
			if err == io.EOF {
				break
			}

			if len(row) > 6 {

				trid, _ := strconv.Atoi(row[0])

				lat, err := strconv.ParseFloat(row[4], 64)
				if err != nil {
					fmt.Printf("GeocodeDB No lat: (%v)  %v\n", row[4], row)
					nbError++
				}
				long, err := strconv.ParseFloat(row[5], 64)
				if err != nil {
					fmt.Printf("GeocodeDB No long: (%v)  %v\n", row[5], row)
				}

				tr2update = append(tr2update, map[string]interface{}{"tr_id": trid, "lat": lat, "long": long})

			} else {
				fmt.Printf("Cannot geocode: %v\n", row)
				nbError++
			}
		}

		// bulk update
		updresult := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tr_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"lat", "long"}),
		}).Table("transactions").Create(&tr2update)

		if updresult.Error != nil {
			fmt.Printf("Error GeocodeDB update: %v\n", updresult.Error)
			nbError++
		} else {
			nbprocessed += int(updresult.RowsAffected)
		}

		fmt.Printf("[%v] GeocodeDB processed batch %v (size %v) elt %v/%v, err: %v\n", time.Now().Format("15:04:05"), batch, len(trans), nbprocessed, count, nbError)
		return nil
	})

	if result.Error != nil {
		fmt.Printf("Error GeocodeDB: %v\n", result.Error)
		return
	}
}

func getGPSCoord(csvdata string) (*csv.Reader, error) {

	// create multipart body message
	request_body := &bytes.Buffer{}
	w := multipart.NewWriter(request_body)
	// specify columns to use
	w.WriteField("columns", "Address")
	w.WriteField("columns", "City")
	w.WriteField("columns", "ZipCode")
	// add data file
	datapart, _ := w.CreateFormFile("data", "address.csv")
	// copy csv data it to its part
	io.Copy(datapart, strings.NewReader(csvdata))
	// close multipart body
	w.Close()

	// send request
	response, err := http.Post("https://api-adresse.data.gouv.fr/search/csv/", w.FormDataContentType(), request_body)
	if err != nil {
		fmt.Printf("getGPSCoord error in HTTP POST: %v\n", err)
		return nil, err
	}
	defer response.Body.Close()

	response_body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// parse CSV
	reader := csv.NewReader(strings.NewReader(string(response_body)))
	reader.LazyQuotes = true

	return reader, nil
}
