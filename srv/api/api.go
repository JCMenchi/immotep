package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jc.org/immotep/model"
)

// store Postgresql connection
var immotepDB *gorm.DB
var immotepDSN string

// Start HTTP server
func Serve(dsn string, staticDir string, port int) {
	router := BuildRouter(dsn, staticDir)

	router.Run(fmt.Sprintf(":%v", port)) // listen and serve on 0.0.0.0:${port}
}

func BuildRouter(dsn, staticDir string) *gin.Engine {
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

	engine.Static("/immotep", staticDir)

	return engine
}

type LatLongInfo struct {
	Lat  float64 `json:"lat" binding:"required"`
	Long float64 `json:"lng" binding:"required"`
}

type FilterInfoBody struct {
	NorthEast LatLongInfo `json:"northEast" binding:"required"`
	SouthWest LatLongInfo `json:"southWest" binding:"required"`
	DepCode   int         `json:"code"`
	After     string      `form:"after"`
}

type POISQuery struct {
	Limit   int    `form:"limit"`
	ZipCode int    `form:"zip"`
	DepCode int    `form:"dep"`
	After   string `form:"after"`
}

func addRoutes(rg *gin.RouterGroup) {

	rg.GET("/pois", func(c *gin.Context) {
		if immotepDB == nil {
			immotepDB = model.ConnectToDB(immotepDSN)
		}

		zip := -1
		dep := -1
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

			if param.DepCode >= 0 {
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
			log.Printf("Error in POST /players: %v\n", err)
			c.JSON(500, "")
			return
		}

		pois := model.GetPOIFromBounds(immotepDB,
			body.NorthEast.Lat, body.NorthEast.Long,
			body.SouthWest.Lat, body.SouthWest.Long,
			limit, body.DepCode, body.After)

		if pois == nil {
			c.JSON(500, "[]")
			return
		}
		c.JSON(200, pois)
	})

}
