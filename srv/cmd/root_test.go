package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Test that all commands are properly registered
	if RootCmd.Use != "immotep" {
		t.Errorf("Expected root command name to be 'immotep', got %s", RootCmd.Use)
	}

	commands := []string{"load", "loadconf", "geocode", "serve", "compute", "aggregate"}
	for _, searchCmd := range commands {
		var found = false
		for _, cmd := range RootCmd.Commands() {
			if strings.HasPrefix(cmd.Use, searchCmd) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command %s not found in root commands", searchCmd)
		}
	}
}

func TestGetDSN(t *testing.T) {
	tests := []struct {
		name     string
		dbType   string
		host     string
		port     int
		user     string
		password string
		dbname   string
		filename string
		want     string
	}{
		{
			name:   "default sqlite memory",
			dbType: "",
			want:   "file::memory:?cache=shared",
		},
		{
			name:     "postgres config",
			dbType:   "pgsql",
			host:     "localhost",
			port:     5432,
			user:     "testuser",
			password: "testpass",
			dbname:   "testdb",
			want:     "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
		},
		{
			name:     "sqlite file",
			dbType:   "sqlite",
			filename: "test.db",
			want:     "file:test.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper config before each test
			viper.Reset()

			viper.Set("dsn.type", tt.dbType)
			viper.Set("dsn.host", tt.host)
			viper.Set("dsn.port", tt.port)
			viper.Set("dsn.user", tt.user)
			viper.Set("dsn.password", tt.password)
			viper.Set("dsn.dbname", tt.dbname)
			viper.Set("dsn.filename", tt.filename)

			got := getDSN()
			if got != tt.want {
				t.Errorf("getDSN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDebugMode(t *testing.T) {
	// Reset debug mode
	debugMode = false

	// Test debug flag
	RootCmd.PersistentFlags().Set("debug", "true")
	initConfig()

	if !debugMode {
		t.Error("Debug mode not set after flag enabled")
	}
}

func TestConfigFileLoading(t *testing.T) {
	// Create temporary config file
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test config
	config := []byte(`
dsn:
  type: "sqlite"
  filename: "test.db"
`)
	if _, err := tmpfile.Write(config); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Set config file and initialize
	oldCfgFile := cfgFile
	cfgFile = tmpfile.Name()
	defer func() { cfgFile = oldCfgFile }()

	initConfig()

	// Verify config was loaded
	if viper.GetString("dsn.type") != "sqlite" {
		t.Errorf("Expected dsn.type to be 'sqlite', got %s", viper.GetString("dsn.type"))
	}
	if viper.GetString("dsn.filename") != "test.db" {
		t.Errorf("Expected dsn.filename to be 'test.db', got %s", viper.GetString("dsn.filename"))
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "sqlite")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	// Initialize config
	initConfig()

	// Verify env var was loaded
	if viper.GetString("dsn.type") != "sqlite" {
		t.Errorf("Expected dsn.type to be 'sqlite', got %s", viper.GetString("dsn.type"))
	}
}

func TestExecute(t *testing.T) {
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"--version"})
	Execute()
	outStr := out.String()
	expectedText := fmt.Sprintf("immotep version %s\n", VERSION)
	if outStr != expectedText {
		t.Errorf("expected: %q; got: %q", expectedText, outStr)
	}
}

func TestExecuteGeocode(t *testing.T) {

	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "unknown")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	// Initialize config
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"geocode"})
	err := RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing geocode command: %v", err)
	}

	RootCmd.SetArgs([]string{"geocode", "1", "2"})
	err = RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing geocode command: %v", err)
	}

}

func TestExecuteLoad(t *testing.T) {

	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "unknown")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	// Initialize config
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"load", "data.csv"})
	err := RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing load command: %v", err)
	}
}

func TestExecuteLoadConf(t *testing.T) {

	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "unknown")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	viper.Set("file.region", "region.geo")
	viper.Set("file.department", "department.geo")
	viper.Set("file.city", "city.geo")
	viper.Set("file.citygeo", "citygeo.geo")

	// Initialize config
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"loadconf"})
	err := RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing loadconf command: %v", err)
	}
}

func TestExecuteCompute(t *testing.T) {

	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "unknown")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	// Initialize config
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"compute"})
	err := RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing compute command: %v", err)
	}
}

func TestExecuteAggregate(t *testing.T) {

	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "unknown")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	// Initialize config
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"aggregate"})
	err := RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing aggregate command: %v", err)
	}
}

func TestExecuteServe(t *testing.T) {

	// Reset viper
	viper.Reset()

	// Set test environment variable
	os.Setenv("IMMOTEP_DSN_TYPE", "unknown")
	defer os.Unsetenv("IMMOTEP_DSN_TYPE")

	// use bad port to force early exit
	os.Setenv("IMMOTEP_SERVE_PORT", "70000")
	defer os.Unsetenv("IMMOTEP_SERVE_PORT")

	// Initialize config
	out := new(bytes.Buffer)
	RootCmd.SetOut(out)
	RootCmd.SetErr(out)
	RootCmd.SetArgs([]string{"serve"})
	err := RootCmd.Execute()
	if err != nil {
		t.Errorf("unexpected error executing serve command: %v", err)
	}
}
