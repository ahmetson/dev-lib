package dep_manager

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"github.com/asaskevich/govalidator"
	"github.com/go-git/go-git/v5"
	"github.com/pebbe/zmq4"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// A Dep Manager in the config.DevContext context
type Dep struct {
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
func NewDev(srcPath string, binPath string) (*Dep, error) {
	if err := path.MakeDir(binPath); err != nil {
		return nil, fmt.Errorf("path.MakeDir(%s): %w", binPath, err)
	}

	if err := path.MakeDir(srcPath); err != nil {
		return nil, fmt.Errorf("path.MakeDir(%s): %w", srcPath, err)
	}

	return &Dep{Src: srcPath, Bin: binPath}, nil
}

// Installed checks is the binary exist.
func (dep *Dep) Installed(url string) bool {
	binPath := path.BinPath(dep.Bin, urlToFileName(url))
	exists, _ := path.FileExist(binPath)
	return exists
}

// Install loads the dependency source code, and builds it.
func (dep *Dep) Install(url string, logger *log.Logger) error {
	logger.Info("Starting the installation of the dependency", "url", url)

	// check for a source exist
	srcExist, err := dep.srcExist(url)
	if err != nil {
		return fmt.Errorf("dep_manager.srcExist(%s): %w", url, err)
	}

	logger.Info("Checking the source code", "srcExist", srcExist)

	if srcExist {
		logger.Info("src exists, we need to build it")
		err := dep.build(url, logger)
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}

		return nil
	}

	logger.Info("downloadSrc the source code from remote repository")

	err = dep.downloadSrc(url, logger)
	if err != nil {
		return fmt.Errorf("downloadSrc: %w", err)
	}

	err = dep.build(url, logger)
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func (dep *Dep) srcPath(url string) string {
	return filepath.Join(dep.Src, urlToFileName(url))
}

// srcExist checks is the source code exist or not
func (dep *Dep) srcExist(url string) (bool, error) {
	dataPath := dep.srcPath(url)
	exists, err := path.DirExist(dataPath)
	if err != nil {
		return false, fmt.Errorf("path.DirExists('%s'): %w", dataPath, err)
	}
	return exists, nil
}

// Running checks whether the given client running or not
func (dep *Dep) Running(c *clientConfig.Client) (bool, error) {
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

// builds the application
func (dep *Dep) build(url string, logger *log.Logger) error {
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
func (dep *Dep) Run(url string, id string, parent *clientConfig.Client, logger *log.Logger) error {
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
func (dep *Dep) onEnd(url string, logger *log.Logger) {
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
func (dep *Dep) wait(url string, logger *log.Logger) {
	go func() {
		logger.Info("waiting for dep_manager to end", "dep_manager", url)
		err := dep.cmd.Wait()
		logger.Error("dependency closed itself", "dep_manager", url, "error", err)
		dep.done <- err
	}()
}

// convertToGitUrl converts the url without any protocol schema part into https link to the git.
// It supports only the remote urls.
// The file paths are not supported.
func convertToGitUrl(rawUrl string) (string, error) {
	_, err := url.ParseRequestURI(rawUrl)
	if err == nil {
		return "", fmt.Errorf("url should be not an absolute path")
	}

	absPath := "https://" + rawUrl + ".git"
	fmt.Printf("%s\n", absPath)
	URL, err := url.ParseRequestURI(absPath)
	if err != nil {
		return "", fmt.Errorf("invalid '%s' url: %w", rawUrl, err)
	}

	hostName := URL.Hostname()
	if !govalidator.IsDNSName(hostName) {
		return "", fmt.Errorf("not a valid DNS Name: %s", hostName)
	}

	return absPath, nil
}

// downloadSrc gets the remote source code using Git
func (dep *Dep) downloadSrc(url string, logger *log.Logger) error {
	gitUrl, err := convertToGitUrl(url)
	if err != nil {
		return fmt.Errorf("convertToGitUrl(%s): %w", url, err)
	}
	srcUrl := dep.srcPath(url)
	_, err = git.PlainClone(srcUrl, false, &git.CloneOptions{
		URL:      gitUrl,
		Progress: logger,
	})

	if err != nil {
		return fmt.Errorf("git.PlainClone --url %s --o %s: %w", gitUrl, srcUrl, err)
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
