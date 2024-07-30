package api

import "testing"

func TestServe(t *testing.T) {
	type args struct {
		dsn       string
		staticDir string
		port      int
		debug     bool
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Serve(tt.args.dsn, tt.args.staticDir, tt.args.port, tt.args.debug)
		})
	}
}
