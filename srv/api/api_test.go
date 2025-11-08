package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"jc.org/immotep/model"
)

// helper to open a temporary sqlite DB and return db + dsn
func openTestDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	dsn := "file::memory:?cache=shared"
	db := model.ConnectToDB(dsn)
	if db == nil {
		t.Fatalf("ConnectToDB returned nil for dsn %s", dsn)
	}
	// ensure yearly agg tables exist for some tests
	db.AutoMigrate(&model.CityYearlyAgg{}, &model.DepartmentYearlyAgg{}, &model.RegionYearlyAgg{})
	return db, dsn
}

// seed a minimal dataset: region, department, city and two transactions in two years
func seedMinimal(db *gorm.DB, t *testing.T) {
	t.Helper()

	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM cities")
	db.Exec("DELETE FROM departments")
	db.Exec("DELETE FROM regions")

	// minimal geometries as GeoJSON features
	feat := `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]},"properties":{}}`

	r := model.Region{Code: "R1", Name: "Reg1", Contour: feat}
	d := model.Department{Code: "D1", Name: "Dep1", Contour: feat}
	c := model.City{Code: "C1", Name: "City1", NameUpper: "CITY1", ZipCode: 75000, Population: 1000, Contour: feat, CodeDepartment: "D1", CodeRegion: "R1"}

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
	tr1 := model.Transaction{
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
	tr2 := model.Transaction{
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

func TestServe(t *testing.T) {
	Serve("file::memory:?cache=shared", "asset", 70000, true)
}

// Mock POI struct for testing
type POI struct {
	City    string `json:"city"`
	ZipCode string `json:"zip_code"`
}

func TestInfoRoute(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/info", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	body := w.Body.String()
	expectedBody := `{"status":"UP"}`
	assert.Equal(t, expectedBody, body)
}

func TestRedirectRoute(t *testing.T) {
	router := BuildRouter("memfile", "asset", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 301, w.Code)
	body := w.Body.String()
	expectedBody := "<a href=\"/immotep\">Moved Permanently</a>.\n\n"
	assert.Equal(t, expectedBody, body)
}

func TestPOIsEndpoint(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)
	router := BuildRouter(dsn, "", true)

	tests := []struct {
		name         string
		query        string
		wantStatus   int
		expectedBody string
	}{
		{
			name:         "Basic POIs request",
			query:        "/api/pois",
			wantStatus:   http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "POIs with limit",
			query:        "/api/pois?limit=10",
			wantStatus:   http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "POIs with zip",
			query:        "/api/pois?zip=75001",
			wantStatus:   http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "POIs with zip",
			query:        "/api/pois?after=2023",
			wantStatus:   http.StatusOK,
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			var pois []POI
			err := json.Unmarshal(w.Body.Bytes(), &pois)
			assert.NoError(t, err)

		})
	}
}

func TestPOIsFilterEndpoint(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	filterBody := FilterInfoBody{
		NorthEast: LatLongInfo{Lat: 48.86, Long: 2.35},
		SouthWest: LatLongInfo{Lat: 48.85, Long: 2.34},
		After:     "2020-01-01",
	}

	body, _ := json.Marshal(filterBody)

	tests := []struct {
		name       string
		query      string
		body       []byte
		wantStatus int
	}{
		{
			name:       "Basic filter request",
			query:      "/api/pois/filter",
			body:       body,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Filter with limit",
			query:      "/api/pois/filter?limit=10",
			body:       body,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid body",
			query:      "/api/pois/filter",
			body:       []byte(`{"invalid": "json"`),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.query, bytes.NewBuffer(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestCitiesEndpoint(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "Basic cities request",
			query:      "/api/cities",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Cities with department",
			query:      "/api/cities?dep=75",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestCitiesPostEndpoint(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	body := []byte(`{
		"northEast": {"lat": 48.86, "lng": 2.35},
		"southWest": {"lat": 48.85, "lng": 2.34}
	}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/cities", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRegionsEndpoint(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/regions", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDepartmentsEndpoint(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/departments", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetPOIs(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	// Create a new recorder
	w := httptest.NewRecorder()

	// Create a new request
	req, _ := http.NewRequest("GET", "/api/pois?zip=75001", nil)

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check the response body
	var pois []POI
	err := json.Unmarshal(w.Body.Bytes(), &pois)
	assert.NoError(t, err)
	assert.Len(t, pois, 0)
}
