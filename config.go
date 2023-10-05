package context

import (
	"fmt"
	configClient "github.com/ahmetson/config-lib/client"
	"github.com/ahmetson/os-lib/path"
	"path/filepath"
)

// Specifically for Dev Context
const (
	// SrcKey is the path of source directory from the configuration
	SrcKey = "SERVICE_DEPS_SRC"
	// BinKey is the path of bin directory from the configuration
	BinKey = "SERVICE_DEPS_BIN"
)

// SetDevDefaults sets the required developer context's parameters in the configuration engine.
//
// It sets the source manager's bin path and source path in (dot is current dir by executable):
//
//		/bin.exe
//		/_sds/source/
//		/_sds/bin/
//	 /_sds/source/github.com.ahmetson.proxy-lib/main.go
//	 /_sds/bin/github.com.ahmetson.proxy-lib.exe
func SetDevDefaults(engine configClient.Interface) error {
	currentDir, err := path.CurrentDir()
	if err != nil {
		return fmt.Errorf("path.CurrentDir: %w", err)
	}

	srcPath := filepath.Join(currentDir, "_sds", "src")
	binPath := filepath.Join(currentDir, "_sds", "bin")

	if err := engine.SetDefault(SrcKey, srcPath); err != nil {
		return fmt.Errorf("configClient.SetDefault('%s', '%s'): %w", SrcKey, srcPath, err)
	}
	if err := engine.SetDefault(BinKey, binPath); err != nil {
		return fmt.Errorf("configClient.SetDefault('%s', '%s'): %w", BinKey, binPath, err)
	}

	return nil
}
