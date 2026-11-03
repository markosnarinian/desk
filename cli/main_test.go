package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func invoke(t *testing.T, input string, args ...string) (string, string, error) {
	t.Helper()
	var out, stderr bytes.Buffer
	err := run(args, strings.NewReader(input), &out, &stderr)
	return strings.TrimSpace(out.String()), stderr.String(), err
}

func setup(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("DESK_HOME", root)
	t.Chdir(t.TempDir())
	return filepath.Join(root, "workspaces")
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, b)
	}
}

func TestNewScratchUsesTwoWordName(t *testing.T) {
	dir := setup(t)
	path, _, err := invoke(t, "", "new")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != dir || !regexp.MustCompile(`^[a-z]+-[a-z]+$`).MatchString(filepath.Base(path)) {
		t.Fatalf("unexpected workspace path %s", path)
	}
	if entries, err := os.ReadDir(path); err != nil || len(entries) != 0 {
		t.Fatalf("scratch workspace not empty: %v %v", entries, err)
	}
}

func TestResolveByPrefixAndCurrentDirectory(t *testing.T) {
	dir := setup(t)
	for _, n := range []string{"misty-ember", "misty-river", "quiet-dune"} {
		if err := os.MkdirAll(filepath.Join(dir, n, "src", "deep"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if path, _, err := invoke(t, "", "info", "--path", "q"); err != nil || path != filepath.Join(dir, "quiet-dune") {
		t.Fatalf("unique prefix: %s %v", path, err)
	}
	if _, _, err := invoke(t, "", "info", "misty", "--path"); err == nil || !strings.Contains(err.Error(), "misty-ember, misty-river") {
		t.Fatalf("ambiguous prefix accepted: %v", err)
	}
	if _, _, err := invoke(t, "", "info", "--path"); err == nil {
		t.Fatal("info without NAME outside a workspace succeeded")
	}
	t.Chdir(filepath.Join(dir, "misty-river", "src", "deep"))
	if path, _, err := invoke(t, "", "info", "--path"); err != nil || path != filepath.Join(dir, "misty-river") {
		t.Fatalf("current workspace: %s %v", path, err)
	}
	if list, _, _ := invoke(t, "", "list"); !strings.Contains(list, "* misty-river") || strings.Contains(list, "* misty-ember") {
		t.Fatalf("list does not mark the current workspace:\n%s", list)
	}
}

func TestRemoveAsksFirst(t *testing.T) {
	dir := setup(t)
	path, _, _ := invoke(t, "", "new")
	name := filepath.Base(path)
	if _, _, err := invoke(t, "\n", "rm", name); err == nil {
		t.Fatal("deleted without confirmation")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("workspace gone after declining")
	}
	t.Chdir(path)
	if _, _, err := invoke(t, "y\n", "rm"); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("workspace remains: %v", entries)
	}
}

func TestOpeningARepositoryOffersTasksAndRemembersReuse(t *testing.T) {
	dir := setup(t)
	origin := filepath.Join(t.TempDir(), "origin")
	gitRun(t, ".", "init", "-q", origin)
	gitRun(t, origin, "commit", "-q", "--allow-empty", "-m", "init")
	first, _, err := invoke(t, "", "new", origin)
	if err != nil {
		t.Fatal(err)
	}
	gitRun(t, first, "commit", "-q", "--allow-empty", "-m", "local work")
	second, _, err := invoke(t, "1\n", "new", "file://"+origin+"/")
	if err != nil || second != first {
		t.Fatalf("second open made %s, want %s: %v", second, first, err)
	}
	_, prompt, _ := invoke(t, "n\n", "rm", filepath.Base(first))
	if !strings.Contains(prompt, "1 unpushed commits") {
		t.Fatalf("rm prompt hides unpushed work: %s", prompt)
	}
	if _, _, err := invoke(t, "", "new", filepath.Join(origin, "missing")); err == nil {
		t.Fatal("clone of a missing repository succeeded")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("failed clone left a workspace behind: %v", entries)
	}
	fresh, _, err := invoke(t, "0\n", "new", origin)
	if err != nil || fresh == first {
		t.Fatalf("fresh choice reused %s: %v", fresh, err)
	}
	if got := git(fresh, "log", "-1", "--format=%s"); got != "init" {
		t.Fatalf("fresh clone inherited another task's local changes: %s", got)
	}
	if got, _, err := invoke(t, "", "new", origin, "--reuse"); err != nil || got != fresh {
		t.Fatalf("reuse not most recent: %s %v", got, err)
	}
	if got, _, err := invoke(t, "2\n", "new", origin, "--choose"); err != nil || got != first {
		t.Fatalf("forced chooser: %s %v", got, err)
	}
	if got, _, err := invoke(t, "", "new", origin); err != nil || got != first {
		t.Fatalf("remembered reuse didn't follow last selected task: %s %v", got, err)
	}
	third, _, err := invoke(t, "", "new", origin, "--fresh")
	if err != nil || third == first || third == fresh {
		t.Fatalf("--fresh didn't override preference: %s %v", third, err)
	}
	list, _, err := invoke(t, "", "list")
	if err != nil || strings.Count(list, origin) != 1 || !strings.Contains(list, filepath.Base(third)) || !strings.Contains(list, filepath.Base(first)) {
		t.Fatalf("repository grouping: %s %v", list, err)
	}
}

func TestSourceKeyIgnoresURLSpelling(t *testing.T) {
	short, _ := repository("Octocat/Hello-World")
	same := []string{short, "https://github.com/octocat/hello-world", "https://github.com/octocat/Hello-World.git/",
		"git@github.com:octocat/Hello-World.git", "ssh://git@github.com/octocat/Hello-World"}
	for _, s := range same {
		if sourceKey(s) != sourceKey(short) {
			t.Errorf("%s: %s != %s", s, sourceKey(s), sourceKey(short))
		}
	}
	different := []string{"https://github.com/octocat/Hello-World-2", "https://gitlab.com/octocat/Hello-World", "git@github.com:other/Hello-World"}
	for _, s := range different {
		if sourceKey(s) == sourceKey(short) {
			t.Errorf("%s wrongly matches %s", s, short)
		}
	}
	// Outside GitHub, path case can matter, and local x.git and x are different directories.
	if sourceKey("https://example.com/A/b") == sourceKey("https://example.com/a/b") || sourceKey("/srv/x.git") == sourceKey("/srv/x") {
		t.Error("distinct repositories share a key")
	}
}

func TestInfoAndListContracts(t *testing.T) {
	dir := setup(t)
	path := filepath.Join(dir, "quiet-dune")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"scratch", "git@example.com:owner/repo.git"} {
		if source != "scratch" {
			gitRun(t, path, "init", "-q")
			gitRun(t, path, "remote", "add", "origin", source)
		}
		want := "Name: quiet-dune\nPath: " + path + "\nSource: " + source
		if info, stderr, err := invoke(t, "", "info", "quiet"); err != nil || info != want || stderr != "" {
			t.Fatalf("info = %q, %q, %v; want %q", info, stderr, err, want)
		}
		for _, args := range [][]string{{"info", "quiet", "--path"}, {"info", "--path", "quiet"}} {
			if got, stderr, err := invoke(t, "", args...); err != nil || got != path || stderr != "" || !filepath.IsAbs(got) {
				t.Fatalf("%v = %q, %q, %v", args, got, stderr, err)
			}
		}
		list, _, err := invoke(t, "", "list")
		group := "playgrounds"
		if source != "scratch" {
			group = sourceKey(source)
		}
		if err != nil || !strings.HasPrefix(list, group+"\n") || !strings.Contains(list, "quiet-dune") || !strings.Contains(list, path) {
			t.Fatalf("list groups = %q: %v", list, err)
		}
		t.Chdir(path)
		if got, _, err := invoke(t, "", "info"); err != nil || got != want {
			t.Fatalf("current info = %q %v", got, err)
		}
	}
}

func TestCommandFailures(t *testing.T) {
	setup(t)
	for _, args := range [][]string{{"cd"}, {"path"}, {"info"}, {"info", "missing"}, {"info", "a", "b"}, {"info", "--bad"}, {"info", "--path", "--path"}, {"info", "a", "--path", "b"}, {"new", "--fresh"}, {"new", "o/r", "--bad"}, {"new", "--fresh", "--reuse"}, {"cleanup", "--force"}, {"rename"}, {"finish", "missing"}} {
		if out, _, err := invoke(t, "", args...); err == nil || out != "" {
			t.Fatalf("%v: stdout %q, error %v", args, out, err)
		}
	}
	if out, _, err := invoke(t, "", "switch", "missing"); out != "" || err == nil || !strings.Contains(err.Error(), "desk init") || !strings.Contains(err.Error(), "new terminal") {
		t.Fatalf("switch without integration: %q %v", out, err)
	}
	t.Setenv("DESK_HOME", "relative")
	if _, _, err := invoke(t, "", "info", "--path"); err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("relative DESK_HOME accepted: %v", err)
	}
}
