package github

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5"
)

// RepoInfo contains repository owner and name
type RepoInfo struct {
	Owner string
	Repo  string
}

// GetCurrentRepoInfo detects the current repository from git remote
func GetCurrentRepoInfo() (*RepoInfo, error) {
	// Open the git repository in current directory
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get the remote named 'origin'
	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, fmt.Errorf("failed to get origin remote: %w", err)
	}

	if len(remote.Config().URLs) == 0 {
		return nil, fmt.Errorf("no URLs found for origin remote")
	}

	remoteURL := remote.Config().URLs[0]
	return parseGitHubURL(remoteURL)
}

// parseGitHubURL parses a GitHub URL and extracts owner and repo name
func parseGitHubURL(rawURL string) (*RepoInfo, error) {
	// Handle SSH URLs like git@github.com:owner/repo.git
	if after, ok := strings.CutPrefix(rawURL, "git@github.com:"); ok {
		path := after
		return parseRepoPath(path)
	}

	// Handle HTTPS URLs
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Host != "github.com" {
		return nil, fmt.Errorf("not a GitHub URL: %s", rawURL)
	}

	return parseRepoPath(u.Path)
}

// parseRepoPath parses a GitHub repository path like /owner/repo.git
func parseRepoPath(path string) (*RepoInfo, error) {
	// Remove leading slash and .git suffix
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, ".git")

	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository path: %s", path)
	}

	return &RepoInfo{
		Owner: parts[0],
		Repo:  parts[1],
	}, nil
}

// GetRepoInfoFromArgs gets repository info from command line arguments or current directory
func GetRepoInfoFromArgs(args []string) (*RepoInfo, error) {
	if len(args) >= 1 {
		// Parse from argument like "owner/repo"
		parts := strings.Split(args[0], "/")
		if len(parts) == 2 {
			return &RepoInfo{
				Owner: parts[0],
				Repo:  parts[1],
			}, nil
		}
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", args[0])
	}

	// Try to detect from current directory
	return GetCurrentRepoInfo()
}
