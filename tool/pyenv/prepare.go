package pyenv

import (
	"os"
	"runtime"
)

func Prepare() {
	pyHome := os.Getenv("PYTHONHOME")
	if pyHome == "" {		// use system
		return
	}
	// lib
	switch runtime.GOOS {
	case "darwin":
		libPath := os.Getenv("DYLD_LIBRARY_PATH")
		os.Setenv("DYLD_LIBRARY_PATH", pyHome+"/lib:"+libPath)
	case "linux":
		libPath := os.Getenv("LD_LIBRARY_PATH")
		os.Setenv("LD_LIBRARY_PATH", pyHome+"/lib:"+libPath)
	}
}