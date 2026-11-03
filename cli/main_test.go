package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func invoke(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := run(args, &out)
	return strings.TrimSpace(out.String()), err
}

func TestStorageLocation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "other-data"))
	for _, override := range []string{"", filepath.Join(home, "custom"), "relative"} {
		t.Run("DESK_HOME="+override, func(t *testing.T) {
			t.Setenv("DESK_HOME", override)
			id, err := invoke(t, "scratch", "location", "--no-enter")
			if override == "relative" {
				if err == nil || !strings.Contains(err.Error(), "DESK_HOME must be absolute") {
					t.Fatalf("relative override: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			root := override
			if root == "" {
				root = filepath.Join(home, ".desk")
			}
			for _, path := range []string{"registry.json", "registry.lock", "locks", "quarantine", filepath.Join("workspaces", id)} {
				if _, err := os.Stat(filepath.Join(root, path)); err != nil {
					t.Fatal(err)
				}
			}
			listing, err := invoke(t, "list")
			if err != nil || !strings.Contains(listing, filepath.Join(root, "workspaces", id)) {
				t.Fatalf("listing: %s: %v", listing, err)
			}
		})
	}
	if _, err := os.Stat(os.Getenv("XDG_DATA_HOME")); !os.IsNotExist(err) {
		t.Fatalf("XDG_DATA_HOME must not be used: %v", err)
	}
}

func TestScratchLifecycle(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DESK_HOME", root)
	id, err := invoke(t, "scratch", "experiment", "--no-enter")
	if err != nil {
		t.Fatal(err)
	}
	listing, err := invoke(t, "list")
	if err != nil || !strings.Contains(listing, "experiment") {
		t.Fatalf("%s: %v", listing, err)
	}
	if _, err := invoke(t, "remove", id); err == nil {
		t.Fatal("missing acknowledgment accepted")
	}
	p := filepath.Join(root, "workspaces", id)
	if err := os.WriteFile(filepath.Join(p, ".env"), []byte("valuable"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err == nil {
		t.Fatal("nonempty scratch removed")
	}
	if err := os.Remove(filepath.Join(p, ".env")); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "quarantine", id)); err != nil {
		t.Fatal(err)
	}
}

func gitCommand(t *testing.T, path string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", path}, args...)...)
	cmd.Env = append(gitEnv(), "GIT_AUTHOR_NAME=Desk Test", "GIT_AUTHOR_EMAIL=test@example.invalid", "GIT_COMMITTER_NAME=Desk Test", "GIT_COMMITTER_EMAIL=test@example.invalid")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s: %v", args, b, err)
	}
	return strings.TrimSpace(string(b))
}

func repository(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("DESK_HOME", filepath.Join(root, "desk"))
	origin := filepath.Join(root, "origin")
	if err := os.Mkdir(origin, 0700); err != nil {
		t.Fatal(err)
	}
	gitCommand(t, origin, "init", "-b", "main")
	for name, contents := range map[string]string{"README.md": "original\n", ".gitignore": ".env\ndata/\n"} {
		if err := os.WriteFile(filepath.Join(origin, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	gitCommand(t, origin, "add", ".")
	gitCommand(t, origin, "commit", "-m", "initial")
	id, err := invoke(t, "open", origin, "--no-enter")
	if err != nil {
		t.Fatal(err)
	}
	return id, origin
}

func TestRepositoryResume(t *testing.T) {
	id, origin := repository(t)
	again, err := invoke(t, "open", origin, "--no-enter")
	if err != nil || again != id {
		t.Fatalf("resume: %q %v", again, err)
	}
	p := filepath.Join(os.Getenv("DESK_HOME"), "workspaces", id)
	b, err := os.ReadFile(filepath.Join(p, "README.md"))
	if err != nil || string(b) != "original\n" {
		t.Fatalf("clone contents: %q %v", b, err)
	}
	if _, err := os.Stat(filepath.Join(p, ".git", "objects", "info", "alternates")); !os.IsNotExist(err) {
		t.Fatal("clone depends on alternate object store")
	}
}

func TestRepositoryCleanupSafety(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, string, string)
	}{
		{"tracked", func(t *testing.T, p, origin string) {
			os.WriteFile(filepath.Join(p, "README.md"), []byte("edited"), 0600)
		}},
		{"untracked", func(t *testing.T, p, origin string) { os.WriteFile(filepath.Join(p, "notes"), []byte("notes"), 0600) }},
		{"untracked-case-variant", func(t *testing.T, p, origin string) {
			if _, err := os.Stat(filepath.Join(p, "readme.md")); err == nil {
				t.Skip("case-insensitive filesystem")
			}
			gitCommand(t, p, "config", "core.ignorecase", "true")
			os.WriteFile(filepath.Join(p, "readme.md"), []byte("different"), 0600)
		}},
		{"ignored", func(t *testing.T, p, origin string) { os.WriteFile(filepath.Join(p, ".env"), []byte("secret"), 0600) }},
		{"ignored-directory", func(t *testing.T, p, origin string) {
			os.Mkdir(filepath.Join(p, "data"), 0700)
			os.WriteFile(filepath.Join(p, "data", "db"), []byte("data"), 0600)
		}},
		{"unpushed", func(t *testing.T, p, origin string) { gitCommand(t, p, "commit", "--allow-empty", "-m", "local") }},
		{"branch-name-only", func(t *testing.T, p, origin string) { gitCommand(t, p, "branch", "private-name") }},
		{"tag-name-only", func(t *testing.T, p, origin string) { gitCommand(t, p, "tag", "private-name") }},
		{"local-branch", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "switch", "-c", "private")
			gitCommand(t, p, "commit", "--allow-empty", "-m", "private")
			gitCommand(t, p, "switch", "main")
		}},
		{"reflog-only", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "commit", "--allow-empty", "-m", "recoverable")
			gitCommand(t, p, "reset", "--hard", "HEAD~1")
		}},
		{"stash", func(t *testing.T, p, origin string) {
			os.WriteFile(filepath.Join(p, "README.md"), []byte("stashed"), 0600)
			gitCommand(t, p, "stash")
		}},
		{"detached", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "checkout", "--detach")
			gitCommand(t, p, "commit", "--allow-empty", "-m", "detached")
		}},
		{"local-tag", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "checkout", "--detach")
			gitCommand(t, p, "commit", "--allow-empty", "-m", "tagged")
			gitCommand(t, p, "tag", "private")
			gitCommand(t, p, "checkout", "main")
			gitCommand(t, p, "reflog", "expire", "--expire=now", "--all")
		}},
		{"remote-rewritten", func(t *testing.T, p, origin string) {
			gitCommand(t, origin, "checkout", "--orphan", "replacement")
			gitCommand(t, origin, "commit", "--allow-empty", "-m", "replacement")
			gitCommand(t, origin, "branch", "-D", "main")
		}},
		{"offline", func(t *testing.T, p, origin string) { os.Rename(origin, origin+"-unavailable") }},
		{"self-origin", func(t *testing.T, p, origin string) { gitCommand(t, p, "remote", "set-url", "origin", p) }},
		{"redirected-worktree", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "config", "core.worktree", origin)
			os.WriteFile(filepath.Join(p, "private"), []byte("must stay"), 0600)
		}},
		{"assume-unchanged", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "update-index", "--assume-unchanged", "README.md")
			os.WriteFile(filepath.Join(p, "README.md"), []byte("hidden"), 0600)
		}},
		{"skip-worktree", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "update-index", "--skip-worktree", "README.md")
			os.WriteFile(filepath.Join(p, "README.md"), []byte("hidden"), 0600)
		}},
		{"extra-worktree", func(t *testing.T, p, origin string) {
			gitCommand(t, p, "worktree", "add", "--detach", filepath.Join(t.TempDir(), "extra"))
		}},
		{"content-filter", func(t *testing.T, p, origin string) {
			os.WriteFile(filepath.Join(p, ".gitattributes"), []byte("README.md filter=lfs\n"), 0600)
			gitCommand(t, p, "add", ".gitattributes")
			gitCommand(t, p, "commit", "-m", "filter")
			gitCommand(t, origin, "config", "receive.denyCurrentBranch", "updateInstead")
			gitCommand(t, p, "push", "origin", "main")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, origin := repository(t)
			p := filepath.Join(os.Getenv("DESK_HOME"), "workspaces", id)
			tc.change(t, p, origin)
			if _, err := invoke(t, "remove", id, "--quiescent"); err == nil {
				t.Fatal("unsafe workspace quarantined")
			}
			if _, err := invoke(t, "done", id); err != nil {
				t.Fatal(err)
			}
			out, err := invoke(t, "gc", "--apply", "--quiescent")
			if err != nil || !strings.Contains(out, "retained:") {
				t.Fatalf("gc: %s %v", out, err)
			}
			if _, err := os.Stat(p); err != nil {
				t.Fatalf("workspace lost: %v", err)
			}
		})
	}
}

func TestPublishedBranchAndTagCanBeRecovered(t *testing.T) {
	id, _ := repository(t)
	p := filepath.Join(os.Getenv("DESK_HOME"), "workspaces", id)
	gitCommand(t, p, "switch", "-c", "published")
	gitCommand(t, p, "commit", "--allow-empty", "-m", "published")
	gitCommand(t, p, "tag", "-a", "v-test", "-m", "published tag")
	gitCommand(t, p, "push", "origin", "published", "--tags")
	if _, err := invoke(t, "remove", id, "--quiescent"); err != nil {
		t.Fatal(err)
	}
}

func TestFailedAndEmptyClones(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DESK_HOME", filepath.Join(root, "desk"))
	if _, err := invoke(t, "open", filepath.Join(root, "missing"), "--no-enter"); err == nil {
		t.Fatal("nonexistent clone succeeded")
	}
	b, err := os.ReadFile(filepath.Join(root, "desk", "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	var items []workspace
	if err := json.Unmarshal(b, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatal("failed clone not registered")
	}
	if _, err := invoke(t, "remove", items[0].ID, "--quiescent"); err != nil {
		t.Fatal("empty failed attempt cannot be cleaned:", err)
	}
	origin := filepath.Join(root, "empty")
	os.Mkdir(origin, 0700)
	gitCommand(t, origin, "init", "--bare")
	id, err := invoke(t, "open", origin, "--no-enter")
	if err != nil {
		t.Fatal(err)
	}
	again, err := invoke(t, "open", origin, "--no-enter")
	if err != nil || again != id {
		t.Fatalf("empty clone resume: %q %v", again, err)
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err == nil {
		t.Fatal("empty repository without recovery evidence removed")
	}
}

func TestCleanupLifecycle(t *testing.T) {
	id, _ := repository(t)
	root := os.Getenv("DESK_HOME")
	p := filepath.Join(root, "workspaces", id)
	for _, args := range [][]string{{"keep", id}, {"done", id}, {"remove", id, "--quiescent"}} {
		_, err := invoke(t, args...)
		if args[0] == "keep" && err != nil {
			t.Fatal(err)
		}
		if args[0] != "keep" && err == nil {
			t.Fatal("pin bypassed")
		}
	}
	if _, err := invoke(t, "keep", id, "--off"); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "done", id); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "registry.json"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := invoke(t, "gc", "--dry-run")
	if err != nil || !strings.Contains(out, "would quarantine") {
		t.Fatalf("%s %v", out, err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "registry.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("dry run modified registry")
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("dry run moved checkout")
	}
	if _, err := invoke(t, "gc", "--apply"); err == nil {
		t.Fatal("missing acknowledgment accepted")
	}
	if _, err := invoke(t, "gc", "--apply", "--dry-run", "--quiescent"); err == nil {
		t.Fatal("contradictory flags accepted")
	}
	if _, err := invoke(t, "gc", "--apply", "--quiescent"); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "purge", id, "--quiescent"); err == nil {
		t.Fatal("early purge allowed")
	}
	if _, err := invoke(t, "restore", id); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p, "README.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err != nil {
		t.Fatal(err)
	}
	// Advance persisted quarantine age rather than weakening the production clock.
	registry := filepath.Join(root, "registry.json")
	b, _ := os.ReadFile(registry)
	var items []workspace
	if err := json.Unmarshal(b, &items); err != nil {
		t.Fatal(err)
	}
	items[0].Quarantined = time.Now().Add(-8 * 24 * time.Hour)
	b, _ = json.Marshal(items)
	os.WriteFile(registry, b, 0600)
	q := filepath.Join(root, "quarantine", id)
	os.WriteFile(filepath.Join(q, ".env"), []byte("added after quarantine"), 0600)
	if _, err := invoke(t, "purge", id, "--quiescent"); err == nil {
		t.Fatal("purge skipped safety recheck")
	}
	os.Remove(filepath.Join(q, ".env"))
	if _, err := invoke(t, "purge", id, "--quiescent"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(q); !os.IsNotExist(err) {
		t.Fatal("purge left checkout")
	}
	listing, err := invoke(t, "list")
	if err != nil || strings.Contains(listing, id) {
		t.Fatalf("purged entry: %s %v", listing, err)
	}
}

func TestSessionProcess(t *testing.T) {
	if os.Getenv("DESK_TEST_SESSION") == "" {
		return
	}
	if err := run([]string{"enter", os.Getenv("DESK_TEST_SESSION")}, os.Stdout); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestActiveSession(t *testing.T) {
	t.Setenv("DESK_HOME", t.TempDir())
	id, err := invoke(t, "scratch", "--no-enter")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestSessionProcess$")
	cmd.Env = append(os.Environ(), "DESK_TEST_SESSION="+id, "SHELL=/bin/sh")
	in, _ := cmd.StdinPipe()
	out, _ := cmd.StdoutPipe()
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { in.Close(); cmd.Wait() }()
	fmtReady := "printf 'READY\\n'\n"
	io.WriteString(in, fmtReady)
	ready := make(chan bool, 1)
	go func() { scanner := bufio.NewScanner(out); ready <- scanner.Scan() && scanner.Text() == "READY" }()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("shell failed to start")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shell timeout")
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err == nil {
		t.Fatal("active shell quarantined")
	}
	listing, _ := invoke(t, "list")
	if !strings.Contains(listing, "active session") {
		t.Fatal(listing)
	}
	io.WriteString(in, "exit\n")
	in.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceSymlinkRejected(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DESK_HOME", root)
	id, err := invoke(t, "scratch", "--no-enter")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "workspaces", id)
	os.Remove(p)
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "valuable"), []byte("keep"), 0600)
	if err := os.Symlink(outside, p); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke(t, "remove", id, "--quiescent"); err == nil {
		t.Fatal("symlink followed")
	}
	if _, err := os.Stat(filepath.Join(outside, "valuable")); err != nil {
		t.Fatal(err)
	}
}
