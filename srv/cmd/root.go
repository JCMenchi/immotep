// Package cmd implements the command-line interface for the immotep application.
// It uses cobra & viper for command line parsing.
//
// The main commands provided are:
// - load: Load raw data into the database
// - loadconf: Load configuration data (regions, departments, cities)
// - geocode: Geocode addresses in the database
// - compute: Compute statistics on the data
// - aggregate: Aggregate data for analysis
// - serve: Start the REST API server & UI asset server
//
// Configuration can be provided via:
// - Command line flags
// - Config file ($HOME/.immotep.yaml by default)
// - Environment variables (prefixed with IMMOTEP_)
//
// Database support includes:
// - PostgreSQL: Using connection string format postgres://user:pass@host:port/dbname
// - SQLite: Using file:filename or in-memory with file::memory:?cache=shared
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"jc.org/immotep/api"
	"jc.org/immotep/loader"
	"jc.org/immotep/model"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// package variable
// name of viper config file set as command line arg
var cfgFile string

var debugMode bool

const VERSION = "1.1.0"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:       "immotep",
	Short:     "immotep",
	Long:      `immotep.`,
	Version:   VERSION,
	ValidArgs: []string{"load", "serve"},
}

// Execute runs the root command. It is called by main.main() to start the application
// and handles all command line parsing and validation.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

// init initializes the command line interface by setting up all flags, commands,
// and their bindings to viper configuration. It configures:
// - Global flags for configuration and debugging
// - Database connection parameters
// - All subcommands (load, loadconf, geocode, serve, compute, aggregate)
func init() {
	cobra.OnInitialize(initConfig)

	// define flags common to all commands

	// define viper config file
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.immotep.yaml)")

	// debug
	RootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "activate debug logs")

	// define Database connection string
	RootCmd.PersistentFlags().StringP("dsn-type", "t", "pgsql", "immotep Database type one of pgsql or sqlite (default pgsql)")
	viper.BindPFlag("dsn.type", RootCmd.PersistentFlags().Lookup("dsn-type"))
	viper.SetDefault("dsn.type", "sqlite")

	RootCmd.PersistentFlags().StringP("dsn-host", "H", "pgsql", "immotep Database Hostname")
	viper.BindPFlag("dsn.host", RootCmd.PersistentFlags().Lookup("dsn-host"))
	viper.SetDefault("dsn.host", "pgsql")

	RootCmd.PersistentFlags().StringP("dsn-dbname", "d", "db", "immotep Database name")
	viper.BindPFlag("dsn.dbname", RootCmd.PersistentFlags().Lookup("dsn-dbname"))
	viper.SetDefault("dsn.dbname", "db")

	RootCmd.PersistentFlags().IntP("dsn-port", "P", 5432, "immotep Database port")
	viper.BindPFlag("dsn.port", RootCmd.PersistentFlags().Lookup("dsn-port"))
	viper.SetDefault("dsn.port", 5432)

	RootCmd.PersistentFlags().StringP("dsn-user", "u", "", "immotep Database User Name")
	viper.BindPFlag("dsn.user", RootCmd.PersistentFlags().Lookup("dsn-user"))

	RootCmd.PersistentFlags().StringP("dsn-password", "p", "", "immotep Database User password")
	viper.BindPFlag("dsn.password", RootCmd.PersistentFlags().Lookup("dsn-password"))

	RootCmd.PersistentFlags().StringP("dsn-filename", "f", "imm.db", "immotep Database file name (for SQLITE)")
	viper.BindPFlag("dsn.filename", RootCmd.PersistentFlags().Lookup("dsn-filename"))

	// create subcommand
	RootCmd.AddCommand(loadCmd)

	RootCmd.AddCommand(geocodeCmd)

	loadConfCmd.PersistentFlags().StringP("region", "r", "", "region GEOJSON file")
	viper.BindPFlag("file.region", loadConfCmd.PersistentFlags().Lookup("region"))
	loadConfCmd.PersistentFlags().String("department", "", "department GEOJSON file")
	viper.BindPFlag("file.department", loadConfCmd.PersistentFlags().Lookup("department"))
	loadConfCmd.PersistentFlags().String("city", "", "city JSON file")
	viper.BindPFlag("file.city", loadConfCmd.PersistentFlags().Lookup("city"))
	loadConfCmd.PersistentFlags().String("citygeo", "", "city GEOJSON file")
	viper.BindPFlag("file.citygeo", loadConfCmd.PersistentFlags().Lookup("citygeo"))
	RootCmd.AddCommand(loadConfCmd)

	serveCmd.PersistentFlags().Int("port", 8080, "api server port")
	viper.BindPFlag("serve.port", serveCmd.PersistentFlags().Lookup("port"))
	serveCmd.PersistentFlags().String("static", "", "asset folder")
	viper.BindPFlag("serve.static", serveCmd.PersistentFlags().Lookup("static"))
	RootCmd.AddCommand(serveCmd)

	RootCmd.AddCommand(computeCmd)
	RootCmd.AddCommand(aggregateCmd)
}

// initConfig reads in configuration from config files and environment variables.
// It searches for config in the following order:
// 1. Specified config file via --config flag
// 2. $HOME/.immotep.yaml
// 3. Environment variables prefixed with IMMOTEP_
// It also initializes logging based on debug mode.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".immotep.yaml".
		viper.AddConfigPath(home)
		viper.SetConfigName(".immotep")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("IMMOTEP")
	replacer := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	if debugMode {
		log.SetLevel(log.DebugLevel)
	}
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v\n", viper.ConfigFileUsed())
	}
}

// getDSN builds database connection string
//
// e.g.
//
//	postgres://<user>:userpwd@pgsql:5432/db?sslmode=disable
//	sqlite file:mydata.db
func getDSN() string {
	dbtype := viper.GetString("dsn.type")

	// by default use in mem db
	dsn := "file::memory:?cache=shared"

	switch dbtype {
	case "pgsql":
		dsn = "postgres://"
		dsn += viper.GetString("dsn.user") + ":"
		dsn += viper.GetString("dsn.password") + "@"
		dsn += viper.GetString("dsn.host") + ":"
		dsn += fmt.Sprint(viper.GetInt("dsn.port")) + "/"
		dsn += viper.GetString("dsn.dbname")
		dsn += "?sslmode=disable"
	case "sqlite":
		dsn = "file:"
		dsn += viper.GetString("dsn.filename")
	}

	return dsn
}

// loadCmd represents the command for loading raw data into the database.
// Usage: immotep load [rawdatafile...]
// It accepts multiple data files as arguments and loads them sequentially.
var loadCmd = &cobra.Command{
	Use:   "load [rawdatafile.]",
	Short: "load raw data",
	Long:  `load raw data`,
	Run: func(cmd *cobra.Command, args []string) {
		// load data
		dsn := getDSN()
		log.Infof("load data to db: %v\n", dsn)
		for i, a := range args {
			log.Infof("load data file(%v): %v\n", i, a)
			loader.LoadRawData(getDSN(), a)
		}
	},
}

// loadConfCmd represents the command for loading configuration data like regions,
// departments, and cities into the database.
// Usage: immotep loadconf [flags]
// Flags:
//
//	--region: region GEOJSON file
//	--department: department GEOJSON file
//	--city: city JSON file
//	--citygeo: city GEOJSON file
var loadConfCmd = &cobra.Command{
	Use:   "loadconf",
	Short: "load config",
	Long:  `load config`,
	Run: func(cmd *cobra.Command, args []string) {
		// load region
		region := viper.GetString("file.region")
		department := viper.GetString("file.department")
		city := viper.GetString("file.city")
		cityGeo := viper.GetString("file.citygeo")
		// load data
		dsn := getDSN()
		log.Infof("load conf to db: %v\n", dsn)

		if region != "" {
			loader.LoadRegion(dsn, region)
		}

		if department != "" {
			loader.LoadDepartment(dsn, department)
		}

		if city != "" {
			loader.LoadCity(dsn, city, cityGeo)
		}

	},
}

// geocodeCmd represents the command for geocoding addresses in the database.
// Usage: immotep geocode [department...]
// If no department is specified, it geocodes all entries.
// If departments are specified, it only geocodes entries in those departments.
var geocodeCmd = &cobra.Command{
	Use:   "geocode",
	Short: "geocode db",
	Long:  `geocode db`,
	Run: func(cmd *cobra.Command, args []string) {
		// geo code address
		dsn := getDSN()
		log.Infof("geocode db: %v\n", dsn)
		if len(args) > 0 {
			for _, a := range args {
				log.Infof("geocode dep: %v\n", a)
				loader.GeocodeDB(dsn, false, a)
			}
		} else {
			loader.GeocodeDB(dsn, true, "")
		}
	},
}

// computeCmd represents the command for computing statistics on the data.
// Usage: immotep compute
// It processes the data and generates statistical computations stored in the database.
var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "compute db",
	Long:  `compute db`,
	Run: func(cmd *cobra.Command, args []string) {
		// geo code address
		dsn := getDSN()
		log.Infof("compute db: %v\n", dsn)
		model.ComputeStat(dsn)
	},
}

// aggregateCmd represents the command for aggregating data for analysis.
// Usage: immotep aggregate
// It processes the data and creates aggregate views for analysis purposes.
var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "aggregate db",
	Long:  `aggregate db`,
	Run: func(cmd *cobra.Command, args []string) {
		// geo code address
		dsn := getDSN()
		log.Infof("aggregate db: %v\n", dsn)
		model.AggregateData(dsn)
	},
}

// serveCmd represents the command for starting the REST API server.
// Usage: immotep serve [flags]
// Flags:
//
//	--port: API server port (default 8080)
//	--static: Static assets folder
//
// It starts an HTTP server that provides access to the immotep data via REST API.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "immotep backend",
	Long:  `immotep REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		dsn := getDSN()
		port := viper.GetInt("serve.port")
		staticDir := viper.GetString("serve.static")

		log.Infof("load conf to db: %v\n", dsn)

		api.Serve(dsn, staticDir, port, debugMode)
	},
}
