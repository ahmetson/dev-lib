// Package dep_manager tracks the dependency manager in the local context.
package dep_manager

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/dev-lib/source"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pebbe/zmq4"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// A DepManager Manager in the config.DevContext context
type DepManager struct {
	cmd     map[string]*exec.Cmd
	done    map[string]chan error
	exitErr error

	Src string `json:"SERVICE_DEPS_SRC"`
	Bin string `json:"SERVICE_DEPS_BIN"`

	parent *clientConfig.Client
}

// New source manager in the Dev context.
//
// It will prepare the directories for source codes and binary.
// If preparation fails, it will throw an error.
func New() *DepManager {
	return &DepManager{
		Src:  "",
		Bin:  "",
		cmd:  make(map[string]*exec.Cmd, 0),
		done: make(map[string]chan error, 0),
	}
}

func (dep *DepManager) SetPaths(srcPath string, binPath string) error {
	if err := path.MakeDir(binPath); err != nil {
		return fmt.Errorf("path.MakeDir(%s): %w", binPath, err)
	}
	if err := path.MakeDir(srcPath); err != nil {
		return fmt.Errorf("path.MakeDir(%s): %w", srcPath, err)
	}

	dep.Src = srcPath
	dep.Bin = binPath

	return nil
}

// Close the dependency
func (dep *DepManager) Close(c *clientConfig.Client) error {
	sock, err := client.New(c)
	if err != nil {
		return fmt.Errorf("zmq.NewSocket: %w", err)
	}

	closeRequest := &message.Request{
		Command:    "close",
		Parameters: key_value.Empty(),
	}
	reply, err := sock.Request(closeRequest)
	if err != nil {
		return fmt.Errorf("socket.Request('close'): %w", err)
	}

	if !reply.IsOK() {
		return fmt.Errorf("dependency replied: %s", reply.Message)
	}

	err = sock.Close()
	if err != nil {
		return fmt.Errorf("socket.Close: %w", err)
	}

	return nil
}

// Installed checks is the binary exist.
func (dep *DepManager) Installed(url string) bool {
	binPath := path.BinPath(dep.Bin, urlToFileName(url))
	exists, _ := path.FileExist(binPath)
	return exists
}

// Install loads the dependency source code, and builds it.
func (dep *DepManager) Install(src *source.Src, parent *log.Logger) error {
	logger := parent.Child("install", "srcUrl", src.Url)
	// check for a source exist
	srcExist, err := dep.srcExist(src.Url)
	if err != nil {
		return fmt.Errorf("dep_manager.srcExist(%s): %w", src.Url, err)
	}

	if !srcExist {
		err = dep.downloadSrc(src, logger)
		if err != nil {
			return fmt.Errorf("downloadSrc: %w", err)
		}
	}

	err = dep.build(src.Url, logger)
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func (dep *DepManager) srcPath(url string) string {
	return filepath.Join(dep.Src, urlToFileName(url))
}

// srcExist checks is the source code exist or not
func (dep *DepManager) srcExist(url string) (bool, error) {
	dataPath := dep.srcPath(url)
	exists, err := path.DirExist(dataPath)
	if err != nil {
		return false, fmt.Errorf("path.DirExists('%s'): %w", dataPath, err)
	}
	return exists, nil
}

// Running checks whether the given client running or not.
// If the service is running on another process or on another node,
// then that service should expose the port.
func (dep *DepManager) Running(c *clientConfig.Client) (bool, error) {
	depUrl := clientConfig.Url(c)

	sock, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		return false, fmt.Errorf("zmq.NewSocket: %w", err)
	}
	bindErr := sock.Bind(depUrl)

	if bindErr != nil {
		return true, nil
	}

	err = sock.Close()
	if err != nil {
		return false, fmt.Errorf("socket.Close: %w", err)
	}

	// if bind error, then its running
	// if nil bind error, then it's not running
	return false, nil
}

// build the application from source code.
func (dep *DepManager) build(url string, logger *log.Logger) error {
	srcUrl := dep.srcPath(url)
	binUrl := path.BinPath(dep.Bin, urlToFileName(url))

	err := cleanBuild(srcUrl, logger)
	if err != nil {
		return fmt.Errorf("cleanBuild(%s): %w", srcUrl, err)
	}

	cmd := exec.Command("go", "build", "-o", binUrl)
	cmd.Stdout = logger.Child("build", "binUrl", binUrl)
	cmd.Dir = srcUrl
	cmd.Stderr = logger.Child("buildErr", "binUrl", binUrl)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}
	return nil
}

// Run runs the binary. If the binary isn't running, then it will return an error.
func (dep *DepManager) Run(url string, id string, parent *clientConfig.Client) error {
	binUrl := path.BinPath(dep.Bin, urlToFileName(url))
	configFlag := fmt.Sprintf("--url=%s", url)
	idFlag := fmt.Sprintf("--id=%s", id)
	parentFlag := fmt.Sprintf("--parent=%s", clientConfig.Url(parent))

	args := []string{configFlag, idFlag, parentFlag}

	dep.exitErr = nil
	dep.done[id] = make(chan error, 1)
	dep.onStop(id, dep.done[id])

	logger, err := log.New(id, false)
	if err != nil {
		return fmt.Errorf("log.New('%s'): %w", id, err)
	}
	errLogger, err := log.New(id+"Err", false)
	if err != nil {
		return fmt.Errorf("log.New('%sErr'): %w", id, err)
	}

	cmd := exec.Command(binUrl, args...)
	cmd.Stdout = logger
	cmd.Stderr = errLogger
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}
	dep.cmd[id] = cmd
	dep.wait(id)

	return nil
}

// onStop invoked when the dependency stops. It cleans out the dependency parameters.
func (dep *DepManager) onStop(id string, errChan chan error) {
	go func() {
		err := <-errChan
		dep.exitErr = err
		delete(dep.cmd, id)

	}()
}

// wait until the dependency stops
func (dep *DepManager) wait(id string) {
	go func() {
		dep.done[id] <- dep.cmd[id].Wait()
	}()
}

// downloadSrc gets the remote source code using Git
func (dep *DepManager) downloadSrc(src *source.Src, logger *log.Logger) error {
	srcUrl := dep.srcPath(src.Url)

	options := &git.CloneOptions{
		URL:      src.GitUrl,
		Progress: logger.Child("download"),
	}

	if len(src.Branch) > 0 {
		options.ReferenceName = plumbing.NewBranchReferenceName(src.Branch)
	}

	_, err := git.PlainClone(srcUrl, false, options)

	if err != nil {
		return fmt.Errorf("git.PlainClone --url %s --o %s: %w", src.Url, srcUrl, err)
	}

	return nil
}

// deleteSrc deletes the source code
func (dep *DepManager) deleteSrc(url string) error {
	srcUrl := dep.srcPath(url)

	err := os.RemoveAll(srcUrl)
	if err != nil {
		return fmt.Errorf("os.RemoveAll('%s'): %s", srcUrl, err)
	}

	return nil
}

// deleteBin deletes the binary from the directory.
// If there is no binary, it will throw an error.
// If attempt to delete failed, it will throw an error.
func (dep *DepManager) deleteBin(url string) error {
	if !dep.Installed(url) {
		return fmt.Errorf("'%s' not installed", url)
	}

	binPath := path.BinPath(dep.Bin, urlToFileName(url))
	if err := os.Remove(binPath); err != nil {
		return fmt.Errorf("os.Remove('%s'): %w", binPath, err)
	}

	return nil
}

// Uninstall deletes the dependency source code, and its binary.
// Trying to uninstall already running application will fail.
//
// Uninstall will omit if no binary or source code exists.
func (dep *DepManager) Uninstall(src *source.Src) error {
	exist, err := dep.srcExist(src.Url)
	if err != nil {
		return fmt.Errorf("dep_manager.exist(%s): %w", src.Url, err)
	}

	if exist {
		err := dep.deleteSrc(src.Url)
		if err != nil {
			return fmt.Errorf("source.deleteSrc: %w", err)
		}
	}

	exist = dep.Installed(src.Url)
	if exist {
		err := dep.deleteBin(src.Url)
		if err != nil {
			return fmt.Errorf("source.deleteBin('%s'): %w", src.Url, err)
		}
	}

	return nil
}

// calls `go mod tidy`
func cleanBuild(srcUrl string, logger *log.Logger) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = logger.Child("clean")
	cmd.Dir = srcUrl
	cmd.Stderr = logger.Child("cleanErr")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}

	return nil
}

// urlToFileName converts the given url to the file name. Simply it replaces the slashes with dots.
//
// Url returns the full url to connect to the orchestra.
//
// The orchestra url is defined from the main service's url.
//
// For example:
//
//	serviceUrl = "github.com/ahmetson/sample-service"
//	contextUrl = "orchestra.github.com.ahmetson.sample-service"
//
// This controllerName is set as the handler's name in the config.
// Then the handler package will generate an inproc:// url based on the handler name.
func urlToFileName(url string) string {
	str := strings.ReplaceAll(strings.ReplaceAll(url, "/", "."), "\\", ".")
	return regexp.MustCompile(`[^a-zA-Z0-9-_.]+`).ReplaceAllString(str, "")
}
