package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func oldWorkspace(t *testing.T, path string, age time.Duration) {
	t.Helper()
	old := time.Now().Add(-age)
	if err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return os.Chtimes(p, old, old)
	}); err != nil {
		t.Fatal(err)
	}
	state := statePath(filepath.Dir(path), filepath.Base(path))
	b, err := os.ReadFile(state)
	if err != nil {
		t.Fatal(err)
	}
	var s workspaceState
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	s.LastUsed = old
	if err := writeJSON(state, s); err != nil {
		t.Fatal(err)
	}
}

func repositoryFixture(t *testing.T) string {
	t.Helper()
	repo := filepath.Join(t.TempDir(), "upstream")
	gitRun(t, ".", "init", "-q", repo)
	mustWrite(t, filepath.Join(repo, "code.txt"), "original\n")
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "initial")
	return repo
}

func TestCleanupProtectsLocalWorkAndActivity(t *testing.T) {
	for _, scenario := range []string{"clean", "recent-visit", "saved-editor-change", "untracked", "ignored-local-file", "finished-repository", "unpushed-branch", "detached-commit", "local-tag", "stash", "nested-repository", "scratch", "finished-scratch", "recent-file", "open-file", "cwd", "live-follow", "unmanaged", "git-error"} {
		t.Run(scenario, func(t *testing.T) {
			dir := setup(t)
			input := repositoryFixture(t)
			if scenario == "scratch" || scenario == "finished-scratch" {
				input = ""
			}
			args := []string{"new"}
			if input != "" {
				args = append(args, input)
			}
			path, _, err := invoke(t, "", args...)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Base(path)
			var opened *os.File
			switch scenario {
			case "saved-editor-change":
				mustWrite(t, filepath.Join(path, "code.txt"), "editor changes\n")
			case "untracked":
				mustWrite(t, filepath.Join(path, "notes.txt"), "keep me\n")
			case "ignored-local-file", "finished-repository":
				mustWrite(t, filepath.Join(path, ".git", "info", "exclude"), ".env\n")
				mustWrite(t, filepath.Join(path, ".env"), "LOCAL_ONLY=keep\n")
				if scenario == "finished-repository" {
					if _, _, err := invoke(t, "", "finish", name); err != nil {
						t.Fatal(err)
					}
				}
			case "unpushed-branch":
				gitRun(t, path, "switch", "-qc", "other-task")
				gitRun(t, path, "commit", "-qm", "local", "--allow-empty")
				gitRun(t, path, "switch", "-")
			case "detached-commit":
				gitRun(t, path, "checkout", "--detach", "-q")
				gitRun(t, path, "commit", "-qm", "detached local", "--allow-empty")
			case "local-tag":
				gitRun(t, path, "checkout", "--detach", "-q")
				gitRun(t, path, "commit", "-qm", "tagged local", "--allow-empty")
				gitRun(t, path, "tag", "experiment")
				gitRun(t, path, "checkout", "-q", "-")
			case "stash":
				mustWrite(t, filepath.Join(path, "code.txt"), "stashed changes\n")
				gitRun(t, path, "stash", "push", "-qm", "keep")
			case "nested-repository":
				mustWrite(t, filepath.Join(path, ".git", "info", "exclude"), "nested/\n")
				gitRun(t, path, "clone", "-q", input, "nested")
				gitRun(t, filepath.Join(path, "nested"), "commit", "-qm", "local nested", "--allow-empty")
			case "finished-scratch":
				mustWrite(t, filepath.Join(path, "prototype.txt"), "throwaway\n")
				if _, _, err := invoke(t, "", "finish", name); err != nil {
					t.Fatal(err)
				}
			}
			oldWorkspace(t, path, 73*time.Hour)
			switch scenario {
			case "recent-visit":
				if _, _, err := invoke(t, "", "_visit", name); err != nil {
					t.Fatal(err)
				}
			case "recent-file":
				// A clean repository can still have been used recently in an editor.
				now := time.Now()
				if err := os.Chtimes(filepath.Join(path, "code.txt"), now, now); err != nil {
					t.Fatal(err)
				}
			case "open-file":
				f, err := os.Open(filepath.Join(path, "code.txt"))
				if err != nil {
					t.Fatal(err)
				}
				opened = f
				defer f.Close()
			case "cwd":
				t.Chdir(path)
			case "live-follow":
				mustWrite(t, filepath.Join(filepath.Dir(dir), "state", name+".follow"), strconv.Itoa(os.Getpid()))
			case "unmanaged":
				if err := os.Remove(statePath(dir, name)); err != nil {
					t.Fatal(err)
				}
			case "git-error":
				t.Setenv("PATH", t.TempDir())
			}
			out, stderr, err := invoke(t, "", "cleanup", "--yes")
			if err != nil || out != "" {
				t.Fatalf("cleanup: %q %s %v", out, stderr, err)
			}
			if opened != nil {
				opened.Close()
				if _, _, err := invoke(t, "", "cleanup", "--yes"); err != nil {
					t.Fatal(err)
				}
			}
			_, err = os.Stat(path)
			removed := os.IsNotExist(err)
			wantRemoved := scenario == "clean" || scenario == "finished-scratch"
			if removed != wantRemoved {
				t.Fatalf("removed=%v, want %v; %s", removed, wantRemoved, stderr)
			}
			if removed {
				entries, err := os.ReadDir(filepath.Join(filepath.Dir(dir), "garbage"))
				if err != nil || len(entries) != 1 {
					t.Fatalf("workspace not retained: %v %v", entries, err)
				}
				if list, _, _ := invoke(t, "", "list"); strings.Contains(list, name) {
					t.Fatalf("garbage still listed: %s", list)
				}
			}
		})
	}
}

type onRead struct {
	fn func()
}

func (r onRead) Read(p []byte) (int, error) {
	r.fn()
	return copy(p, "y\n"), io.EOF
}

func TestCleanupRequiresConfirmationAndRechecksWork(t *testing.T) {
	setup(t)
	path, _, err := invoke(t, "", "new", repositoryFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	oldWorkspace(t, path, 73*time.Hour)
	if _, _, err := invoke(t, "n\n", "cleanup"); err == nil {
		t.Fatal("declining cleanup succeeded")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("declined workspace removed")
	}
	var stderr bytes.Buffer
	err = run([]string{"cleanup"}, onRead{func() {
		mustWrite(t, filepath.Join(path, "code.txt"), "resumed during confirmation\n")
	}}, io.Discard, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(path, "code.txt")); err != nil || string(b) != "resumed during confirmation\n" {
		t.Fatalf("resumed work lost: %s %v", b, err)
	}
}

func TestGarbageRetentionStartsAtRemovalAndCollectionIsSilent(t *testing.T) {
	dir := setup(t)
	path, _, err := invoke(t, "", "new")
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(path, "keep.txt"), "recoverable")
	oldWorkspace(t, path, 1000*time.Hour)
	if _, _, err := invoke(t, "y\n", "rm", filepath.Base(path)); err != nil {
		t.Fatal(err)
	}
	garbage := filepath.Join(filepath.Dir(dir), "garbage")
	entries, err := os.ReadDir(garbage)
	if err != nil || len(entries) != 1 {
		t.Fatalf("garbage: %v %v", entries, err)
	}
	entry := filepath.Join(garbage, entries[0].Name())
	if out, stderr, err := invoke(t, "", "version"); err != nil || out != version || stderr != "" {
		t.Fatalf("GC leaks output: %q %q %v", out, stderr, err)
	}
	if b, err := os.ReadFile(filepath.Join(entry, "workspace", "keep.txt")); err != nil || string(b) != "recoverable" {
		t.Fatalf("old workspace wasn't retained from removal: %s %v", b, err)
	}
	for _, age := range []time.Duration{71 * time.Hour, 73 * time.Hour} {
		if err := writeJSON(filepath.Join(entry, "deleted.json"), time.Now().Add(-age)); err != nil {
			t.Fatal(err)
		}
		if _, stderr, err := invoke(t, "", "help"); err != nil || stderr != "" {
			t.Fatalf("help collection: %s %v", stderr, err)
		}
		_, err := os.Stat(entry)
		if os.IsNotExist(err) != (age == 73*time.Hour) {
			t.Fatalf("wrong retention boundary at %s: %v", age, err)
		}
	}
	unknown := filepath.Join(garbage, "unknown")
	mustWrite(t, filepath.Join(unknown, "deleted.json"), "broken")
	outside := t.TempDir()
	mustWrite(t, filepath.Join(outside, "keep.txt"), "outside")
	if err := os.Symlink(outside, filepath.Join(garbage, "linked")); err != nil {
		t.Fatal(err)
	}
	invoke(t, "", "version")
	if _, err := os.Stat(unknown); err != nil {
		t.Fatal("unknown garbage deleted")
	}
	if _, err := os.Stat(filepath.Join(outside, "keep.txt")); err != nil {
		t.Fatal("collection followed a symlink")
	}
}

func TestRenamePreservesWorkAndFinishedState(t *testing.T) {
	dir := setup(t)
	path, _, err := invoke(t, "", "new")
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(path, "prototype.txt"), "my prototype")
	if _, _, err := invoke(t, "", "finish", filepath.Base(path)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(path)
	newPath, _, err := invoke(t, "", "rename", "explore-parser")
	if err != nil || newPath != filepath.Join(dir, "explore-parser") {
		t.Fatalf("rename: %s %v", newPath, err)
	}
	if b, err := os.ReadFile(filepath.Join(newPath, "prototype.txt")); err != nil || string(b) != "my prototype" {
		t.Fatalf("rename lost work: %s %v", b, err)
	}
	list, _, _ := invoke(t, "", "list")
	if !strings.Contains(list, "explore-parser [finished]") || strings.Contains(list, filepath.Base(path)) {
		t.Fatalf("rename lost state: %s", list)
	}
	for _, invalid := range []string{"../escape", ".hidden", "explore-parser", "has space", "a/b"} {
		if _, _, err := invoke(t, "", "rename", "explore-parser", invalid); err == nil {
			t.Fatalf("accepted invalid or conflicting name %q", invalid)
		}
	}
}
