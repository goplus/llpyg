package pygen

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"path/filepath"
	"runtime"
)

func prepareEnv(dir string) {
	os.Setenv("PYTHONPATH", dir)
	pyHome := os.Getenv("PYTHONHOME")
	if pyHome == "" {
		return
	}
	// lib
	switch runtime.GOOS {
	case "darwin":
		os.Setenv("DYLD_LIBRARY_PATH", pyHome+"/lib")
	case "linux":
		os.Setenv("LD_LIBRARY_PATH", pyHome+"/lib")
	}
}

func TestGenFunc(t *testing.T) {
	prepareEnv("./testdata/func")
	mod, err := pydump("demo")
	if err != nil {
		t.Fatal(err)
	}
	ctx := createGoPackage(mod)
	for _, sym := range mod.Functions {
		ctx.genFunc(ctx.pkg, sym)
	}
	err = compareWithExpected(t, ctx, "testdata/func/expect.go")
	if err != nil {
		t.Fatalf("test gen func failed: %v", err)
	}
	t.Logf("test gen func pass")
}

func TestGenVar(t *testing.T) {
	prepareEnv("./testdata/var")
	mod, err := pydump("demo")
	if err != nil {
		t.Fatal(err)
	}
	ctx := createGoPackage(mod)
	ctx.genVars(ctx.pkg, mod.Variables)
	err = compareWithExpected(t, ctx, "testdata/var/expect.go")
	if err != nil {
		t.Fatalf("test gen var failed: %v", err)
	}
	t.Logf("test gen var pass")
}

func compareWithExpected(t *testing.T, ctx *context, expectedPath string) error {
	outFilePath := "./temp/actual_git.go"
	dir := filepath.Dir(outFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	ctx.pkg.WriteTo(outFile)
	cmd := exec.Command("go", "fmt", outFilePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go fmt failed: %v", err)
	}
	cmd = exec.Command("git", "diff", "--no-index", outFilePath, expectedPath)
	diffOutput, err := cmd.CombinedOutput()
	if err != nil {
		if len(diffOutput) > 0 {
			return fmt.Errorf("get not match expected: \n%s", string(diffOutput))
		}
		t.Fatal(err)
	}
	return nil
}
