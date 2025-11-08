// Package api implements the HTTP server, routing and request handlers for the
// immotep application. It exposes REST endpoints to query transactions (POIs),
// cities, departments and regions and serves the UI static assets.
//
// Responsibilities:
// - Build and configure a Gin router with API routes and static file serving.
// - Manage a shared GORM DB connection used by handlers.
// - Provide request/response payload types used by the API.
package api

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"jc.org/immotep/model"
)

// immotepDB holds the shared GORM DB connection used by handlers.
// It is initialized by BuildRouter/Serve when a DSN is provided.
var immotepDB *gorm.DB

// immotepDSN stores the database connection string used to reconnect
// lazily if immotepDB is nil inside handlers.
var immotepDSN string

// staticFS contains embedded UI assets used when no external staticDir is set.
//
//go:embed immotep/*
var staticFS embed.FS

// Serve starts the HTTP server.
//
// Parameters:
//   - dsn: database connection string (Postgres or SQLite)
//   - staticDir: path to local static assets directory; when empty, embedded assets are used
//   - port: TCP port to listen on
//   - debug: when true, enable Gin debug mode
//
// Behavior:
//   - Builds the router, installs a signal handler (SIGINT/SIGTERM) to exit
//     cleanly and runs the Gin engine listening on the specified port.
func Serve(dsn string, staticDir string, port int, debug bool) {
	router := BuildRouter(dsn, staticDir, debug)

	// setup signal handling for termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Infof("CTRL+C or SIGTERM to stop...")
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		os.Exit(0)
	}()

	// start HTTP server
	router.Run(fmt.Sprintf(":%v", port)) // listen and serve on 0.0.0.0:${port}
}

// BuildRouter creates and configures a Gin engine with API routes and static
// file serving. It also establishes a DB connection used by the handlers.
//
// Parameters:
//   - dsn: database connection string used to initialise immotepDB
//   - staticDir: when non-empty, local static files are served from this folder
//   - debug: enable Gin debug mode when true
//
// Returns:
//   - *gin.Engine: the configured Gin engine ready to Run().
func BuildRouter(dsn, staticDir string, debug bool) *gin.Engine {

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// connect to Database
	immotepDSN = dsn
	immotepDB = model.ConnectToDB(immotepDSN)

	apigroup := engine.Group("/api")
	addRoutes(apigroup)

	engine.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	if staticDir == "" {
		log.Infof("Serve static data internally.\n")
		s, _ := fs.Sub(staticFS, "immotep")
		engine.StaticFS("/immotep", http.FS(s))
	} else {
		log.Infof("Serve static data from: %v\n", staticDir)
		engine.Static("/immotep", staticDir)
	}

	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/immotep")
	})

	return engine
}

// LatLongInfo represents a latitude/longitude pair used in JSON payloads.
//
// Fields are validated via Gin binding tags when used as handler parameters.
type LatLongInfo struct {
	Lat  float64 `json:"lat" binding:"required"`
	Long float64 `json:"lng" binding:"required"`
}

// FilterInfoBody is the JSON body expected by spatial filter endpoints.
//
// It contains a bounding box (northEast / southWest), an optional department
// code and an optional 'after' date string.
type FilterInfoBody struct {
	NorthEast LatLongInfo `json:"northEast" binding:"required"`
	SouthWest LatLongInfo `json:"southWest" binding:"required"`
	After     string      `form:"after"`
}

// POISQuery models query parameters accepted by POI/city endpoints.
type POISQuery struct {
	Limit   int    `form:"limit"`
	Year    int    `form:"year"`
	ZipCode int    `form:"zip"`
	DepCode string `form:"dep"`
	After   string `form:"after"`
}

// addRoutes registers all API endpoints on the provided router group.
//
// It wires handlers for:
//   - GET  /api/pois        : query POIs with optional filters
//   - POST /api/pois/filter : bounding-box search for POIs
//   - GET  /api/cities      : list cities (optional department filter)
//   - POST /api/cities      : bounding-box search for cities
//   - GET  /api/regions     : list regions
//   - GET  /api/departments : list departments
//
// Handlers lazily ensure immotepDB is connected (reconnect using immotepDSN).
func addRoutes(rg *gin.RouterGroup) {

	/*
		/pois?zip={}&limit={}&dep={}&after={}
	*/
	rg.GET("/pois", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		zip := -1
		limit := -1
		after := ""

		// get value from query param
		var param POISQuery
		if c.ShouldBindQuery(&param) == nil {
			if param.Limit >= 0 {
				limit = param.Limit
			}
			if param.ZipCode >= 0 {
				zip = param.ZipCode
			}

			if param.After != "" {
				after = param.After
			}
		}

		pois := model.GetPOI(immotepDB, limit, zip, after)
		if pois == nil {
			c.JSON(500, []model.TransactionPOI{})
			return
		}
		c.JSON(200, pois)
	})

	rg.POST("/pois/filter", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		limit := -1
		year := -1

		// get value from query param
		var param POISQuery
		if c.ShouldBindQuery(&param) == nil {
			if param.Limit >= 0 {
				limit = param.Limit
			}
			if param.Year >= 0 {
				year = param.Year
			}
		}

		var body FilterInfoBody
		err := c.BindJSON(&body)
		if err != nil {
			log.Printf("Error in POST /pois/filter: %v\n", err)
			c.JSON(500, model.BoundedTransactionInfo{})
			return
		}

		pois := model.GetPOIFromBounds(immotepDB,
			body.NorthEast.Lat, body.NorthEast.Long,
			body.SouthWest.Lat, body.SouthWest.Long,
			limit, body.After, year)

		if pois == nil {
			c.JSON(500, model.BoundedTransactionInfo{})
			return
		}
		c.JSON(200, pois)
	})

	/*
		/city?limit={}&dep={}
	*/
	rg.GET("/cities", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		dep := ""

		// get value from query param
		var param POISQuery
		if c.ShouldBindQuery(&param) == nil {
			if param.DepCode != "" {
				dep = param.DepCode
			}
		}

		log.Debugf("Get city info for dep %v\n", dep)

		infos := model.GetCityDetails(immotepDB, dep)

		c.JSON(200, infos)

	})

	rg.POST("/cities", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		limit := -1

		// get value from query param
		var param POISQuery
		if c.ShouldBindQuery(&param) == nil {
			if param.Limit >= 0 {
				limit = param.Limit
			}
		}

		var body FilterInfoBody
		err := c.BindJSON(&body)
		if err != nil {
			log.Printf("Error in POST /cities: %v\n", err)
			c.JSON(400, nil)
			return
		}

		infos := model.GetCitiesFromBounds(immotepDB,
			body.NorthEast.Lat, body.NorthEast.Long,
			body.SouthWest.Lat, body.SouthWest.Long,
			limit)

		if infos == nil {
			c.JSON(500, nil)
			return
		}
		c.JSON(200, infos)
	})

	rg.GET("/regions", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		infos := model.GetRegionDetails(immotepDB)

		c.JSON(200, infos)

	})

	rg.GET("/departments", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		infos := model.GetDepartmentDetails(immotepDB)

		c.JSON(200, infos)

	})
}
