//go:build !windows
// +build !windows

package dep

import (
	"path/filepath"
)

// Full file path by the given name
func (dep *Dep) binPath(url string) string {
	return filepath.Join(dep.Bin, urlToFileName(url))
}
