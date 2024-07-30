package loader

import (
	"reflect"
	"testing"
)

func TestReadZipcodeMap(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		// TODO: Add test cases.
		{"unkown_file", args{filename: "unknown.csv"}, map[string]int{}},
		{"normal", args{filename: "normal.csv"}, map[string]int{"L ABERGEMENT CLEMENCIAT": 1400, "L-ABERGEMENT-CLEMENCIAT": 1400}},
		{"empty", args{filename: "empty.csv"}, map[string]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadZipcodeMap(tt.args.filename); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadZipcodeMap() case[%v] = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_lineCounter(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"normal", args{filename: "normal.csv"}, 2, false},
		{"unkown_file", args{filename: "unknown.csv"}, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lineCounter(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("lineCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("lineCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadRawData(t *testing.T) {
	type args struct {
		dsn      string
		filename string
	}
	tests := []struct {
		name string
		args args
	}{
		{"no_file", args{"file::memory:?cache=shared", "unknown.csv"}},
		{"normal", args{"file::memory:?cache=shared", "valeurs.csv"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoadRawData(tt.args.dsn, tt.args.filename)
		})
	}
}

func Test_checkNotDuplicate(t *testing.T) {
	type args struct {
		previousRow []string
		row         []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkNotDuplicate(tt.args.previousRow, tt.args.row); got != tt.want {
				t.Errorf("checkNotDuplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Use DSN file::memory:?cache=shared to create sqlite DB in memory
func TestLoadRegion(t *testing.T) {
	type args struct {
		dsn      string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no_file", args{"file::memory:?cache=shared", "unknown.json"}, true},
		{"bad_format", args{"file::memory:?cache=shared", "bad_region.geojson"}, true},
		{"bad_prop", args{"file::memory:?cache=shared", "regions_bad_prop.geojson"}, false},
		{"normal", args{"file::memory:?cache=shared", "regions.geojson"}, false},
		{"reload", args{"file::memory:?cache=shared", "regions.geojson"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadRegion(tt.args.dsn, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("LoadRegion() case[%v] error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestLoadDepartment(t *testing.T) {
	type args struct {
		dsn      string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no_file", args{"file::memory:?cache=shared", "unknown.json"}, true},
		{"bad_format", args{"file::memory:?cache=shared", "bad_region.geojson"}, true},
		{"bad_prop", args{"file::memory:?cache=shared", "departements_bad_prop.geojson"}, false},
		{"normal", args{"file::memory:?cache=shared", "departements.geojson"}, false},
		{"reload", args{"file::memory:?cache=shared", "departements.geojson"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadDepartment(tt.args.dsn, tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("LoadDepartment() case[%v] error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestLoadCity(t *testing.T) {
	type args struct {
		dsn         string
		filename    string
		geofilename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no_file", args{"file::memory:?cache=shared", "unknown.json", "communes.geojson"}, true},
		{"bad_format", args{"file::memory:?cache=shared", "bad_region.geojson", "communes.geojson"}, true},
		{"bad_prop", args{"file::memory:?cache=shared", "communes_bad_prop.json", "communes.geojson"}, false},
		{"normal", args{"file::memory:?cache=shared", "communes.json", "communes.geojson"}, false},
		{"reload", args{"file::memory:?cache=shared", "communes.json", "communes.geojson"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadCity(tt.args.dsn, tt.args.filename, tt.args.geofilename); (err != nil) != tt.wantErr {
				t.Errorf("LoadCity() case[%v] error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
