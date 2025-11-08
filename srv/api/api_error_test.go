package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfoRouteError(t *testing.T) {
	router := BuildRouter("memfile", "", false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/info", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	body := w.Body.String()
	expectedBody := `{"status":"UP"}`
	assert.Equal(t, expectedBody, body)
}

func TestRedirectRouteError(t *testing.T) {
	router := BuildRouter("memfile", "asset", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 301, w.Code)
	body := w.Body.String()
	expectedBody := "<a href=\"/immotep\">Moved Permanently</a>.\n\n"
	assert.Equal(t, expectedBody, body)
}

func TestPOIsEndpointError(t *testing.T) {
	router := BuildRouter("memfile", "", true)

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "Basic POIs request",
			query:      "/api/pois",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "POIs with limit",
			query:      "/api/pois?limit=10",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "POIs with zip",
			query:      "/api/pois?zip=75001",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "POIs with zip",
			query:      "/api/pois?after=2023",
			wantStatus: http.StatusInternalServerError,
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

func TestPOIsFilterEndpointError(t *testing.T) {
	router := BuildRouter("memfile", "", true)

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
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Filter with limit",
			query:      "/api/pois/filter?limit=10",
			body:       body,
			wantStatus: http.StatusInternalServerError,
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

func TestCitiesEndpointError(t *testing.T) {
	router := BuildRouter("memfile", "", true)

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "Basic cities request",
			query:      "/api/cities",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Cities with department",
			query:      "/api/cities?dep=75",
			wantStatus: http.StatusInternalServerError,
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

func TestCitiesPostEndpointError(t *testing.T) {
	db, dsn := openTestDB(t)
	seedMinimal(db, t)

	router := BuildRouter(dsn, "", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/cities", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRegionsEndpointError(t *testing.T) {
	router := BuildRouter("memfile", "", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/regions", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDepartmentsEndpointError(t *testing.T) {
	router := BuildRouter("memfile", "", true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/departments", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetPOIsError(t *testing.T) {
	router := BuildRouter("memfile", "", true)

	// Create a new recorder
	w := httptest.NewRecorder()

	// Create a new request
	req, _ := http.NewRequest("GET", "/api/pois?zip=75001", nil)

	// Serve the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check the response body
	var pois []POI
	err := json.Unmarshal(w.Body.Bytes(), &pois)
	assert.NoError(t, err)
	assert.Len(t, pois, 0)
}
