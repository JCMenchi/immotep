package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"jc.org/immotep/loader"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// package variable
// name of viper config file set as command line arg
var cfgFile string

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

	// define flags comon to all commands

	// define viper config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.immotep.yaml)")

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
	loadCmd.PersistentFlags().StringP("zipcode-filename", "z", "", "zip code definition file")
	viper.BindPFlag("zipcode.filename", loadCmd.PersistentFlags().Lookup("zipcode-filename"))

	rootCmd.AddCommand(loadCmd)

	rootCmd.AddCommand(geocodeCmd)
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

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

/* Build database connection string
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
		// load zip code
		zipcodeMap := loader.ReadZipcodeMap(viper.GetString("zipcode.filename"))

		// load data
		dsn := getDSN()
		fmt.Printf("load data to db: %v\n", dsn)
		for i, a := range args {
			fmt.Printf("load data file(%v): %v\n", i, a)
			loader.LoadRawData(getDSN(), a, zipcodeMap)
		}
	},
}

var geocodeCmd = &cobra.Command{
	Use:   "geocode",
	Short: "geocode db",
	Long:  `geocode db`,
	Run: func(cmd *cobra.Command, args []string) {
		// load data
		dsn := getDSN()
		loader.GeocodeDB(dsn)
	},
}