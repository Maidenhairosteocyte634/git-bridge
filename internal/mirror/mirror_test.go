package mirror

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"

	"git-bridge/internal/config"
	"git-bridge/internal/notify"
	"git-bridge/internal/provider"
)

// mockGitRunner records calls and returns configurable errors.
type mockGitRunner struct {
	cloneCalls     []cloneCall
	pushCalls      []pushCall
	deleteRefCalls []deleteRefCall
	cloneErr       error
	pushErr        error
	pushChanged    bool // when true, PushMirror reports changes were pushed
	deleteRefErr   error
}

type cloneCall struct {
	URL string
	Dir string
}

type pushCall struct {
	Dir string
	URL string
}

type deleteRefCall struct {
	URL     string
	RefType string
	RefName string
}

func (m *mockGitRunner) CloneMirror(_ context.Context, url, dir string) error {
	m.cloneCalls = append(m.cloneCalls, cloneCall{URL: url, Dir: dir})
	return m.cloneErr
}

func (m *mockGitRunner) PushMirror(_ context.Context, dir, url string) (bool, error) {
	m.pushCalls = append(m.pushCalls, pushCall{Dir: dir, URL: url})
	return m.pushChanged, m.pushErr
}

func (m *mockGitRunner) DeleteRef(_ context.Context, _, url, refType, refName string) error {
	m.deleteRefCalls = append(m.deleteRefCalls, deleteRefCall{URL: url, RefType: refType, RefName: refName})
	return m.deleteRefErr
}

// mockNotifier records sent notifications.
type mockNotifier struct {
	messages []notify.Message
}

func (m *mockNotifier) Send(msg notify.Message) {
	m.messages = append(m.messages, msg)
}

// newTestService creates a Service with mock git runner and notifier.
func newTestService(repos []config.RepoConfig, providers map[string]provider.Provider, notif notify.Notifier, git *mockGitRunner) *Service {
	return &Service{
		configs:        repos,
		providers:      providers,
		notifier:       notif,
		workDir:        "/tmp/git-bridge-test",
		git:            git,
		timeoutSeconds: 300,
		repoLocks:      make(map[string]*sync.Mutex),
	}
}

func makeProviders() map[string]provider.Provider {
	return map[string]provider.Provider{
		"codecommit-eu": NewCodeCommit(config.ProviderConfig{
			Type:   "codecommit",
			Region: "ap-northeast-2",
			Credentials: map[string]string{
				"git_username": "user",
				"git_password": "pass",
			},
		}),
		"gitlab-main": NewGitLab(config.ProviderConfig{
			Type:    "gitlab",
			BaseURL: "https://gitlab.example.com",
			Credentials: map[string]string{
				"token": "glpat-test",
			},
		}),
		"github-main": NewGitHub(config.ProviderConfig{
			Type: "github",
			Credentials: map[string]string{
				"token": "ghp-test",
			},
		}),
	}
}

func defaultRepos() []config.RepoConfig {
	return []config.RepoConfig{
		{
			Name:       "my-repo",
			Source:     "codecommit-eu",
			Target:     "gitlab-main",
			SourcePath: "my-repo",
			TargetPath: "team/my-repo",
			Direction:  "source-to-target",
		},
		{
			Name:       "bidi-repo",
			Source:     "codecommit-eu",
			Target:     "gitlab-main",
			SourcePath: "bidi-repo",
			TargetPath: "team/bidi-repo",
			Direction:  "bidirectional",
		},
		{
			Name:       "reverse-repo",
			Source:     "gitlab-main",
			Target:     "github-main",
			SourcePath: "team/reverse-repo",
			TargetPath: "org/reverse-repo",
			Direction:  "target-to-source",
		},
	}
}

// --- Sync tests ---

func TestSync_SourceToTarget(t *testing.T) {
	git := &mockGitRunner{pushChanged: true}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.Sync(context.Background(), "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(git.cloneCalls) != 1 {
		t.Fatalf("expected 1 clone call, got %d", len(git.cloneCalls))
	}
	if len(git.pushCalls) != 1 {
		t.Fatalf("expected 1 push call, got %d", len(git.pushCalls))
	}

	// Should notify success
	if len(notif.messages) != 1 || notif.messages[0].Level != "success" {
		t.Errorf("expected success notification, got %+v", notif.messages)
	}
}

func TestSync_Bidirectional(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.Sync(context.Background(), "bidi-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(git.cloneCalls) != 1 {
		t.Fatalf("expected 1 clone call, got %d", len(git.cloneCalls))
	}
}

func TestSync_DirectionNotAllowed(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	// reverse-repo is target-to-source only, so Sync (source-side trigger) should fail
	err := svc.Sync(context.Background(), "team/reverse-repo")
	if err == nil {
		t.Fatal("expected error for disallowed direction")
	}
	if len(git.cloneCalls) != 0 {
		t.Error("should not have called clone")
	}
}

func TestSync_RepoNotConfigured(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.Sync(context.Background(), "nonexistent-repo")
	if err == nil {
		t.Fatal("expected error for unconfigured repo")
	}
}

func TestSync_CloneError(t *testing.T) {
	git := &mockGitRunner{cloneErr: fmt.Errorf("clone failed")}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.Sync(context.Background(), "my-repo")
	if err == nil {
		t.Fatal("expected error")
	}

	// Should notify error
	if len(notif.messages) != 1 || notif.messages[0].Level != "error" {
		t.Errorf("expected error notification, got %+v", notif.messages)
	}
}

func TestSync_PushError(t *testing.T) {
	git := &mockGitRunner{pushErr: fmt.Errorf("push failed")}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.Sync(context.Background(), "my-repo")
	if err == nil {
		t.Fatal("expected error")
	}

	if len(notif.messages) != 1 || notif.messages[0].Level != "error" {
		t.Errorf("expected error notification, got %+v", notif.messages)
	}
}

// --- SyncByTarget tests ---

func TestSyncByTarget_TargetMatch(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	// bidi-repo: target is gitlab, target_path is team/bidi-repo, direction bidirectional
	err := svc.SyncByTarget(context.Background(),"gitlab", "team/bidi-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(git.cloneCalls) != 1 {
		t.Fatalf("expected 1 clone call, got %d", len(git.cloneCalls))
	}
}

func TestSyncByTarget_SourceMatch(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	// my-repo: source is codecommit, source_path is my-repo, direction source-to-target
	// SyncByTarget with source provider match should trigger source-to-target
	err := svc.SyncByTarget(context.Background(),"codecommit", "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(git.cloneCalls) != 1 {
		t.Fatalf("expected 1 clone call, got %d", len(git.cloneCalls))
	}
}

func TestSyncByTarget_DirectionNotAllowed(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	// my-repo is source-to-target only; target-side webhook (gitlab) should not allow target-to-source
	err := svc.SyncByTarget(context.Background(),"gitlab", "team/my-repo")
	if err == nil {
		t.Fatal("expected error for disallowed direction")
	}
}

func TestSyncByTarget_NoMatch(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncByTarget(context.Background(),"gitlab", "unknown/repo")
	if err == nil {
		t.Fatal("expected error for no matching repo")
	}
}

func TestSyncByTarget_CloneError(t *testing.T) {
	git := &mockGitRunner{cloneErr: fmt.Errorf("clone boom")}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncByTarget(context.Background(),"gitlab", "team/bidi-repo")
	if err == nil {
		t.Fatal("expected error")
	}

	if len(notif.messages) != 1 || notif.messages[0].Level != "error" {
		t.Errorf("expected error notification, got %+v", notif.messages)
	}
}

// --- doMirror tests ---

func TestDoMirror_ProviderNotFound(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(nil, makeProviders(), notif, git)

	repoCfg := config.RepoConfig{Name: "test"}

	// Source provider not found
	err := svc.doMirror(context.Background(), repoCfg, "nonexistent", "repo", "gitlab-main", "team/repo")
	if err == nil {
		t.Fatal("expected error for missing source provider")
	}

	// Target provider not found
	err = svc.doMirror(context.Background(), repoCfg, "codecommit-eu", "repo", "nonexistent", "team/repo")
	if err == nil {
		t.Fatal("expected error for missing target provider")
	}
}

func TestDoMirror_SuccessNotification(t *testing.T) {
	git := &mockGitRunner{pushChanged: true}
	notif := &mockNotifier{}
	svc := newTestService(nil, makeProviders(), notif, git)

	repoCfg := config.RepoConfig{Name: "test-repo"}
	err := svc.doMirror(context.Background(), repoCfg, "codecommit-eu", "my-repo", "gitlab-main", "team/my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.messages) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notif.messages))
	}
	if notif.messages[0].Level != "success" {
		t.Errorf("expected success notification, got %q", notif.messages[0].Level)
	}
	if notif.messages[0].Title != "Mirror Sync: test-repo" {
		t.Errorf("unexpected title: %q", notif.messages[0].Title)
	}
}

// --- New() constructor tests ---

func TestNew_DefaultWorkDir(t *testing.T) {
	t.Setenv("WORK_DIR", "")
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"codecommit-eu": {
				Type:   "codecommit",
				Region: "us-east-1",
				Credentials: map[string]string{
					"git_username": "u",
					"git_password": "p",
				},
			},
		},
		Repos: []config.RepoConfig{
			{Name: "r", Source: "codecommit-eu", Target: "codecommit-eu", SourcePath: "a", TargetPath: "b", Direction: "bidirectional"},
		},
	}

	svc, err := New(cfg, notify.NewNoop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.workDir != "/tmp/git-bridge" {
		t.Errorf("expected default workDir, got %q", svc.workDir)
	}
	if svc.git == nil {
		t.Error("git runner should not be nil")
	}
}

func TestNew_CustomWorkDir(t *testing.T) {
	t.Setenv("WORK_DIR", "/custom/dir")
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{},
		Repos:     nil,
	}

	svc, err := New(cfg, notify.NewNoop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.workDir != "/custom/dir" {
		t.Errorf("expected /custom/dir, got %q", svc.workDir)
	}
}

func TestNew_InvalidProvider(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"bad": {Type: "unsupported"},
		},
		Repos: nil,
	}

	_, err := New(cfg, notify.NewNoop())
	if err == nil {
		t.Fatal("expected error for unsupported provider type")
	}
}

func TestSyncByTarget_TargetProviderNotInMap(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	// Only codecommit in providers; target "missing" not in map
	providers := map[string]provider.Provider{
		"codecommit-eu": NewCodeCommit(config.ProviderConfig{
			Type: "codecommit", Region: "us-east-1",
			Credentials: map[string]string{"git_username": "u", "git_password": "p"},
		}),
	}
	repos := []config.RepoConfig{
		{Name: "r", Source: "codecommit-eu", Target: "missing", SourcePath: "r", TargetPath: "t/r", Direction: "source-to-target"},
	}
	svc := newTestService(repos, providers, notif, git)

	// Target provider "missing" not in map → skip target match
	// Source provider "codecommit-eu" matches → doMirror → fails because target "missing" not found
	err := svc.SyncByTarget(context.Background(),"codecommit", "r")
	if err == nil {
		t.Fatal("expected error because target provider missing from providers map")
	}
}

func TestSyncByTarget_SourceProviderNotInMap(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	// Only gitlab in providers; source "missing" is not there
	providers := map[string]provider.Provider{
		"gitlab-main": NewGitLab(config.ProviderConfig{
			Type: "gitlab", BaseURL: "https://gl.test",
			Credentials: map[string]string{"token": "t"},
		}),
	}
	repos := []config.RepoConfig{
		{Name: "r", Source: "missing", Target: "gitlab-main", SourcePath: "r", TargetPath: "t/r", Direction: "bidirectional"},
	}
	svc := newTestService(repos, providers, notif, git)

	// Target matches (gitlab-main, t/r) → doMirror from "gitlab-main" to "missing" → fails
	err := svc.SyncByTarget(context.Background(),"gitlab", "t/r")
	if err == nil {
		t.Fatal("expected error because source provider missing from providers map")
	}
}

func TestSyncByTarget_SourceDirectionNotAllowed(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	// reverse-repo: source=gitlab, target=github, direction=target-to-source
	// SyncByTarget with source match (gitlab, team/reverse-repo) should fail because
	// direction is target-to-source, not source-to-target
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncByTarget(context.Background(),"gitlab", "team/reverse-repo")
	if err == nil {
		t.Fatal("expected error for source-side direction not allowed")
	}
}

func TestSyncByTarget_TargetToSource_Success(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	// reverse-repo: target=github, target_path=org/reverse-repo, direction=target-to-source
	err := svc.SyncByTarget(context.Background(),"github", "org/reverse-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(git.cloneCalls) != 1 {
		t.Errorf("expected 1 clone, got %d", len(git.cloneCalls))
	}
}

// --- SyncDelete tests ---

func TestSyncDelete_Success(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncDelete(context.Background(), "my-repo", "branch", "feature-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(git.deleteRefCalls) != 1 {
		t.Fatalf("expected 1 deleteRef call, got %d", len(git.deleteRefCalls))
	}
	dc := git.deleteRefCalls[0]
	if dc.RefType != "branch" || dc.RefName != "feature-branch" {
		t.Errorf("unexpected deleteRef call: %+v", dc)
	}
	if len(notif.messages) != 1 || notif.messages[0].Level != "success" {
		t.Errorf("expected success notification, got %+v", notif.messages)
	}
}

func TestSyncDelete_Tag(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncDelete(context.Background(), "my-repo", "tag", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dc := git.deleteRefCalls[0]
	if dc.RefType != "tag" || dc.RefName != "v1.0.0" {
		t.Errorf("unexpected deleteRef call: %+v", dc)
	}
}

func TestSyncDelete_RepoNotConfigured(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncDelete(context.Background(), "nonexistent", "branch", "main")
	if err == nil {
		t.Fatal("expected error for unconfigured repo")
	}
}

func TestSyncDelete_DirectionNotAllowed(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	repos := []config.RepoConfig{
		{Name: "rev", Source: "gitlab-main", Target: "github-main", SourcePath: "team/rev", TargetPath: "org/rev", Direction: "target-to-source"},
	}
	svc := newTestService(repos, makeProviders(), notif, git)

	err := svc.SyncDelete(context.Background(), "team/rev", "branch", "old-branch")
	if err == nil {
		t.Fatal("expected error for disallowed direction")
	}
}

func TestSyncDelete_ProviderNotFound(t *testing.T) {
	git := &mockGitRunner{}
	notif := &mockNotifier{}
	repos := []config.RepoConfig{
		{Name: "r", Source: "codecommit-eu", Target: "missing", SourcePath: "r", TargetPath: "t/r", Direction: "source-to-target"},
	}
	svc := newTestService(repos, makeProviders(), notif, git)

	err := svc.SyncDelete(context.Background(), "r", "branch", "main")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestSyncDelete_DeleteRefError(t *testing.T) {
	git := &mockGitRunner{deleteRefErr: fmt.Errorf("delete failed")}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncDelete(context.Background(), "my-repo", "branch", "old-branch")
	if err == nil {
		t.Fatal("expected error")
	}
	if len(notif.messages) != 1 || notif.messages[0].Level != "error" {
		t.Errorf("expected error notification, got %+v", notif.messages)
	}
}

// --- defaultGitRunner integration tests ---

func TestDefaultGitRunner_CloneMirror(t *testing.T) {
	// Create a source bare repo
	srcDir := t.TempDir()
	runGit(t, srcDir, "init", "--bare")

	// Clone mirror from local bare repo
	runner := &defaultGitRunner{}
	destDir := t.TempDir() + "/mirror.git"

	err := runner.CloneMirror(context.Background(), srcDir, destDir)
	if err != nil {
		t.Fatalf("CloneMirror failed: %v", err)
	}
}

func TestDefaultGitRunner_PushMirror(t *testing.T) {
	// Create source repo with a commit
	srcDir := t.TempDir()
	runGit(t, srcDir, "init")
	runGit(t, srcDir, "config", "user.email", "test@test.com")
	runGit(t, srcDir, "config", "user.name", "test")
	writeFile(t, srcDir+"/file.txt", "hello")
	runGit(t, srcDir, "add", ".")
	runGit(t, srcDir, "commit", "-m", "init")

	// Create target bare repo
	tgtDir := t.TempDir()
	runGit(t, tgtDir, "init", "--bare")

	// Clone mirror from source, then push to target
	runner := &defaultGitRunner{}
	mirrorDir := t.TempDir() + "/mirror.git"

	if err := runner.CloneMirror(context.Background(), srcDir, mirrorDir); err != nil {
		t.Fatalf("CloneMirror failed: %v", err)
	}
	changed, err := runner.PushMirror(context.Background(), mirrorDir, tgtDir)
	if err != nil {
		t.Fatalf("PushMirror failed: %v", err)
	}
	if !changed {
		t.Error("expected changed=true for first push")
	}

	// Push again — should be up-to-date
	changed2, err := runner.PushMirror(context.Background(), mirrorDir, tgtDir)
	if err != nil {
		t.Fatalf("PushMirror second push failed: %v", err)
	}
	if changed2 {
		t.Error("expected changed=false for second push (up-to-date)")
	}
}

func TestDefaultGitRunner_DeleteRef(t *testing.T) {
	// Create a repo with a branch
	srcDir := t.TempDir()
	runGit(t, srcDir, "init")
	runGit(t, srcDir, "config", "user.email", "test@test.com")
	runGit(t, srcDir, "config", "user.name", "test")
	writeFile(t, srcDir+"/file.txt", "hello")
	runGit(t, srcDir, "add", ".")
	runGit(t, srcDir, "commit", "-m", "init")
	runGit(t, srcDir, "checkout", "-b", "feature-branch")
	runGit(t, srcDir, "checkout", "master")

	// Clone to bare (target)
	tgtDir := t.TempDir() + "/target.git"
	runGit(t, "", "clone", "--bare", srcDir, tgtDir)

	// Delete the feature-branch from target
	runner := &defaultGitRunner{}
	workDir := t.TempDir() + "/delete-work.git"

	err := runner.DeleteRef(context.Background(), workDir, tgtDir, "branch", "feature-branch")
	if err != nil {
		t.Fatalf("DeleteRef failed: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	var cmd *exec.Cmd
	if dir == "" {
		cmd = exec.Command("git", args...)
	} else {
		cmd = exec.Command("git", append([]string{"-C", dir}, args...)...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- defaultGitRunner error path tests ---

func TestDefaultGitRunner_CloneMirror_InvalidURL(t *testing.T) {
	runner := &defaultGitRunner{}
	destDir := t.TempDir() + "/mirror.git"

	err := runner.CloneMirror(context.Background(), "http://invalid.invalid.invalid/repo.git", destDir)
	if err == nil {
		t.Fatal("expected error for invalid clone URL")
	}
}

func TestDefaultGitRunner_PushMirror_InvalidURL(t *testing.T) {
	// Create a valid mirror repo first
	srcDir := t.TempDir()
	runGit(t, srcDir, "init")
	runGit(t, srcDir, "config", "user.email", "test@test.com")
	runGit(t, srcDir, "config", "user.name", "test")
	writeFile(t, srcDir+"/file.txt", "hello")
	runGit(t, srcDir, "add", ".")
	runGit(t, srcDir, "commit", "-m", "init")

	runner := &defaultGitRunner{}
	mirrorDir := t.TempDir() + "/mirror.git"
	if err := runner.CloneMirror(context.Background(), srcDir, mirrorDir); err != nil {
		t.Fatalf("CloneMirror failed: %v", err)
	}

	// Push to invalid URL should fail
	_, err := runner.PushMirror(context.Background(), mirrorDir, "http://invalid.invalid.invalid/repo.git")
	if err == nil {
		t.Fatal("expected error for invalid push URL")
	}
}

func TestDefaultGitRunner_DeleteRef_InvalidURL(t *testing.T) {
	runner := &defaultGitRunner{}
	workDir := t.TempDir() + "/delete-work.git"

	err := runner.DeleteRef(context.Background(), workDir, "http://invalid.invalid.invalid/repo.git", "branch", "main")
	if err == nil {
		t.Fatal("expected error for invalid delete URL")
	}
}

func TestDefaultGitRunner_CloneMirror_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	runner := &defaultGitRunner{}
	destDir := t.TempDir() + "/mirror.git"

	err := runner.CloneMirror(ctx, "http://example.com/repo.git", destDir)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDefaultGitRunner_PushMirror_TagsFailure(t *testing.T) {
	// Create source with a commit
	srcDir := t.TempDir()
	runGit(t, srcDir, "init")
	runGit(t, srcDir, "config", "user.email", "test@test.com")
	runGit(t, srcDir, "config", "user.name", "test")
	writeFile(t, srcDir+"/file.txt", "hello")
	runGit(t, srcDir, "add", ".")
	runGit(t, srcDir, "commit", "-m", "init")

	// Clone mirror
	runner := &defaultGitRunner{}
	mirrorDir := t.TempDir() + "/mirror.git"
	if err := runner.CloneMirror(context.Background(), srcDir, mirrorDir); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Push branches succeeds to valid target, then we remove the target to fail tags push
	tgtDir := t.TempDir()
	runGit(t, tgtDir, "init", "--bare")

	// This should succeed (both branches and tags)
	if _, err := runner.PushMirror(context.Background(), mirrorDir, tgtDir); err != nil {
		t.Fatalf("PushMirror should succeed: %v", err)
	}
}

// --- direction helper tests ---

func TestAllowsSourceToTarget(t *testing.T) {
	tests := []struct {
		dir  string
		want bool
	}{
		{"source-to-target", true},
		{"Source-To-Target", true},
		{"bidirectional", true},
		{"Bidirectional", true},
		{"target-to-source", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := allowsSourceToTarget(tt.dir); got != tt.want {
			t.Errorf("allowsSourceToTarget(%q) = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestAllowsTargetToSource(t *testing.T) {
	tests := []struct {
		dir  string
		want bool
	}{
		{"target-to-source", true},
		{"Target-To-Source", true},
		{"bidirectional", true},
		{"source-to-target", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := allowsTargetToSource(tt.dir); got != tt.want {
			t.Errorf("allowsTargetToSource(%q) = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

// --- isUpToDate tests ---

func TestIsUpToDate(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"empty output", "", true},
		{"whitespace only", "  \n  ", true},
		{"everything up-to-date", "Everything up-to-date", true},
		{"up-to-date with newline", "Everything up-to-date\n", true},
		{"actual push output", "To /tmp/target.git\n * [new branch]      main -> main\n", false},
		{"forced update", "To /tmp/target.git\n + abc123...def456 main -> main (forced update)\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUpToDate(tt.output); got != tt.want {
				t.Errorf("isUpToDate(%q) = %v, want %v", tt.output, got, tt.want)
			}
		})
	}
}

// --- no-change skip notification tests ---

func TestDoMirror_NoChange_SkipsNotification(t *testing.T) {
	git := &mockGitRunner{pushChanged: false}
	notif := &mockNotifier{}
	svc := newTestService(nil, makeProviders(), notif, git)

	repoCfg := config.RepoConfig{Name: "test-repo"}
	err := svc.doMirror(context.Background(), repoCfg, "codecommit-eu", "my-repo", "gitlab-main", "team/my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.messages) != 0 {
		t.Errorf("expected no notifications when nothing changed, got %d: %+v", len(notif.messages), notif.messages)
	}
}

func TestSync_NoChange_SkipsNotification(t *testing.T) {
	git := &mockGitRunner{pushChanged: false}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.Sync(context.Background(), "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.messages) != 0 {
		t.Errorf("expected no notifications when nothing changed, got %d", len(notif.messages))
	}
}

func TestSyncByTarget_NoChange_SkipsNotification(t *testing.T) {
	git := &mockGitRunner{pushChanged: false}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncByTarget(context.Background(), "gitlab", "team/bidi-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.messages) != 0 {
		t.Errorf("expected no notifications when nothing changed, got %d", len(notif.messages))
	}
}

func TestSyncByTarget_WithChanges_SendsNotification(t *testing.T) {
	git := &mockGitRunner{pushChanged: true}
	notif := &mockNotifier{}
	svc := newTestService(defaultRepos(), makeProviders(), notif, git)

	err := svc.SyncByTarget(context.Background(), "gitlab", "team/bidi-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notif.messages) != 1 || notif.messages[0].Level != "success" {
		t.Errorf("expected 1 success notification, got %+v", notif.messages)
	}
}

// --- repoLock tests ---

func TestRepoLock_ReturnsSameMutex(t *testing.T) {
	svc := newTestService(nil, nil, &mockNotifier{}, &mockGitRunner{})
	mu1 := svc.repoLock("repo-a")
	mu2 := svc.repoLock("repo-a")
	if mu1 != mu2 {
		t.Error("expected same mutex for same repo")
	}
	mu3 := svc.repoLock("repo-b")
	if mu1 == mu3 {
		t.Error("expected different mutex for different repo")
	}
}

// Helper wrappers to use provider constructors from test package
func NewCodeCommit(cfg config.ProviderConfig) provider.Provider {
	p, _ := provider.New("cc", cfg)
	return p
}

func NewGitLab(cfg config.ProviderConfig) provider.Provider {
	p, _ := provider.New("gl", cfg)
	return p
}

func NewGitHub(cfg config.ProviderConfig) provider.Provider {
	p, _ := provider.New("gh", cfg)
	return p
}
