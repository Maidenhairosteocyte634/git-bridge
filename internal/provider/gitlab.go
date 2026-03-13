package provider

import (
	"fmt"
	"net/url"
	"strings"

	"git-bridge/internal/config"
)

type GitLab struct {
	baseURL string
	token   string
}

func NewGitLab(cfg config.ProviderConfig) *GitLab {
	return &GitLab{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		token:   cfg.Credentials["token"],
	}
}

func (g *GitLab) CloneURL(repoPath string) string {
	u, err := url.Parse(g.baseURL)
	if err != nil {
		return fmt.Sprintf("%s/%s.git", g.baseURL, repoPath)
	}
	u.User = url.UserPassword("oauth2", g.token)
	u.Path = "/" + repoPath + ".git"
	return u.String()
}

func (g *GitLab) WebURL(repoPath string) string {
	return fmt.Sprintf("%s/%s", g.baseURL, repoPath)
}

func (g *GitLab) Type() string { return "gitlab" }
