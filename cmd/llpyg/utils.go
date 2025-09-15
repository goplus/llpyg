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
	if err := os.Chdir(outDir); err != nil {
		return fmt.Errorf("error: failed to change directory: %w", err)
	}
	// init go module
	cmd := exec.Command("go", "mod", "init", modName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error: failed to initialize Go module: %w", err)
	}

	getCmd := exec.Command("go", "get", "github.com/goplus/lib/py")
	getCmd.Stdout = os.Stdout
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		return fmt.Errorf("error: failed to get github.com/goplus/lib/py: %w", err)
	}
	return nil
}

func goModTidy(outDir string) error {
	if err := os.Chdir(outDir); err != nil {
		return fmt.Errorf("error: failed to change directory: %w", err)
	}
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("error: failed to tidy Go module: %w", err)
	}
	return nil
}


func goCodeTidy(outDir string) error {
	if err := os.Chdir(outDir); err != nil {
		return fmt.Errorf("error: failed to change directory: %w", err)
	}
	tidyCmd := exec.Command("go", "fmt", "./...")
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("error: failed to tidy Go code: %w", err)
	}
	return nil
}
