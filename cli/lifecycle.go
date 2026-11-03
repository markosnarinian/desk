//go:build linux || darwin

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const retention = 72 * time.Hour

// State stays outside the workspace, so a new playground is genuinely empty.
type workspaceState struct {
	LastUsed   time.Time `json:"last_used"`
	Disposable bool      `json:"disposable,omitempty"`
}

func statePath(dir, name string) string {
	return filepath.Join(filepath.Dir(dir), "state", name+".json")
}

func readState(dir, name string) (workspaceState, error) {
	var s workspaceState
	b, err := os.ReadFile(statePath(dir, name))
	if err == nil {
		err = json.Unmarshal(b, &s)
	}
	return s, err
}

func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".desk-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func recordUse(dir, name string) error {
	return lifecycleLock(dir, func() error { return writeUse(dir, name) })
}

// The caller holds the lifecycle lock while updating or moving this state.
func writeUse(dir, name string) error {
	if info, err := os.Lstat(filepath.Join(dir, name)); err != nil || !info.IsDir() {
		return errors.New("workspace no longer exists")
	}
	s, err := readState(dir, name)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.LastUsed = time.Now().UTC()
	return writeJSON(statePath(dir, name), s)
}

// Serialize moves and collection, but never hold a lock while asking a user or
// following an Amp thread. Independent commands must remain usable then.
func lifecycleLock(dir string, fn func() error) error {
	root := filepath.Dir(dir)
	if err := os.MkdirAll(root, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(root, ".lifecycle.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

// The caller holds the lifecycle lock across its safety check and this move.
func trashWorkspace(dir, name string) error {
	garbage := filepath.Join(filepath.Dir(dir), "garbage")
	if err := os.MkdirAll(garbage, 0700); err != nil {
		return err
	}
	entry, err := os.MkdirTemp(garbage, name+"-")
	if err != nil {
		return err
	}
	// Write the timestamp before moving. A failed move leaves no garbage.
	if err := writeJSON(filepath.Join(entry, "deleted.json"), time.Now().UTC()); err != nil {
		os.RemoveAll(entry)
		return err
	}
	if err := os.Rename(filepath.Join(dir, name), filepath.Join(entry, "workspace")); err != nil {
		os.RemoveAll(entry)
		return err
	}
	os.Remove(statePath(dir, name))
	return nil
}

// Collection is deliberately silent and only recognizes Desk garbage entries.
// Missing/corrupt timestamps are not permission to delete data.
func collectGarbage(dir string, now time.Time) {
	garbage := filepath.Join(filepath.Dir(dir), "garbage")
	if _, err := os.Stat(garbage); err != nil {
		return
	}
	_ = lifecycleLock(dir, func() error {
		entries, err := os.ReadDir(garbage)
		if err != nil {
			return err
		}
		cwd, _ := os.Getwd()
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(garbage, e.Name())
			b, err := os.ReadFile(filepath.Join(path, "deleted.json"))
			var deleted time.Time
			if err != nil || json.Unmarshal(b, &deleted) != nil || deleted.IsZero() || now.Before(deleted.Add(retention)) {
				continue
			}
			if cwd != "" && within(path, cwd) {
				continue
			}
			_ = os.RemoveAll(path)
		}
		return nil
	})
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func checkedGit(path string, args ...string) (string, error) {
	cmd := gitCommand(path, args...)
	// A safety inspection must not refresh the index and look like new work.
	cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0")
	b, err := cmd.Output()
	return strings.TrimSpace(string(b)), err
}

// Fail closed. Every local branch must be represented by a remote-tracking
// ref, and stashes, ignored nested repositories and submodules need care too.
func safeRepository(path string) bool {
	checks := [][]string{
		{"status", "--porcelain", "--untracked-files=all", "--ignore-submodules=none"},
		{"stash", "list"},
		{"rev-list", "HEAD", "--branches", "--tags", "--not", "--remotes"},
		{"ls-files", "--others", "--ignored", "--exclude-standard"},
	}
	for _, args := range checks {
		result, err := checkedGit(path, args...)
		if err != nil || result != "" {
			return false
		}
	}
	remote, err := checkedGit(path, "config", "--get", "remote.origin.url")
	return err == nil && remote != ""
}

func followActive(dir, name string) bool {
	b, err := os.ReadFile(filepath.Join(filepath.Dir(dir), "state", name+".follow"))
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		return true
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return true
	}
	err = syscall.Kill(pid, 0)
	return err == nil || !errors.Is(err, syscall.ESRCH)
}

func cleanupEligible(dir, name string, now time.Time, busy []string) bool {
	s, err := readState(dir, name)
	if err != nil || s.LastUsed.IsZero() || now.Before(s.LastUsed.Add(retention)) || followActive(dir, name) {
		return false
	}
	path := filepath.Join(dir, name)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	for _, p := range busy {
		if within(path, p) {
			return false
		}
	}
	if origin(path) == "" && !s.Disposable {
		return false
	}
	safe := true
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if now.Before(info.ModTime().Add(retention)) {
			safe = false
			return fs.SkipAll
		}
		// Inspect every repository, even one in an ignored playground folder.
		if d.Name() == ".git" {
			if !safeRepository(filepath.Dir(p)) {
				safe = false
				return fs.SkipAll
			}
		}
		return nil
	})
	return err == nil && safe
}

func cleanup(dir string, yes bool, in io.Reader, stderr io.Writer) error {
	now := time.Now()
	busy, err := busyPaths()
	if err != nil {
		return fmt.Errorf("cannot check running workspaces: %w", err)
	}
	all, err := names(dir)
	if err != nil {
		return err
	}
	var candidates []string
	for _, name := range all {
		// Remember observed process activity, so closing an editor or server
		// doesn't make a long-running workspace immediately stale.
		if _, err := readState(dir, name); err == nil {
			path := filepath.Join(dir, name)
			if resolved, err := filepath.EvalSymlinks(path); err == nil {
				path = resolved
			}
			for _, p := range busy {
				if within(path, p) {
					if err := recordUse(dir, name); err != nil {
						return err
					}
					break
				}
			}
		}
		if cleanupEligible(dir, name, now, busy) {
			candidates = append(candidates, name)
			fmt.Fprintf(stderr, "  %s — unused for at least 72h; no local-only work (or explicitly finished scratch)\n", name)
		}
	}
	if len(candidates) == 0 {
		fmt.Fprintln(stderr, "desk: no workspaces eligible for cleanup")
		return nil
	}
	if !yes {
		fmt.Fprintf(stderr, "Move these %d workspace(s) to garbage for 72 hours? [y/N] ", len(candidates))
		answer, _ := bufio.NewReader(in).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			return errors.New("nothing removed")
		}
	}
	for _, name := range candidates {
		// A user may have resumed work while the confirmation was open.
		err := lifecycleLock(dir, func() error {
			busy, err := busyPaths()
			if err != nil {
				return err
			}
			if !cleanupEligible(dir, name, time.Now(), busy) {
				fmt.Fprintf(stderr, "desk: kept %s; activity or local work changed\n", name)
				return nil
			}
			if err := trashWorkspace(dir, name); err != nil {
				return err
			}
			fmt.Fprintf(stderr, "desk: moved %s to garbage (retained for 72 hours)\n", name)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func renameWorkspace(dir, name, newName string) (string, error) {
	if !workspaceName.MatchString(newName) || strings.HasPrefix(newName, ".") {
		return "", errors.New("use a name starting with a letter or number, followed by letters, numbers, hyphens or underscores")
	}
	path := filepath.Join(dir, newName)
	err := lifecycleLock(dir, func() error {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("workspace %q already exists", newName)
		}
		if followActive(dir, name) {
			return errors.New("stop live follow before renaming this workspace")
		}
		if err := os.Rename(filepath.Join(dir, name), path); err != nil {
			return err
		}
		if err := os.Rename(statePath(dir, name), statePath(dir, newName)); err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = os.Rename(path, filepath.Join(dir, name))
			return err
		}
		return writeUse(dir, newName)
	})
	return path, err
}

type preferences struct {
	Reuse map[string]bool `json:"reuse_repositories"`
}

func loadPreferences(dir string) (preferences, error) {
	p := preferences{Reuse: map[string]bool{}}
	b, err := os.ReadFile(filepath.Join(filepath.Dir(dir), "preferences.json"))
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err == nil {
		err = json.Unmarshal(b, &p)
	}
	if p.Reuse == nil {
		p.Reuse = map[string]bool{}
	}
	return p, err
}

func openWorkspace(dir string, args []string, in io.Reader, stderr io.Writer) (string, error) {
	input, mode := "", ""
	for _, a := range args {
		switch a {
		case "--fresh", "--choose", "--reuse":
			if mode != "" {
				return "", errors.New("choose only one of --fresh, --choose, --reuse")
			}
			mode = a
		default:
			if strings.HasPrefix(a, "-") || input != "" {
				return "", errors.New("use desk new [REPOSITORY] [--fresh|--choose|--reuse]")
			}
			input = a
		}
	}
	if input == "" {
		if mode != "" {
			return "", errors.New("workspace selection options need a REPOSITORY")
		}
		return create(dir, "", stderr)
	}
	source, err := repository(input)
	if err != nil {
		return "", err
	}
	p, err := loadPreferences(dir)
	if err != nil {
		return "", err
	}
	key := sourceKey(source)
	if mode == "--reuse" {
		err := lifecycleLock(dir, func() error {
			var err error
			p, err = loadPreferences(dir)
			if err != nil {
				return err
			}
			p.Reuse[key] = true
			return writeJSON(filepath.Join(filepath.Dir(dir), "preferences.json"), p)
		})
		if err != nil {
			return "", err
		}
	}
	all, err := names(dir)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, name := range all {
		if o := origin(filepath.Join(dir, name)); o != "" && sourceKey(o) == key {
			matches = append(matches, name)
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		a, _ := readState(dir, matches[i])
		b, _ := readState(dir, matches[j])
		return a.LastUsed.After(b.LastUsed)
	})
	if len(matches) == 0 || mode == "--fresh" {
		return create(dir, input, stderr)
	}
	selected := matches[0]
	if mode == "--choose" || !p.Reuse[key] {
		fmt.Fprintf(stderr, "Workspaces for %s:\n", key)
		for i, name := range matches {
			fmt.Fprintf(stderr, "  %d) %s\n", i+1, name)
		}
		fmt.Fprintln(stderr, "  0) Fresh workspace\nChoose a number (or use --reuse to remember most-recent reuse): ")
		line, err := bufio.NewReader(in).ReadString('\n')
		if err != nil && line == "" {
			return "", errors.New("choose a workspace interactively, or use --fresh or --reuse")
		}
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < 0 || n > len(matches) {
			return "", errors.New("invalid workspace selection")
		}
		if n == 0 {
			return create(dir, input, stderr)
		}
		selected = matches[n-1]
	}
	if err := recordUse(dir, selected); err != nil {
		return "", err
	}
	return filepath.Join(dir, selected), nil
}

func listWorkspaces(dir string, out, stderr io.Writer) error {
	all, err := names(dir)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Fprintln(stderr, "No workspaces yet. Create one with desk new.")
		return nil
	}
	groups := map[string][]string{}
	for _, name := range all {
		key := "playgrounds"
		if source := origin(filepath.Join(dir, name)); source != "" {
			key = sourceKey(source)
		}
		groups[key] = append(groups[key], name)
	}
	var keys []string
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	here := current(dir)
	for _, key := range keys {
		fmt.Fprintln(out, key)
		for i, name := range groups[key] {
			branch, marker := "├──", " "
			if i == len(groups[key])-1 {
				branch = "└──"
			}
			if name == here {
				marker = "*"
			}
			s, _ := readState(dir, name)
			note := ""
			if s.Disposable {
				note = " [finished]"
			}
			fmt.Fprintf(out, "%s %s %s%s  %s\n", branch, marker, name, note, filepath.Join(dir, name))
		}
	}
	return nil
}
