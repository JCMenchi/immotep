package main

import (
	"bytes"
	"fmt"
	"testing"

	"jc.org/immotep/cmd"
)

func TestExecute(t *testing.T) {
	out := new(bytes.Buffer)
	cmd.RootCmd.SetOut(out)
	cmd.RootCmd.SetErr(out)
	cmd.RootCmd.SetArgs([]string{"--version"})

	main()

	outStr := out.String()
	expectedText := fmt.Sprintf("immotep version %s\n", cmd.VERSION)
	if outStr != expectedText {
		t.Errorf("expected: %q; got: %q", expectedText, outStr)
	}
}
