# Immotep Project

## 1. Overview

The Immotep project is a web application designed to visualize and analyze real estate data across France. It comprises a backend (srv) written in Go and a frontend (ui) built with React and leaflet. The application fetches real estate transaction data, geocodes addresses, aggregates statistics, and presents this information on an interactive map.

## 2. System Architecture

The system follows a multi-tiered architecture:

* **Data Ingestion Tier:** Responsible for collecting and preparing raw data.
* **Backend Tier (API):** Provides RESTful API endpoints for data access and processing.
* **Frontend Tier (UI):** Presents an interactive user interface for data visualization and analysis.
* **Database Tier:** Stores processed and aggregated data.

## 3. Component Details

### 3.1. Data Ingestion Tier

* **Data Source:** The primary data source is the French Land Registry (DVF), providing historical real estate transaction data. Additional data sources include the French government's API for geographical information (regions, departments, cities).
* **Data Download:** The `data/download_raw_data.sh` script automates the download of raw data files from DVF and geographical information from the French government's API.
* **Data Loading:** The `srv/loader` package handles parsing and loading raw data into the database.
  * [`loader.LoadRawData`](srv/loader/loader.go) parses transaction data from text files.
  * [`loader.LoadRegion`](srv/loader/loader.go), [`loader.LoadDepartment`](srv/loader/loader.go), and [`loader.LoadCity`](srv/loader/loader.go) load geographical boundaries from GeoJSON files.
* **Geocoding:** The `srv/loader/geocode.go` package uses the French government's address API to geocode transaction records, adding latitude and longitude coordinates to each transaction.

### 3.2. Backend Tier (srv)

* **Language:** Go
* **Framework:** Gin Web Framework
* **Packages:**
  * `cmd`: Defines command-line interface (CLI) commands using Cobra.
    * [`cmd.rootCmd`](srv/cmd/root.go): The root command defines global flags and subcommands.
    * [`cmd.loadCmd`](srv/cmd/root.go): Loads raw data into the database.
    * [`cmd.loadConfCmd`](srv/cmd/root.go): Loads configuration data (regions, departments, cities) into the database.
    * [`cmd.geocodeCmd`](srv/cmd/root.go): Geocodes transaction data.
    * [`cmd.computeCmd`](srv/cmd/root.go): Computes statistics.
    * [`cmd.aggregateCmd`](srv/cmd/root.go): Aggregates data for yearly analysis.
    * [`cmd.serveCmd`](srv/cmd/root.go): Starts the API server.
  * `api`: Implements the REST API endpoints using Gin.
    * [`api.Serve`](srv/api/api.go): Starts the HTTP server.
    * [`api.BuildRouter`](srv/api/api.go): Configures the Gin router and defines API routes.
    * API endpoints:
      * `/pois`: Returns points of interest (real estate transactions) based on query parameters.
      * `/pois/filter`: Returns points of interest based on a geographic bounding box.
      * `/cities`: Returns city details.
      * `/regions`: Returns region details.
      * `/departments`: Returns department details.
  * `model`: Defines the data model and database interactions using GORM.
    * [`model.ConnectToDB`](srv/model/model.go): Establishes a connection to the database (PostgreSQL or SQLite).
    * Defines data structures (e.g., `Transaction`, `Region`, `Department`, `City`).
    * Provides functions for querying and manipulating data (e.g., [`model.GetPOI`](srv/model/model.go), [`model.GetCityDetails`](srv/model/model.go), [`model.GetRegionDetails`](srv/model/model.go), [`model.GetDepartmentDetails`](srv/model/model.go)).
  * `loader`: Handles loading data from raw files into the database.
    * [`loader.LoadRawData`](srv/loader/loader.go): Parses and loads transaction data.
    * [`loader.LoadRegion`](srv/loader/loader.go), [`loader.LoadDepartment`](srv/loader/loader.go), [`loader.LoadCity`](srv/loader/loader.go): Loads geographical data.
    * [`loader.GeocodeDB`](srv/loader/geocode.go): Geocodes transaction addresses.
  * `aggregate`: Aggregates data for yearly statistics.
    * [`model.AggregateData`](srv/model/aggregate.go): Orchestrates the aggregation process.
    * [`model.aggregateCities`](srv/model/aggregate.go), [`model.aggregateDepartments`](srv/model/aggregate.go), [`model.aggregateRegions`](srv/model/aggregate.go): Performs the aggregation calculations.
  * `compute`: Computes statistics.
    * [`model.ComputeStat`](srv/model/compute.go): Orchestrates the computation process.
    * [`model.ComputeRegions`](srv/model/compute.go), [`model.ComputeDepartments`](srv/model/compute.go), [`model.ComputeCities`](srv/model/compute.go): Performs the computation calculations.
* **Configuration:** Uses Viper/Cobra for parameter management, allowing settings to be defined via command-line flags, environment variables, or configuration files.

### 3.3. Frontend Tier (ui)

* **Language:** JavaScript
* **Framework:** React
* **State Management:** Redux Toolkit
* **UI Components:** Material-UI
* **Mapping Library:** Leaflet
* **Build Tool:** Vite
* **Key Components:**
  * [`App`](ui/frontend/App.js): The main application component that sets up the theme and layout.
  * [`MapViewer`](ui/frontend/MapViewer.js): The core component responsible for rendering the interactive map using Leaflet. It displays tile layers, city, department, and region boundaries, and location markers.
  * [`CityStat`](ui/frontend/CityStat.js), [`DepartmentStat`](ui/frontend/DepartmentStat.js), [`RegionStat`](ui/frontend/RegionStat.js): Components that fetch and display geographical data with statistical information.
  * [`LocationMarker`](ui/frontend/LocationMarker.js): Manages the display of individual transaction points on the map.
  * [`Menubar`](ui/frontend/Menubar.js): Provides a user interface for controlling the map, filtering data, and toggling settings.
  * [`poi_service`](ui/frontend/poi_service.js): An Axios instance configured to make API requests to the backend.
  * [`store`](ui/frontend/store/index.js): Configures the Redux store.
  * [`uiparamSlice`](ui/frontend/store/uiparamSlice.js): A Redux slice that manages UI parameters such as the selected department, map zoom, and theme.
* **Internationalization:** Uses `i18next` for multi-language support, with translations stored in JSON files.

### 3.4. Database Tier

* **Database:** PostgreSQL/PostGIS or SQLite
* **ORM:** GORM
* **Data Model:** Defined in the `srv/model` package, including tables for:
  * `transactions`: Real estate transaction records.
  * `regions`: Geographical region boundaries.
  * `departments`: Geographical department boundaries.
  * `cities`: Geographical city boundaries.
  * `city_yearly_aggs`: Aggregated city data per year.
  * `department_yearly_aggs`: Aggregated department data per year.
  * `region_yearly_aggs`: Aggregated region data per year.

## 5. Build Process

The `build.sh` script automates the build process:

1. Cleans the UI and backend directories.
2. Builds the UI using `npm run build`.
3. Copies the UI build output to the backend's static asset directory.
4. Builds the Go backend.
5. Optionally builds a Docker image.
6. Optionally pushes the Docker image to a container registry and deploys to Kubernetes.

## 6. Installation

See [guide](./install/INSTALLATION.md) in install folder.

## 7. Future Enhancements

* Add more sophisticated data analysis features.
* Support additional data sources.
* Improve the UI with more advanced charting and visualization options.
* Implement user authentication and authorization.
