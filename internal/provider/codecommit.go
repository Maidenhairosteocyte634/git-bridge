package provider

import (
	"fmt"
	"net/url"

	"git-bridge/internal/config"
)

type CodeCommit struct {
	region      string
	gitUsername string
	gitPassword string
}

func NewCodeCommit(cfg config.ProviderConfig) *CodeCommit {
	return &CodeCommit{
		region:      cfg.Region,
		gitUsername: cfg.Credentials["git_username"],
		gitPassword: cfg.Credentials["git_password"],
	}
}

func (c *CodeCommit) CloneURL(repoPath string) string {
	user := url.PathEscape(c.gitUsername)
	pass := url.PathEscape(c.gitPassword)
	return fmt.Sprintf("https://%s:%s@git-codecommit.%s.amazonaws.com/v1/repos/%s",
		user, pass, c.region, repoPath)
}

func (c *CodeCommit) WebURL(repoPath string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s/browse", c.region, repoPath)
}

func (c *CodeCommit) Type() string { return "codecommit" }
