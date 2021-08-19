package loader

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/cheggaaa/pb/v3"
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

	bar := pb.Default.Start(int(2 * count))

	// batch size 5000
	batchSize := 5000
	var trans []model.Transaction
	result := query.FindInBatches(&trans, batchSize, func(tx *gorm.DB, batch int) error {

		nbError := 0

		// create CSV data in memory
		b := new(strings.Builder)
		b.WriteString("trid,Address,ZipCode,City\n")

		for _, item := range trans {
			bar.Increment()
			if item.TrId != 0 {
				b.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", item.TrId, item.Address, item.ZipCode, item.City))
			} else {
				log.Debugf("Bad item: %v\n", item)
			}
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
			bar.Increment()
			if len(row) > 6 && row[0] != "" {

				trid, _ := strconv.Atoi(row[0])

				lat, errlat := strconv.ParseFloat(row[4], 64)
				long, errlong := strconv.ParseFloat(row[5], 64)

				if errlat != nil || errlong != nil {
					log.Debugf("GeocodeDB No coord: (%v, %v)  %v\n", row[4], row[5], row)
					nbError++
				} else {
					tr2update = append(tr2update, map[string]interface{}{"tr_id": trid, "lat": lat, "long": long})
				}
			} else {
				log.Debugf("Cannot geocode: %v\n", row)
				nbError++
			}
		}

		if len(tr2update) > 0 {
			// bulk update
			updresult := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tr_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"lat", "long"}),
			}).Table("transactions").Create(&tr2update)

			if updresult.Error != nil {
				log.Errorf("Error GeocodeDB update: %v\n", updresult.Error)
				nbError++
			} else {
				nbprocessed += int(updresult.RowsAffected)
			}
		}

		log.Debugf("GeocodeDB processed batch %v (size %v) elt %v/%v, err: %v\n", batch, len(trans), nbprocessed, count, nbError)
		return nil
	})

	if result.Error != nil {
		log.Errorf("Error GeocodeDB: %v\n", result.Error)
		return
	}

	bar.Add(int(bar.Total() - bar.Current()))
	bar.Finish()
}

var baseURL string = "https://api-adresse.data.gouv.fr/search/csv"

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
	response, err := http.Post(baseURL, w.FormDataContentType(), request_body)
	if err != nil {
		log.Errorf("getGPSCoord error in HTTP POST: %v\n", err)
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
