package provider

import (
	"strings"
	"testing"

	"git-bridge/internal/config"
)

func TestCodeCommit_CloneURL(t *testing.T) {
	p := NewCodeCommit(config.ProviderConfig{
		Region: "eu-central-1",
		Credentials: map[string]string{
			"git_username": "user-at-123",
			"git_password": "pass123",
		},
	})

	url := p.CloneURL("my-repo")

	if !strings.Contains(url, "git-codecommit.eu-central-1.amazonaws.com") {
		t.Errorf("URL missing codecommit host: %s", url)
	}
	if !strings.Contains(url, "/v1/repos/my-repo") {
		t.Errorf("URL missing repo path: %s", url)
	}
	if !strings.Contains(url, "user-at-123") {
		t.Errorf("URL missing username: %s", url)
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("URL should start with https://: %s", url)
	}
	if p.Type() != "codecommit" {
		t.Errorf("type = %q, want codecommit", p.Type())
	}
}

func TestCodeCommit_CloneURL_SpecialChars(t *testing.T) {
	p := NewCodeCommit(config.ProviderConfig{
		Region: "eu-central-1",
		Credentials: map[string]string{
			"git_username": "user@123",
			"git_password": "pass/with=chars",
		},
	})

	url := p.CloneURL("repo")
	// @ in username should be encoded, / in password should be encoded
	if strings.Contains(url, "pass/with") {
		t.Errorf("slash in password should be encoded: %s", url)
	}
}

func TestGitLab_CloneURL(t *testing.T) {
	p := NewGitLab(config.ProviderConfig{
		BaseURL: "http://gitlab.example.com",
		Credentials: map[string]string{
			"token": "glpat-test123",
		},
	})

	url := p.CloneURL("server/my-repo")

	if !strings.Contains(url, "oauth2:glpat-test123@") {
		t.Errorf("URL missing oauth2 token: %s", url)
	}
	if !strings.Contains(url, "gitlab.example.com/server/my-repo.git") {
		t.Errorf("URL missing repo path: %s", url)
	}
	if !strings.HasPrefix(url, "http://") {
		t.Errorf("URL should start with http://: %s", url)
	}
	if p.Type() != "gitlab" {
		t.Errorf("type = %q, want gitlab", p.Type())
	}
}

func TestGitLab_CloneURL_HTTPS(t *testing.T) {
	p := NewGitLab(config.ProviderConfig{
		BaseURL: "https://gitlab.com",
		Credentials: map[string]string{
			"token": "glpat-xyz",
		},
	})

	url := p.CloneURL("org/repo")
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("URL should start with https://: %s", url)
	}
}

func TestGitLab_CloneURL_TrailingSlash(t *testing.T) {
	p := NewGitLab(config.ProviderConfig{
		BaseURL: "http://gitlab.example.com/",
		Credentials: map[string]string{
			"token": "tok",
		},
	})

	url := p.CloneURL("team/repo")
	if strings.Contains(url, "//team") {
		t.Errorf("URL has double slash: %s", url)
	}
}

func TestGitHub_CloneURL(t *testing.T) {
	p := NewGitHub(config.ProviderConfig{
		Credentials: map[string]string{
			"token": "ghp_test123",
		},
	})

	url := p.CloneURL("org/my-repo")

	if !strings.Contains(url, "ghp_test123@github.com") {
		t.Errorf("URL missing token: %s", url)
	}
	if !strings.Contains(url, "/org/my-repo.git") {
		t.Errorf("URL missing repo path: %s", url)
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("URL should start with https://: %s", url)
	}
	if p.Type() != "github" {
		t.Errorf("type = %q, want github", p.Type())
	}
}

func TestNew_ValidProviders(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.ProviderConfig
		wantType string
	}{
		{
			name:     "codecommit",
			cfg:      config.ProviderConfig{Type: "codecommit", Region: "us-east-1", Credentials: map[string]string{}},
			wantType: "codecommit",
		},
		{
			name:     "gitlab",
			cfg:      config.ProviderConfig{Type: "gitlab", BaseURL: "http://gl.test", Credentials: map[string]string{}},
			wantType: "gitlab",
		},
		{
			name:     "github",
			cfg:      config.ProviderConfig{Type: "github", Credentials: map[string]string{}},
			wantType: "github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.name, tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Type() != tt.wantType {
				t.Errorf("type = %q, want %q", p.Type(), tt.wantType)
			}
		})
	}
}

// --- WebURL tests ---

func TestCodeCommit_WebURL(t *testing.T) {
	p := NewCodeCommit(config.ProviderConfig{
		Region:      "eu-central-1",
		Credentials: map[string]string{"git_username": "u", "git_password": "p"},
	})

	url := p.WebURL("my-repo")
	if !strings.Contains(url, "eu-central-1.console.aws.amazon.com") {
		t.Errorf("URL missing region console host: %s", url)
	}
	if !strings.Contains(url, "/repositories/my-repo/browse") {
		t.Errorf("URL missing repo browse path: %s", url)
	}
}

func TestGitLab_WebURL(t *testing.T) {
	p := NewGitLab(config.ProviderConfig{
		BaseURL:     "https://gitlab.example.com",
		Credentials: map[string]string{"token": "t"},
	})

	url := p.WebURL("team/my-repo")
	if url != "https://gitlab.example.com/team/my-repo" {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestGitHub_WebURL(t *testing.T) {
	p := NewGitHub(config.ProviderConfig{
		Credentials: map[string]string{"token": "t"},
	})

	url := p.WebURL("org/my-repo")
	if url != "https://github.com/org/my-repo" {
		t.Errorf("unexpected URL: %s", url)
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New("test", config.ProviderConfig{Type: "bitbucket"})
	if err == nil {
		t.Fatal("expected error for unknown provider type")
	}
}

func TestGitLab_CloneURL_InvalidURL(t *testing.T) {
	p := &GitLab{
		baseURL: "://invalid",
		token:   "tok",
	}
	url := p.CloneURL("repo")
	if !strings.Contains(url, "://invalid/repo.git") {
		t.Errorf("expected fallback URL, got %s", url)
	}
}
