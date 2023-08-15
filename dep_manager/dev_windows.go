//go:build windows

package dep_manager

import (
	"path/filepath"
)

func (dep *Dep) binPath(url string) string {
	return filepath.Join(dep.Bin, urlToFileName(url)+".exe")
}
