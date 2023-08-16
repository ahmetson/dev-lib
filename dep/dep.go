// Package dep defines the dependency parameters.
// It's used by dependency_manager to articulate with the dependencies.
package dep

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/url"
)

// The Src struct is used to fetch the source code.
// It has the optional Branch option.
// When the Branch is set, then the dependency manager will check out that from remote.
type Src struct {
	Url    string
	GitUrl string
	Branch string // Branch to fetch. Leave it empty to get the certain branch.
}

// New dependency by its source code
func New(url string) (*Src, error) {
	gitUrl, err := convertToGitUrl(url)
	if err != nil {
		return nil, fmt.Errorf("convertToGitUrl('%s'): %w", url, err)
	}

	return &Src{Url: url, GitUrl: gitUrl}, nil
}

// SetBranch sets the branch name of the repository.
func (src *Src) SetBranch(branch string) {
	src.Branch = branch
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
