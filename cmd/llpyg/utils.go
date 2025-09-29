package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func createFileWithDirs(filePath string) (*os.File, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.Create(filePath)
}

func writeConfig(cfg Config, outDir string) error {
	cfgPath := filepath.Join(outDir, "llpyg.cfg")
	file, err := createFileWithDirs(cfgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create file: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("error: failed to write configuration: %w", err)
	}
	return nil
}

func initGoModule(modName string, outDir string) error {
	// init go module
	cmd := exec.Command("go", "mod", "init", modName)
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: failed to initialize Go module: %w", err)
	}

	cmd = exec.Command("go", "get", "github.com/goplus/lib/py")
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: failed to get github.com/goplus/lib/py: %w", err)
	}
	return nil
}

func goModTidy(outDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: failed to tidy Go module: %w", err)
	}
	return nil
}

func codeFormat(outDir string) error {
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: failed to format Go code: %w", err)
	}
	return nil
}
