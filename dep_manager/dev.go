package dep_manager

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/dev-lib/dep"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pebbe/zmq4"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// A DepManager Manager in the config.DevContext context
type DepManager struct {
	cmd  *exec.Cmd
	done chan error

	Src string `json:"SERVICE_DEPS_SRC"`
	Bin string `json:"SERVICE_DEPS_BIN"`

	parent *clientConfig.Client
}

// NewDev creates the dep manager in the Dev context.
//
// It will prepare the directories for source codes and binary.
// If preparation fails, it will throw an error.
func NewDev(srcPath string, binPath string) (*DepManager, error) {
	if err := path.MakeDir(binPath); err != nil {
		return nil, fmt.Errorf("path.MakeDir(%s): %w", binPath, err)
	}

	if err := path.MakeDir(srcPath); err != nil {
		return nil, fmt.Errorf("path.MakeDir(%s): %w", srcPath, err)
	}

	return &DepManager{Src: srcPath, Bin: binPath}, nil
}

// Installed checks is the binary exist.
func (dep *DepManager) Installed(url string) bool {
	binPath := path.BinPath(dep.Bin, urlToFileName(url))
	exists, _ := path.FileExist(binPath)
	return exists
}

// Install loads the dependency source code, and builds it.
func (dep *DepManager) Install(srcUrl *dep.Src, logger *log.Logger) error {
	// check for a source exist
	srcExist, err := dep.srcExist(srcUrl.Url)
	if err != nil {
		return fmt.Errorf("dep_manager.srcExist(%s): %w", srcUrl.Url, err)
	}

	logger.Info("Checking the source code", "srcExist", srcExist)

	if srcExist {
		logger.Info("src exists, we need to build it")
		err := dep.build(srcUrl.Url, logger)
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}

		return nil
	}

	logger.Info("downloadSrc the source code from remote repository")

	err = dep.downloadSrc(srcUrl, logger)
	if err != nil {
		return fmt.Errorf("downloadSrc: %w", err)
	}

	err = dep.build(srcUrl.Url, logger)
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

// Running checks whether the given client running or not
func (dep *DepManager) Running(c *clientConfig.Client) (bool, error) {
	depUrl := client.ClientUrl(c.Id, c.Port)

	sock, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		return false, fmt.Errorf("zmq.NewSocket: %w", err)
	}
	bindErr := sock.Bind(depUrl)
	err = sock.Close()
	if err != nil {
		return false, fmt.Errorf("socket.Close: %w", err)
	}

	// if bind error, then its running
	// if nil bind error, then it's not running
	return bindErr != nil, nil
}

// build the application from source code.
func (dep *DepManager) build(url string, logger *log.Logger) error {
	srcUrl := dep.srcPath(url)
	binUrl := path.BinPath(dep.Bin, urlToFileName(url))

	logger.Info("building", "src", srcUrl, "bin", binUrl)

	err := cleanBuild(srcUrl, logger)
	if err != nil {
		return fmt.Errorf("cleanBuild(%s): %w", srcUrl, err)
	}

	cmd := exec.Command("go", "build", "-o", binUrl)
	cmd.Stdout = logger
	cmd.Dir = srcUrl
	cmd.Stderr = logger
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("cmd.Run: %w", err)
	}
	return nil
}

// Run downloads the binary if it wasn't.
func (dep *DepManager) Run(url string, id string, parent *clientConfig.Client, logger *log.Logger) error {
	binUrl := path.BinPath(dep.Bin, urlToFileName(url))
	configFlag := fmt.Sprintf("--url=%s", url)
	idFlag := fmt.Sprintf("--id=%s", id)
	parentFlag := fmt.Sprintf("--parent=%s", client.ClientUrl(parent.Id, parent.Port))

	args := []string{configFlag, idFlag, parentFlag}

	logger.Info("running", "command", binUrl, "arguments", args)

	dep.done = make(chan error, 1)
	dep.onEnd(url, logger)

	cmd := exec.Command(binUrl, args...)
	cmd.Stdout = logger
	cmd.Stderr = logger
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}
	dep.cmd = cmd
	dep.wait(url, logger)

	return nil
}

// Call it before starting the dependency with os/exec.Start
func (dep *DepManager) onEnd(url string, logger *log.Logger) {
	go func() {
		err := <-dep.done
		if err != nil {
			logger.Error("dependency ended with error", "error", err, "dep_manager", url)
		} else {
			logger.Info("dependency ended successfully", "dep_manager", url)
		}
		dep.cmd = nil
	}()
}

// wait until the dependency is not exiting
func (dep *DepManager) wait(url string, logger *log.Logger) {
	go func() {
		logger.Info("waiting for dep_manager to end", "dep_manager", url)
		err := dep.cmd.Wait()
		logger.Error("dependency closed itself", "dep_manager", url, "error", err)
		dep.done <- err
	}()
}

// downloadSrc gets the remote source code using Git
func (dep *DepManager) downloadSrc(src *dep.Src, logger *log.Logger) error {
	srcUrl := dep.srcPath(src.Url)

	options := &git.CloneOptions{
		URL:      src.GitUrl,
		Progress: logger,
	}

	if len(src.Branch) > 0 {
		options.ReferenceName = plumbing.NewBranchReferenceName(src.Branch)
	}

	repo, err := git.PlainClone(srcUrl, false, options)

	if err != nil {
		return fmt.Errorf("git.PlainClone --url %s --o %s: %w", src.Url, srcUrl, err)
	}


	return nil
}
	return nil
}

// calls `go mod tidy`
func cleanBuild(srcUrl string, logger *log.Logger) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = logger
	cmd.Dir = srcUrl
	cmd.Stderr = logger
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("cmd.Run: %w", err)
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
