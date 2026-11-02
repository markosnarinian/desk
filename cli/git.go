package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var shorthand = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`)

func repositorySource(input string) (string, error) {
	if shorthand.MatchString(input) {
		return "https://github.com/" + strings.TrimSuffix(input, ".git") + ".git", nil
	}
	if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") {
		return filepath.Abs(input)
	}
	if strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "ssh://") || strings.HasPrefix(input, "file://") || strings.HasPrefix(input, "git@") {
		return input, nil
	}
	return "", errors.New("use owner/repo, an HTTPS/SSH Git URL, or an explicit local path")
}

func gitEnv() []string {
	var env []string
	for _, v := range os.Environ() {
		key := strings.SplitN(v, "=", 2)[0]
		// Do not let a parent Git invocation redirect checks to another checkout.
		if strings.HasPrefix(key, "GIT_") && key != "GIT_SSH" && key != "GIT_SSH_COMMAND" && key != "GIT_ASKPASS" {
			continue
		}
		env = append(env, v)
	}
	return append(env, "GIT_OPTIONAL_LOCKS=0", "GIT_NO_REPLACE_OBJECTS=1", "GIT_TERMINAL_PROMPT=0")
}

func git(path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base := []string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "core.ignorecase=false", "-C", path}
	cmd := exec.CommandContext(ctx, "git", append(base, args...)...)
	cmd.Env = gitEnv()
	b, err := cmd.Output()
	return strings.TrimSpace(string(b)), err
}

func (s *store) openRepository(input string) (*workspace, error) {
	source, err := repositorySource(input)
	if err != nil {
		return nil, err
	}
	for i := range s.items {
		w := &s.items[i]
		if w.Kind == "repo" && w.Source == source {
			p, q, err := s.location(w)
			if err != nil {
				return nil, err
			}
			if q {
				return nil, fmt.Errorf("workspace %s is quarantined; restore it first", w.ID)
			}
			if inside, err := git(p, "rev-parse", "--is-inside-work-tree"); err != nil || inside != "true" {
				return nil, fmt.Errorf("checkout %s is incomplete; inspect it using desk list", w.ID)
			}
			return w, nil
		}
	}
	name := strings.TrimSuffix(filepath.Base(source), ".git")
	w, err := s.create(name, "repo", source)
	if err != nil {
		return nil, err
	}
	p := filepath.Join(s.root, "workspaces", w.ID)
	cmd := exec.Command("git", "clone", "--no-local", "--", source, p)
	cmd.Env = gitEnv()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Git may remove its destination on failure. Preserve anything left and
		// make the failed attempt an inspectable scratch workspace, not a clone.
		if makeErr := os.MkdirAll(p, 0700); makeErr != nil {
			return nil, makeErr
		}
		w.Kind = "scratch"
		w.Source = ""
		if saveErr := s.save(); saveErr != nil {
			return nil, saveErr
		}
		return nil, fmt.Errorf("clone failed; workspace %s retained for inspection: %w", w.ID, err)
	}
	return w, nil
}

func (s *store) repositoryRetention(p string) string {
	info, err := os.Lstat(filepath.Join(p, ".git"))
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "not an independent Git clone"
	}
	top, err := git(p, "rev-parse", "--show-toplevel")
	resolved, resolveErr := filepath.EvalSymlinks(p)
	if err != nil || resolveErr != nil || top != resolved {
		return "Git worktree root does not match managed directory"
	}
	for _, file := range []string{"shallow", "objects/info/alternates", "info/grafts"} {
		if _, err := os.Lstat(filepath.Join(p, ".git", file)); !errors.Is(err, os.ErrNotExist) {
			return "shallow or externally dependent Git history"
		}
	}
	worktrees, err := git(p, "worktree", "list", "--porcelain")
	if err != nil || strings.Count(worktrees, "worktree ") != 1 {
		return "additional or uncertain Git worktrees"
	}
	index, err := git(p, "ls-files", "--stage")
	if err != nil {
		return "cannot inspect index"
	}
	for _, line := range strings.Split(index, "\n") {
		if strings.HasPrefix(line, "160000 ") {
			return "submodules require manual retention"
		}
	}
	flags, err := git(p, "ls-files", "-v", "-z")
	if err != nil {
		return "cannot inspect index flags"
	}
	for _, entry := range strings.Split(flags, "\x00") {
		if entry != "" && (entry[0] == 'S' || entry[0] >= 'a' && entry[0] <= 'z') {
			return "skip-worktree or assume-unchanged index entries"
		}
	}
	attributes := []string{"check-attr", "-z", "filter", "--"}
	for _, entry := range strings.Split(flags, "\x00") {
		if len(entry) > 2 {
			attributes = append(attributes, entry[2:])
		}
	}
	if len(attributes) > 4 {
		values, err := git(p, attributes...)
		if err != nil {
			return "cannot inspect tracked content filters"
		}
		fields := strings.Split(values, "\x00")
		for i := 2; i < len(fields); i += 3 {
			if fields[i] != "unspecified" && fields[i] != "unset" {
				return "tracked content filters (including Git LFS) require manual retention"
			}
		}
	}
	status, err := git(p, "status", "--porcelain=v1", "--untracked-files=all", "--ignored=matching", "--ignore-submodules=none")
	if err != nil {
		return "cannot inspect Git status"
	}
	if status != "" {
		return "tracked changes, untracked files, or ignored files"
	}
	origin, err := git(p, "remote", "get-url", "origin")
	if err != nil {
		return "no usable origin"
	}
	local := origin
	if strings.HasPrefix(local, "file://") {
		u, err := url.Parse(local)
		if err != nil || u.Host != "" && u.Host != "localhost" {
			return "uncertain local origin"
		}
		local = u.Path
	}
	if !strings.Contains(local, ":") {
		if !filepath.IsAbs(local) {
			local = filepath.Join(p, local)
		}
		resolved, err := filepath.EvalSymlinks(local)
		if err != nil {
			return "origin unavailable"
		}
		root, err := filepath.EvalSymlinks(s.root)
		if err != nil {
			return "cannot resolve storage root"
		}
		rel, err := filepath.Rel(root, resolved)
		if err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return "origin is inside desk storage, not an independent recovery source"
		}
	}
	refs, err := git(p, "ls-remote", "--heads", "--tags", "origin")
	if err != nil {
		return "origin unavailable; cannot confirm recovery"
	}
	advertised := map[string]string{}
	for _, line := range strings.Split(refs, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			advertised[fields[1]] = fields[0]
		}
	}
	localRefs, err := git(p, "for-each-ref", "--format=%(objectname) %(refname)", "refs/heads", "refs/tags")
	if err != nil {
		return "cannot inspect branch and tag names"
	}
	for _, line := range strings.Split(localRefs, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		remoteHash, exists := advertised[fields[1]]
		if !exists {
			return "local-only branch or tag name"
		}
		if strings.HasPrefix(fields[1], "refs/tags/") && remoteHash != fields[0] {
			return "local tag differs from origin"
		}
	}
	args := []string{"rev-list", "--all", "--reflog", "HEAD", "--not"}
	seen := map[string]bool{}
	for _, line := range strings.Split(refs, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		hash := fields[0]
		if seen[hash] {
			continue
		}
		seen[hash] = true
		if _, err := git(p, "cat-file", "-e", hash+"^{commit}"); err == nil {
			args = append(args, hash)
		}
	}
	if len(args) == 5 {
		return "no locally verifiable commits advertised by origin"
	}
	localOnly, err := git(p, args...)
	if err != nil {
		return "cannot verify complete local history"
	}
	if localOnly != "" {
		return "local-only commits (including refs or reflogs)"
	}
	return ""
}
