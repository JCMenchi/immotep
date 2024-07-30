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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:       "immotep",
	Short:     "immotep",
	Long:      `immotep.`,
	Version:   "1.0.0",
	ValidArgs: []string{"load", "serve"},
}

// Decode command line arguments
// This is called by main.main().
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// Initialize cli flags and viper config
func init() {
	cobra.OnInitialize(initConfig)

	// define flags common to all commands

	// define viper config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.immotep.yaml)")

	// debug
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "activate debug logs")

	// define Database connection string
	rootCmd.PersistentFlags().StringP("dsn-type", "t", "pgsql", "immotep Database type one of pgsql or sqlite (default pgsql)")
	viper.BindPFlag("dsn.type", rootCmd.PersistentFlags().Lookup("dsn-type"))
	viper.SetDefault("dsn.type", "sqlite")

	rootCmd.PersistentFlags().StringP("dsn-host", "H", "pgsql", "immotep Database Hostname")
	viper.BindPFlag("dsn.host", rootCmd.PersistentFlags().Lookup("dsn-host"))
	viper.SetDefault("dsn.host", "pgsql")

	rootCmd.PersistentFlags().StringP("dsn-dbname", "d", "db", "immotep Database name")
	viper.BindPFlag("dsn.dbname", rootCmd.PersistentFlags().Lookup("dsn-dbname"))
	viper.SetDefault("dsn.dbname", "db")

	rootCmd.PersistentFlags().IntP("dsn-port", "P", 5432, "immotep Database port")
	viper.BindPFlag("dsn.port", rootCmd.PersistentFlags().Lookup("dsn-port"))
	viper.SetDefault("dsn.port", 5432)

	rootCmd.PersistentFlags().StringP("dsn-user", "u", "", "immotep Database User Name")
	viper.BindPFlag("dsn.user", rootCmd.PersistentFlags().Lookup("dsn-user"))

	rootCmd.PersistentFlags().StringP("dsn-password", "p", "", "immotep Database User password")
	viper.BindPFlag("dsn.password", rootCmd.PersistentFlags().Lookup("dsn-password"))

	rootCmd.PersistentFlags().StringP("dsn-filename", "f", "imm.db", "immotep Database file name (for SQLITE)")
	viper.BindPFlag("dsn.filename", rootCmd.PersistentFlags().Lookup("dsn-filename"))

	// create subcommand
	rootCmd.AddCommand(loadCmd)

	rootCmd.AddCommand(geocodeCmd)

	loadConfCmd.PersistentFlags().StringP("region", "r", "", "region GEOJSON file")
	viper.BindPFlag("file.region", loadConfCmd.PersistentFlags().Lookup("region"))
	loadConfCmd.PersistentFlags().String("department", "", "department GEOJSON file")
	viper.BindPFlag("file.department", loadConfCmd.PersistentFlags().Lookup("department"))
	loadConfCmd.PersistentFlags().String("city", "", "city JSON file")
	viper.BindPFlag("file.city", loadConfCmd.PersistentFlags().Lookup("city"))
	loadConfCmd.PersistentFlags().String("citygeo", "", "city GEOJSON file")
	viper.BindPFlag("file.citygeo", loadConfCmd.PersistentFlags().Lookup("citygeo"))
	rootCmd.AddCommand(loadConfCmd)

	serveCmd.PersistentFlags().Int("port", 8080, "api server port")
	viper.BindPFlag("serve.port", serveCmd.PersistentFlags().Lookup("port"))
	serveCmd.PersistentFlags().String("static", "", "asset folder")
	viper.BindPFlag("serve.static", serveCmd.PersistentFlags().Lookup("static"))
	rootCmd.AddCommand(serveCmd)

	rootCmd.AddCommand(computeCmd)
}

// initConfig reads in config file and ENV variables if set.
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

/*
	Build database connection string

e.g. postgres://user:userpwd@pgsql:5432/db?sslmode=disable

	file:mydata.db
*/
func getDSN() string {
	dbtype := viper.GetString("dsn.type")

	// by default use in mem db
	dsn := "file::memory:?cache=shared"

	if dbtype == "pgsql" {
		dsn = "postgres://"
		dsn += viper.GetString("dsn.user") + ":"
		dsn += viper.GetString("dsn.password") + "@"
		dsn += viper.GetString("dsn.host") + ":"
		dsn += fmt.Sprint(viper.GetInt("dsn.port")) + "/"
		dsn += viper.GetString("dsn.dbname")
		dsn += "?sslmode=disable"
	} else if dbtype == "sqlite" {
		dsn = "file:"
		dsn += viper.GetString("dsn.filename")
	}

	return dsn
}

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

/*
./immotep -f imm.db loadconf --region data/regions.geojson --department data/departements.geojson --city data/communes.json --citygeo data/communes.geojson
*/
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
				loader.GeocodeDB(dsn, a)
			}
		} else {
			loader.GeocodeDB(dsn, "")
		}
	},
}

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

// serveCmd represents the serve command
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
