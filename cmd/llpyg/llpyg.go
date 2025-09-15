package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/goplus/lib/py"
	"github.com/goplus/llpyg/tool/pyenv"
	"github.com/goplus/llpyg/tool/pygen"
)

type Args struct {
	OutputDir string
	ModName   string
	ModDepth  int
	Kwarg     string	// llpyg.cfg or pythonLibName
}

type Config struct {
	Name    string   `json:"name"`    // go module name
	LibName string   `json:"libName"` // Python library name
	Modules []string `json:"modules"` // Python modules
}

type library struct {
	LibName 	string   	`json:"libName"`
	LibVersion 	string 		`json:"libVersion"`
	Depth   	int      	`json:"depth"`
	Modules 	[]string 	`json:"modules"`
}

func main() {
	var cfg Config

	// parse args
	runMode, args := parseArgs()

	// prepare python env
	pyenv.Prepare()

	// get config
	switch runMode {
	case "cmd":
		cfg = genConfig(args)
	case "cfg":
		cfg = readConfig(args.Kwarg)   		// cfgPath
	}

	// init work dir
	initWorkDir(&args, cfg)

	// LLGo Bindings generation
	generateFromConfig(cfg, args.OutputDir)

	// tidy go module
	goModTidy(args.OutputDir)

	// tidy go code
	goCodeTidy(args.OutputDir)

	fmt.Printf("LLGo bindings generated successfully in %s\n", args.OutputDir)
}

// parse args from command line
func parseArgs() (runMode string, args Args) {
	output := flag.String("o", "./test", "Output dir")
	modName := flag.String("mod", "", "Generate Go Bindings module name")
	modDepth := flag.Int("d", 1, "Extract module depth")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Input error: Usage")
		fmt.Fprintln(os.Stderr, "  llpyg [-o outputDir] [-mod modName] [-d modDepth] pythonLibName")
		fmt.Fprintln(os.Stderr, "  llpyg [-o outputDir] [-mod modName] llpyg.cfg")
		os.Exit(1)
	}
	absOutput, err := filepath.Abs(*output)
	if err != nil {
		log.Fatalf("error: failed to resolve output path '%s': %v\n", *output, err)
	}
	args = Args{
		OutputDir: absOutput,
		ModName:   *modName,
		ModDepth:  *modDepth,
		Kwarg:     flag.Arg(0),		// pythonLibName or cfgPath
	}
	if strings.HasSuffix(args.Kwarg, ".cfg") {
		return "cfg", args
	}
	return "cmd", args
}

// get modules info from pymodule
func genConfig(args Args) (cfg Config) {
	lib, err := pymodule(args.Kwarg, args.ModDepth)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s is ready\n", lib.LibName, lib.LibVersion)
	cfg = Config{
		Name:    lib.Modules[0], // go package name
		LibName: lib.LibName,
		Modules: lib.Modules,
	}
	return cfg
}

func pymodule(libName string, depth int) (lib library, err error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("pymodule", "-d", strconv.Itoa(depth), libName)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return lib, fmt.Errorf("get modules from %s failed: %s", libName, stderr.String())
	}
	err = json.Unmarshal(stdout.Bytes(), &lib)
	if err != nil {
		return lib, fmt.Errorf("unmarshal %s failed: %s", libName, stderr.String())
	}
	return lib, nil
}

func readConfig(cfgPath string) (cfg Config) {
	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		log.Fatalf("error: failed to open config file %s: %v\n", cfgPath, err)
	}
	defer cfgFile.Close()
	decoder := json.NewDecoder(cfgFile)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalf("error: failed to decode config file %s: %v\n", cfgPath, err)
	}
	return cfg
}

// init work dir, include go module, llpyg.cfg
func initWorkDir(args *Args, cfg Config) {
	args.OutputDir = filepath.Join(args.OutputDir, cfg.Name)
	// remove origin output dir
	if err := os.RemoveAll(args.OutputDir); err != nil {
		log.Fatalf("error: failed to remove output directory %s: %v\n", args.OutputDir, err)
	}
	// write config file
	if err := writeConfig(cfg, args.OutputDir); err != nil {
		log.Fatalf("error: failed to write config file %s: %v\n", args.OutputDir, err)
	}
	// init go module
	if args.ModName == "" {
		args.ModName = cfg.Name
	}
	if err := initGoModule(args.ModName, args.OutputDir); err != nil {
		log.Fatal(err)
	}
}

func generateFromConfig(cfg Config, outDir string) {
	for _, moduleName := range cfg.Modules {
		fmt.Printf("Generating LLGo bindings for %s...\n", moduleName)
		outFilePath := filepath.Join(outDir, moduleToPath(moduleName))
		file, err := createFileWithDirs(outFilePath)
		if err != nil {
			log.Fatalf("error: failed to create file %s: %v\n", outFilePath, err)
		}
		defer file.Close()
		pygen.GenLLGoBindings(moduleName, file)
	}
}

// module name to file path
func moduleToPath(moduleName string) string {
	parts := strings.Split(moduleName, ".")
	fileName := parts[len(parts)-1] + ".go"
	parts = parts[1:]
	path := strings.Join(parts, "/")
	return filepath.Join(path, fileName)
}

