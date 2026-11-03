//go:build linux || darwin

package main

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

var version = "dev"

//go:embed shell
var shells embed.FS

const help = `desk — workspaces for temporary local work

Usage:
  desk new [REPOSITORY] [--fresh|--choose|--reuse]
                         Create a playground or choose a repository workspace
  desk amp [THREAD] [--copy|--follow] [--repository REPOSITORY]
                         Bring Amp orb work locally; omit THREAD to browse
  desk switch [NAME]      Go to a workspace (needs shell integration)
  desk info [NAME] [--path] Show Name, Path, Source (or just the path)
  desk list              Show workspaces grouped by repository
  desk rename [NAME] NEW_NAME  Rename a task workspace
  desk finish [NAME]     Mark a playground eligible for later cleanup
  desk cleanup [--yes]   Retire safe workspaces unused for 72 hours
  desk rm [NAME]          Retire a workspace after confirmation
  desk init [SHELL] [--print] Install shell integration, or print its script
  desk version            Print version

NAME is a workspace name or a unique prefix of one. Without NAME, desk uses the
workspace containing the current directory. REPOSITORY is owner/repo (GitHub),
a Git URL, or a local path. --reuse remembers most-recent reuse for that repository;
--choose shows the chooser again. --fresh always creates an independent clone.

Removed workspaces stay in ~/.desk/garbage for 72 hours. Every desk command
silently collects expired garbage. DESK_HOME overrides ~/.desk.

Run desk init, then open a new terminal. SHELL defaults to the basename of $SHELL.
Supported shells: zsh (including Oh My Zsh), bash, fish.
Use desk init oh-my-zsh only to force Oh My Zsh detection.
Use desk info [NAME] --path in scripts and shell wrappers.
`

// workspacesDir is where every workspace lives, one directory per name.
func workspacesDir() (string, error) {
	if p := os.Getenv("DESK_HOME"); p != "" {
		if !filepath.IsAbs(p) {
			return "", errors.New("DESK_HOME must be absolute")
		}
		return filepath.Join(filepath.Clean(p), "workspaces"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".desk", "workspaces"), nil
}

func names(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

// current returns the workspace containing the working directory, or "".
func current(dir string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(wd); err == nil {
		wd = resolved
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	rel, err := filepath.Rel(dir, wd)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return ""
	}
	return strings.Split(rel, string(os.PathSeparator))[0]
}

// resolve finds a workspace by exact name or unique prefix; "" means current.
func resolve(dir, query string) (string, error) {
	if query == "" {
		if name := current(dir); name != "" {
			return name, nil
		}
		return "", errors.New("not inside a workspace; give a NAME (see desk list)")
	}
	all, err := names(dir)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, n := range all {
		if n == query {
			return n, nil
		}
		if strings.HasPrefix(n, query) {
			matches = append(matches, n)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no workspace %q (see desk list)", query)
	case 1:
		return matches[0], nil
	}
	return "", fmt.Errorf("%q matches %s", query, strings.Join(matches, ", "))
}

// reserve creates a new, empty workspace directory with an unused name.
func reserve(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	for range 100 {
		name := adjectives[rand.IntN(len(adjectives))] + "-" + nouns[rand.IntN(len(nouns))]
		err := os.Mkdir(filepath.Join(dir, name), 0700)
		if err == nil {
			return name, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
	return "", errors.New("could not find an unused workspace name")
}

var shorthand = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*/[A-Za-z0-9][A-Za-z0-9_.-]*$`)
var scpLike = regexp.MustCompile(`^(?:[^@/]+@)?([^:/]+):(.+)$`)
var workspaceName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// repository turns user input into a clone source.
func repository(input string) (string, error) {
	switch {
	case shorthand.MatchString(input):
		return "https://github.com/" + strings.TrimSuffix(input, ".git") + ".git", nil
	case strings.HasPrefix(input, "/"), strings.HasPrefix(input, "./"), strings.HasPrefix(input, "../"), input == ".", input == "..":
		return filepath.Abs(input)
	case strings.Contains(input, "://"), scpLike.MatchString(input):
		return input, nil
	}
	return "", errors.New("use owner/repo, a Git URL, or a local path such as ./repo")
}

// sourceKey identifies a repository regardless of how its URL is spelled, so
// https://github.com/o/r, git@github.com:o/r.git, and o/r all match.
func sourceKey(source string) string {
	host, path := "", source
	if u, err := url.Parse(source); err == nil && strings.Contains(source, "://") {
		if u.Scheme == "file" {
			return filepath.Clean(u.Path)
		}
		host, path = u.Hostname(), u.Path
	} else if strings.HasPrefix(source, "/") {
		return filepath.Clean(source)
	} else if m := scpLike.FindStringSubmatch(source); m != nil {
		host, path = m[1], m[2]
	}
	host = strings.ToLower(host)
	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	if host == "github.com" {
		path = strings.ToLower(path)
	}
	return host + "/" + path
}

func gitEnv() []string {
	var env []string
	for _, v := range os.Environ() {
		// A parent Git invocation must not redirect desk to another checkout.
		if strings.HasPrefix(v, "GIT_DIR=") || strings.HasPrefix(v, "GIT_WORK_TREE=") || strings.HasPrefix(v, "GIT_INDEX_FILE=") {
			continue
		}
		env = append(env, v)
	}
	return env
}

func gitCommand(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = gitEnv()
	return cmd
}

func git(dir string, args ...string) string {
	cmd := gitCommand(dir, args...)
	b, _ := cmd.Output()
	return strings.TrimSpace(string(b))
}

func origin(path string) string {
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return ""
	}
	return git(path, "config", "--get", "remote.origin.url")
}

func create(dir, input string, stderr io.Writer) (string, error) {
	source := ""
	if input != "" {
		var err error
		if source, err = repository(input); err != nil {
			return "", err
		}
	}
	name, err := reserve(dir)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if source != "" {
		cmd := exec.Command("git", "clone", "--", source, path)
		cmd.Env = gitEnv()
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stderr, stderr
		if err := cmd.Run(); err != nil {
			os.RemoveAll(path)
			return "", fmt.Errorf("clone failed: %w", err)
		}
	}
	if err := recordUse(dir, name); err != nil {
		return "", err
	}
	fmt.Fprintf(stderr, "desk: created %s\n", name)
	return path, nil
}

func remove(dir, name string, in io.Reader, stderr io.Writer) error {
	path := filepath.Join(dir, name)
	if followActive(dir, name) {
		return errors.New("stop live follow before removing this workspace")
	}
	what := "empty"
	if entries, _ := os.ReadDir(path); len(entries) > 0 {
		what = "has files"
	}
	if o := origin(path); o != "" {
		what = o
	}
	var notes []string
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		if status := git(path, "status", "--porcelain"); status != "" {
			notes = append(notes, fmt.Sprintf("%d uncommitted changes", strings.Count(status, "\n")+1))
		}
		if n := git(path, "rev-list", "--count", "--branches", "--not", "--remotes"); n != "" && n != "0" {
			notes = append(notes, n+" unpushed commits")
		}
	}
	if len(notes) > 0 {
		what += "; " + strings.Join(notes, ", ")
	}
	fmt.Fprintf(stderr, "Remove %s (%s)? Retain in garbage for 72 hours. [y/N] ", name, what)
	answer, _ := bufio.NewReader(in).ReadString('\n')
	if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
		return errors.New("nothing deleted")
	}
	err := lifecycleLock(dir, func() error {
		if followActive(dir, name) {
			return errors.New("stop live follow before removing this workspace")
		}
		return trashWorkspace(dir, name)
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "desk: moved %s to garbage (retained for 72 hours)\n", name)
	return nil
}

func run(args []string, in io.Reader, out, stderr io.Writer) error {
	// Even help/version invocations collect expired garbage, without chatter.
	if dir, err := workspacesDir(); err == nil {
		collectGarbage(dir, time.Now())
	}
	command, rest := "help", []string(nil)
	if len(args) > 0 {
		command, rest = args[0], args[1:]
	}
	maxArgs := map[string]int{"new": 2, "amp": 4, "info": 2, "rm": 1, "rename": 2, "finish": 1, "cleanup": 1, "_visit": 1, "list": 0, "init": 2, "version": 0, "help": 0, "--help": 0, "-h": 0}
	limit, ok := maxArgs[command]
	if command == "switch" {
		return errors.New("desk switch needs shell integration: run desk init, then open a new terminal")
	}
	if !ok {
		return fmt.Errorf("unknown command %q; see desk help", command)
	}
	if len(rest) > limit {
		return fmt.Errorf("wrong arguments for %s; see desk help", command)
	}
	arg := ""
	if len(rest) == 1 {
		arg = rest[0]
	}
	switch command {
	case "help", "--help", "-h":
		fmt.Fprint(out, help)
		return nil
	case "version":
		v := version
		if v == "dev" {
			if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
				v = info.Main.Version
			}
		}
		fmt.Fprintln(out, v)
		return nil
	case "init":
		return initShell(rest, out)
	}
	pathOnly := false
	if command == "info" {
		var err error
		arg, pathOnly, err = optionalFlag(rest, "--path")
		if err != nil {
			return err
		}
	}
	dir, err := workspacesDir()
	if err != nil {
		return err
	}
	switch command {
	case "new":
		path, err := openWorkspace(dir, rest, in, stderr)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, path)
	case "amp":
		return ampWorkspace(dir, rest, in, out, stderr)
	case "cleanup":
		if arg != "" && arg != "--yes" {
			return errors.New("use desk cleanup [--yes]")
		}
		return cleanup(dir, arg == "--yes", in, stderr)
	case "rename":
		if len(rest) == 0 {
			return errors.New("use desk rename [NAME] NEW_NAME")
		}
		query, newName := "", rest[len(rest)-1]
		if len(rest) == 2 {
			query = rest[0]
		}
		name, err := resolve(dir, query)
		if err != nil {
			return err
		}
		path, err := renameWorkspace(dir, name, newName)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, path)
	case "finish", "_visit":
		name, err := resolve(dir, arg)
		if err != nil {
			return err
		}
		if command == "_visit" {
			return recordUse(dir, name)
		}
		err = lifecycleLock(dir, func() error {
			if err := writeUse(dir, name); err != nil {
				return err
			}
			s, err := readState(dir, name)
			if err != nil {
				return err
			}
			s.Disposable = true
			return writeJSON(statePath(dir, name), s)
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stderr, "desk: marked %s finished; eligible after 72 hours unused (Git safety checks still apply)\n", name)
	case "info", "rm":
		name, err := resolve(dir, arg)
		if err != nil {
			return err
		}
		if command == "rm" {
			return remove(dir, name, in, stderr)
		}
		path := filepath.Join(dir, name)
		if pathOnly {
			fmt.Fprintln(out, path)
		} else {
			source := origin(path)
			if source == "" {
				source = "scratch"
			}
			fmt.Fprintf(out, "Name: %s\nPath: %s\nSource: %s\n", name, path, source)
		}
	case "list":
		return listWorkspaces(dir, out, stderr)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "desk:", err)
		os.Exit(1)
	}
}
