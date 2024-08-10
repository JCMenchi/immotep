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

// store Postgresql connection
var immotepDB *gorm.DB
var immotepDSN string

//go:embed immotep/*
var staticFS embed.FS

// Start HTTP server
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

type LatLongInfo struct {
	Lat  float64 `json:"lat" binding:"required"`
	Long float64 `json:"lng" binding:"required"`
}

type FilterInfoBody struct {
	NorthEast LatLongInfo `json:"northEast" binding:"required"`
	SouthWest LatLongInfo `json:"southWest" binding:"required"`
	DepCode   string      `json:"code"`
	After     string      `form:"after"`
}

type POISQuery struct {
	Limit   int    `form:"limit"`
	Year    int    `form:"year"`
	ZipCode int    `form:"zip"`
	DepCode string `form:"dep"`
	After   string `form:"after"`
}

func addRoutes(rg *gin.RouterGroup) {

	/*
		/pois?zip={}&limit={}&dep={}&after={}
	*/
	rg.GET("/pois", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		zip := -1
		dep := ""
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

			if param.DepCode != "" {
				dep = param.DepCode
			}

			if param.After != "" {
				after = param.After
			}
		}

		pois := model.GetPOI(immotepDB, limit, zip, dep, after)
		if pois == nil {
			c.JSON(500, "[]")
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
			c.JSON(500, "")
			return
		}

		pois := model.GetPOIFromBounds(immotepDB,
			body.NorthEast.Lat, body.NorthEast.Long,
			body.SouthWest.Lat, body.SouthWest.Long,
			limit, body.DepCode, body.After, year)

		if pois == nil {
			c.JSON(500, "[]")
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
