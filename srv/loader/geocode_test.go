package loader

import (
	"bytes"
	"encoding/csv"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"jc.org/immotep/model"
)

// TestGeocodeDBUpdatesCoordinates verifies that GeocodeDB posts CSV to the geocoding
// endpoint, parses the response and updates the transactions table with returned
// lat/long and normalized address fields.
func TestGeocodeDBUpdatesCoordinates(t *testing.T) {
	// start test server that simulates geocoding CSV endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify Content-Type is multipart/form-data
		ct := r.Header.Get("Content-Type")
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			t.Fatalf("invalid content type: %v", err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])

		var dataBuf bytes.Buffer
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("reading multipart: %v", err)
			}
			if part.FormName() == "data" {
				io.Copy(&dataBuf, part)
				part.Close()
			}
		}

		// parse incoming CSV to extract trids
		cr := csv.NewReader(strings.NewReader(dataBuf.String()))
		cr.FieldsPerRecord = -1
		// skip header
		_, err = cr.Read()
		if err != nil {
			// empty input -> return header only
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hdr\n"))
			return
		}

		var trids []int
		for {
			row, err := cr.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("parse incoming csv: %v", err)
			}
			if len(row) > 0 && row[0] != "" {
				id, _ := strconv.Atoi(row[0])
				trids = append(trids, id)
			}
		}

		// build response CSV with required column indexes up to STATUS_INDEX (20)
		w.Header().Set("Content-Type", "text/csv")
		cw := csv.NewWriter(w)

		// header
		_ = cw.Write([]string{"id", "c1", "c2", "LONG", "LAT", "c5", "c6", "ADDRESS", "c8", "c9", "c10", "c11", "c12", "ZIPCODE", "CITYNAME", "c15", "CITYCODE", "c17", "c18", "c19", "STATUS"})
		// rows
		for i, id := range trids {
			cols := make([]string, STATUS_INDEX+1)
			cols[0] = strconv.Itoa(id)
			// set long and lat to deterministic values
			long := float64(i) * (10.0 + float64(id) + float64(i)/100.0)
			lat := float64(i) * (50.0 + float64(id) + float64(i)/100.0)
			cols[LONG_INDEX] = strconv.FormatFloat(long, 'f', 6, 64)
			cols[LAT_INDEX] = strconv.FormatFloat(lat, 'f', 6, 64)
			cols[ADDRESS_INDEX] = "Normalized Address " + strconv.Itoa(id)
			cols[ZIPCODE_INDEX] = strconv.Itoa(10000 + id)
			cols[CITYNAME_INDEX] = "City" + strconv.Itoa(id)
			cols[CITYCODE_INDEX] = "C" + strconv.Itoa(id)
			cols[STATUS_INDEX] = "OK"
			_ = cw.Write(cols)
		}
		cw.Flush()
	}))
	defer ts.Close()

	// point the package baseURL to our test server
	origBase := geocodeBaseURL
	geocodeBaseURL = ts.URL
	defer func() { geocodeBaseURL = origBase }()

	// create temporary sqlite db file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_geocode.db")
	dsn := "file:" + dbPath

	// connect and prepare schema
	db := model.ConnectToDB(dsn)

	// insert sample transactions with lat=0 to be geocoded
	sample := []model.Transaction{
		{TrId: 1, Address: "addr1", City: "old1", ZipCode: 10000, Lat: 0, Long: 0},
		{TrId: 2, Address: "addr2", City: "old2", ZipCode: 10001, Lat: 0, Long: 0},
		{TrId: 3, Address: "addr3", City: "old3", ZipCode: 10002, Lat: 0, Long: 0},
	}
	if err := db.Create(&sample).Error; err != nil {
		t.Fatalf("insert sample failed: %v", err)
	}

	// run GeocodeDB against the same DSN
	GeocodeDB(dsn, true, "22")
	GeocodeDB(dsn, true, "")
	GeocodeDB(dsn, false, "22")

	// reconnect and verify updates
	db2 := model.ConnectToDB(dsn)
	for _, id := range []int{2, 3} {
		var row map[string]interface{}
		if err := db2.Table("transactions").Where("tr_id = ?", id).Take(&row).Error; err != nil {
			t.Fatalf("select tr_id %d: %v", id, err)
		}
		// lat/long should be updated (non-zero)
		latVal, okLat := row["lat"].(float64)
		longVal, okLong := row["long"].(float64)
		if !okLat || !okLong || latVal == 0 || longVal == 0 {
			t.Errorf("expected non-zero lat/long for tr_id %d, got lat=%v(long ok=%v) long=%v(ok=%v)", id, row["lat"], okLat, row["long"], okLong)
		}
		// address/city/zip_code should match normalized strings from server
		addr, _ := row["address"].(string)
		zip, _ := row["zip_code"].(int64)
		city, _ := row["city"].(string)
		if !strings.HasPrefix(addr, "Normalized Address") {
			t.Errorf("unexpected address for tr_id %d: %v", id, addr)
		}
		if zip < 10000 || zip > 10100 {
			t.Errorf("unexpected zip for tr_id %d: %v", id, zip)
		}
		if !strings.HasPrefix(city, "City") {
			t.Errorf("unexpected city for tr_id %d: %v", id, city)
		}
	}
}

// TestGeocodeDBEmpty ensures GeocodeDB gracefully handles no rows to geocode.
func TestGeocodeDBEmpty(t *testing.T) {
	// create temporary sqlite db file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_geocode_empty.db")
	dsn := "file:" + dbPath

	db := model.ConnectToDB(dsn)
	if err := db.AutoMigrate(&model.Transaction{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// ensure no rows exist and call GeocodeDB
	GeocodeDB(dsn, true, "")

	// nothing to assert beyond no panic; confirm count still zero
	var cnt int64
	if err := db.Table("transactions").Count(&cnt).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 rows, got %d", cnt)
	}
}
