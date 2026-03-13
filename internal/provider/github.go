package provider

import (
	"fmt"

	"git-bridge/internal/config"
)

type GitHub struct {
	token string
}

func NewGitHub(cfg config.ProviderConfig) *GitHub {
	return &GitHub{
		token: cfg.Credentials["token"],
	}
}

func (g *GitHub) CloneURL(repoPath string) string {
	return fmt.Sprintf("https://%s@github.com/%s.git", g.token, repoPath)
}

func (g *GitHub) WebURL(repoPath string) string {
	return fmt.Sprintf("https://github.com/%s", repoPath)
}

func (g *GitHub) Type() string { return "github" }
