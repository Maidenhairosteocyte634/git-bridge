package provider

import (
	"fmt"

	"git-bridge/internal/config"
)

// Provider builds clone URLs for a git provider.
type Provider interface {
	// CloneURL returns the HTTPS clone URL for a repo path.
	CloneURL(repoPath string) string
	// WebURL returns the browser URL for a repo path (no credentials).
	WebURL(repoPath string) string
	// Type returns the provider type name.
	Type() string
}

// New creates a Provider from config.
func New(name string, cfg config.ProviderConfig) (Provider, error) {
	switch cfg.Type {
	case "codecommit":
		return NewCodeCommit(cfg), nil
	case "gitlab":
		return NewGitLab(cfg), nil
	case "github":
		return NewGitHub(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider type %q for %q", cfg.Type, name)
	}
}
