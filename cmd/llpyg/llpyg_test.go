package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		runMode  string
		wantArgs Args
	}{
		{
			name:    "basic_cmd_mode",
			args:    []string{"numpy"},
			runMode: "cmd",
			wantArgs: Args{
				OutputDir: "./out",
				ModName:   "",
				ModDepth:  1,
				Kwarg:     "numpy",
			},
		},
		{
			name:    "cmd_mode_with_options",
			args:    []string{"-o", "test", "-mod", "numpy", "-d", "2", "numpy"},
			runMode: "cmd",
			wantArgs: Args{
				OutputDir: "test",
				ModName:   "numpy",
				ModDepth:  2,
				Kwarg:     "numpy",
			},
		},
		{
			name:    "cfg_mode",
			args:    []string{"-o", "test", "-mod", "numpy", "llpyg.cfg"},
			runMode: "cfg",
			wantArgs: Args{
				OutputDir: "test",
				ModName:   "numpy",
				ModDepth:  1,
				Kwarg:     "llpyg.cfg",
			},
		},
		{
			name:    "default_values",
			args:    []string{"pandas"},
			runMode: "cmd",
			wantArgs: Args{
				OutputDir: "./out",
				ModName:   "",
				ModDepth:  1,
				Kwarg:     "pandas",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			originalFlagSet := flag.CommandLine
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			defer func() {
				flag.CommandLine = originalFlagSet
			}()
			originalArgs := os.Args
			defer func() {
				os.Args = originalArgs
			}()
			os.Args = append([]string{"llpyg"}, c.args...)
			runMode, args := parseArgs()
			if runMode != c.runMode {
				t.Errorf("runMode = %v, want %v", runMode, c.runMode)
			}
			assertArgsEqual(t, args, c.wantArgs)
		})
	}
}

func assertArgsEqual(t *testing.T, got, want Args) {
	t.Helper()

	gotAbs, err := filepath.Abs(got.OutputDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for got.OutputDir %q: %v", got.OutputDir, err)
	}
	wantAbs, err := filepath.Abs(want.OutputDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for want.OutputDir %q: %v", want.OutputDir, err)
	}

	if gotAbs != wantAbs {
		t.Errorf("unexpected OutputDir: got %q, want %q", gotAbs, wantAbs)
	}
	if got.ModName != want.ModName {
		t.Errorf("unexpected ModName: got %q, want %q", got.ModName, want.ModName)
	}
	if got.ModDepth != want.ModDepth {
		t.Errorf("unexpected ModDepth: got %d, want %d", got.ModDepth, want.ModDepth)
	}
	if got.Kwarg != want.Kwarg {
		t.Errorf("unexpected Kwarg: got %q, want %q", got.Kwarg, want.Kwarg)
	}
}


func TestGoModuleUtils(t *testing.T) {
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "test.go")
	goContent := `package main

	import (
		"fmt"
		"github.com/goplus/lib/py"
	)

	func main() {
	a := py.Object{}
	fmt.Printf("hello %v", a)
	}
	`
	err := os.WriteFile(goFile, []byte(goContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}
	err = initGoModule("test", tempDir)
	if err != nil {
		t.Fatal(err)
	}
	err = goModTidy(tempDir)
	if err != nil {
		t.Log(err)
	}
	err = codeFormat(tempDir)
	if err != nil {
		t.Log(err)
	}
}
