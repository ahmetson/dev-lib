package config

import (
	"fmt"
	configLib "github.com/ahmetson/config-lib"
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
// It sets the dep manager's bin path and source path in (dot is current dir by executable):
//
//	./_sds/src
//	./_sds/bin
func SetDevDefaults(engine configLib.Interface) error {
	currentDir, err := path.CurrentDir()
	if err != nil {
		return fmt.Errorf("path.CurrentDir: %w", err)
	}

	srcPath := filepath.Join(currentDir, "_sds", "src")
	binPath := filepath.Join(currentDir, "_sds", "bin")

	engine.SetDefault(SrcKey, srcPath)
	engine.SetDefault(BinKey, binPath)

	return nil
}
