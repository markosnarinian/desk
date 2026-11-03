//go:build linux || darwin

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
)

var version = "dev"
var validID = regexp.MustCompile(`^[a-f0-9]{12}$`)

type workspace struct {
	ID          string
	Name        string
	Kind        string
	Source      string
	Created     time.Time
	Pinned      bool
	Done        bool
	Quarantined time.Time
}

type store struct {
	root  string
	items []workspace
	lock  *os.File
}

func dataRoot() (string, error) {
	if p := os.Getenv("DESK_HOME"); p != "" {
		if !filepath.IsAbs(p) {
			return "", errors.New("DESK_HOME must be absolute")
		}
		return filepath.Clean(p), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".desk"), nil
}

func openStore() (*store, error) {
	root, err := dataRoot()
	if err != nil {
		return nil, err
	}
	for _, p := range []string{root, filepath.Join(root, "workspaces"), filepath.Join(root, "quarantine"), filepath.Join(root, "locks")} {
		if err := os.MkdirAll(p, 0700); err != nil {
			return nil, err
		}
		info, err := os.Lstat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("not an ordinary directory: %s", p)
		}
		if err := os.Chmod(p, 0700); err != nil {
			return nil, err
		}
	}
	f, err := lockFile(filepath.Join(root, "registry.lock"), syscall.LOCK_EX)
	if err != nil {
		return nil, err
	}
	s := &store{root: root, lock: f}
	b, err := os.ReadFile(filepath.Join(root, "registry.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		s.close()
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.items); err != nil {
			s.close()
			return nil, fmt.Errorf("invalid registry: %w", err)
		}
	}
	seen := map[string]bool{}
	for _, w := range s.items {
		if !validID.MatchString(w.ID) || seen[w.ID] || (w.Kind != "scratch" && w.Kind != "repo") {
			s.close()
			return nil, errors.New("invalid registry entry")
		}
		seen[w.ID] = true
	}
	return s, nil
}

func lockFile(path string, mode int) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), mode); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func (s *store) close() {
	if s.lock != nil {
		s.lock.Close()
		s.lock = nil
	}
}
func (s *store) save() error {
	b, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(s.root, "registry-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), filepath.Join(s.root, "registry.json"))
}
func (s *store) find(id string) (*workspace, error) {
	for i := range s.items {
		if s.items[i].ID == id {
			return &s.items[i], nil
		}
	}
	return nil, fmt.Errorf("unknown workspace %q (use desk list)", id)
}
func (s *store) location(w *workspace) (string, bool, error) {
	var found string
	var quarantined bool
	for _, state := range []string{"workspaces", "quarantine"} {
		p := filepath.Join(s.root, state, w.ID)
		info, err := os.Lstat(p)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", false, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", false, errors.New("workspace is not an ordinary directory")
		}
		if found != "" {
			return "", false, errors.New("workspace exists in both storage locations")
		}
		found = p
		quarantined = state == "quarantine"
	}
	if found == "" {
		return "", false, errors.New("workspace directory missing; registry retained")
	}
	return found, quarantined, nil
}
func (s *store) create(name, kind, source string) (*workspace, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	w := workspace{ID: hex.EncodeToString(b), Name: name, Kind: kind, Source: source, Created: time.Now().UTC()}
	p := filepath.Join(s.root, "workspaces", w.ID)
	if err := os.Mkdir(p, 0700); err != nil {
		return nil, err
	}
	s.items = append(s.items, w)
	if err := s.save(); err != nil {
		os.Remove(p)
		s.items = s.items[:len(s.items)-1]
		return nil, err
	}
	return &s.items[len(s.items)-1], nil
}

func (s *store) enter(w *workspace) error {
	p, q, err := s.location(w)
	if err != nil {
		return err
	}
	if q {
		return errors.New("workspace is quarantined; restore it first")
	}
	f, err := lockFile(filepath.Join(s.root, "locks", w.ID), syscall.LOCK_SH)
	if err != nil {
		return err
	}
	defer f.Close()
	w.Done = false
	if err := s.save(); err != nil {
		return err
	}
	s.close()
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell, "-i")
	cmd.Dir = p
	cmd.Env = append(os.Environ(), "DESK_WORKSPACE="+w.ID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Preserve the session lock in children as well as the supervising process.
	cmd.ExtraFiles = []*os.File{f}
	fmt.Fprintf(os.Stderr, "desk: entering %s (%s); exit to return; nothing is deleted\n", w.Name, w.ID)
	// The foreground shell receives terminal signals too. Do not abandon its lock.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT)
	defer signal.Stop(signals)
	return cmd.Run()
}

func (s *store) retention(w *workspace) string {
	if w.Pinned {
		return "pinned"
	}
	p, _, err := s.location(w)
	if err != nil {
		return err.Error()
	}
	f, err := lockFile(filepath.Join(s.root, "locks", w.ID), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return "active session or unavailable session lock"
	}
	defer f.Close()
	if w.Kind == "repo" {
		return s.repositoryRetention(p)
	}
	entries, err := os.ReadDir(p)
	if err != nil {
		return "cannot read workspace"
	}
	if len(entries) != 0 {
		return "scratch contains files"
	}
	return ""
}

func (s *store) quarantine(w *workspace) error {
	if reason := s.retention(w); reason != "" {
		return errors.New(reason)
	}
	p, q, err := s.location(w)
	if err != nil {
		return err
	}
	if q {
		return errors.New("already quarantined")
	}
	if err := os.Rename(p, filepath.Join(s.root, "quarantine", w.ID)); err != nil {
		return err
	}
	w.Quarantined = time.Now().UTC()
	return s.save()
}

func (s *store) restore(w *workspace) error {
	p, q, err := s.location(w)
	if err != nil {
		return err
	}
	if !q {
		return errors.New("workspace is not quarantined")
	}
	if err := os.Rename(p, filepath.Join(s.root, "workspaces", w.ID)); err != nil {
		return err
	}
	w.Quarantined = time.Time{}
	w.Done = false
	return s.save()
}

func (s *store) purge(w *workspace) error {
	p, q, err := s.location(w)
	if err != nil {
		return err
	}
	if !q {
		return errors.New("only quarantined workspaces can be purged")
	}
	if w.Quarantined.IsZero() || time.Since(w.Quarantined) < 7*24*time.Hour {
		return errors.New("quarantine must be at least seven days old; restore and remove again if its timestamp is missing")
	}
	if reason := s.retention(w); reason != "" {
		return errors.New(reason)
	}
	if err := os.RemoveAll(p); err != nil {
		return err
	}
	for i := range s.items {
		if s.items[i].ID == w.ID {
			s.items = append(s.items[:i], s.items[i+1:]...)
			break
		}
	}
	return s.save()
}

func (s *store) gc(apply bool, out io.Writer) error {
	for i := range s.items {
		w := &s.items[i]
		_, q, err := s.location(w)
		reason := ""
		if err != nil {
			reason = err.Error()
		} else if q {
			reason = "already quarantined"
		} else if !w.Done {
			reason = "not marked done"
		} else {
			reason = s.retention(w)
		}
		if reason != "" {
			fmt.Fprintf(out, "%s retained: %s\n", w.ID, reason)
			continue
		}
		if !apply {
			fmt.Fprintf(out, "%s would quarantine (requires --apply --quiescent)\n", w.ID)
			continue
		}
		if err := s.quarantine(w); err != nil {
			return err
		}
		fmt.Fprintln(out, w.ID, "quarantined")
	}
	return nil
}

const help = `desk — managed local workspaces

Usage:
  desk scratch [NAME] [--no-enter]      Create a scratch workspace
  desk open REPOSITORY [--no-enter]     Clone or resume a repository
  desk enter ID                        Enter an existing workspace
  desk list                            Show managed workspaces
  desk keep ID [--off]                  Pin or unpin a workspace
  desk done ID                         Mark a workspace for cleanup
  desk gc [--dry-run]                   Explain cleanup (default: dry run)
  desk gc --apply --quiescent           Quarantine safe, done workspaces
  desk remove ID --quiescent            Quarantine a safe workspace
  desk restore ID                      Recover a quarantined workspace
  desk purge ID --quiescent             Permanently remove after seven days
  desk version                         Print version

Use exit to leave a subshell. Nothing is deleted on exit.
--quiescent acknowledges that all external processes using the workspace stopped.
`

func run(args []string, out io.Writer) error {
	if len(args) > 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "version") {
		return errors.New("help and version take no arguments")
	}
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Fprint(out, help)
		return nil
	}
	if args[0] == "version" {
		fmt.Fprintln(out, version)
		return nil
	}
	command := args[0]
	var pos []string
	flags := map[string]bool{}
	allowed := map[string]string{"scratch": "--no-enter", "open": "--no-enter", "enter": "", "list": "", "remove": "--quiescent", "keep": "--off", "done": "", "gc": "--dry-run --apply --quiescent", "restore": "", "purge": "--quiescent"}
	options, ok := allowed[command]
	if !ok {
		return fmt.Errorf("unknown command %q; use desk help", command)
	}
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "-") {
			if !strings.Contains(" "+options+" ", " "+a+" ") {
				return fmt.Errorf("unknown option %s for %s", a, command)
			}
			flags[a] = true
		} else {
			pos = append(pos, a)
		}
	}
	count := 1
	if command == "list" || command == "gc" {
		count = 0
	}
	if len(pos) != count && !(command == "scratch" && len(pos) == 0) {
		return fmt.Errorf("invalid arguments for %s; use desk help", command)
	}
	if flags["--apply"] && flags["--dry-run"] {
		return errors.New("--apply and --dry-run cannot be combined")
	}
	if (command == "remove" || command == "purge" || flags["--apply"]) && !flags["--quiescent"] {
		return errors.New("requires --quiescent: stop all processes using the workspace first")
	}
	s, err := openStore()
	if err != nil {
		return err
	}
	defer s.close()
	if command == "scratch" {
		name := "scratch"
		if len(pos) > 0 {
			name = pos[0]
		}
		w, err := s.create(name, "scratch", "")
		if err != nil {
			return err
		}
		fmt.Fprintln(out, w.ID)
		if !flags["--no-enter"] {
			return s.enter(w)
		}
		return nil
	}
	if command == "open" {
		w, err := s.openRepository(pos[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(out, w.ID)
		if !flags["--no-enter"] {
			return s.enter(w)
		}
		return nil
	}
	if command == "gc" {
		return s.gc(flags["--apply"], out)
	}
	if command == "list" {
		t := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
		fmt.Fprintln(t, "ID\tNAME\tSTATE\tRETENTION\tPATH")
		for i := range s.items {
			w := &s.items[i]
			p, q, e := s.location(w)
			state := "ready"
			if q {
				state = "quarantined"
			} else if w.Done {
				state = "done"
			}
			reason := s.retention(w)
			if e != nil {
				reason = e.Error()
			}
			if reason == "" {
				if q {
					reason = "eligible for purge with --quiescent"
					if w.Quarantined.IsZero() || time.Since(w.Quarantined) < 7*24*time.Hour {
						reason = "quarantine age: seven days required before purge"
					}
				} else {
					reason = "eligible with --quiescent"
				}
			}
			fmt.Fprintf(t, "%s\t%q\t%s\t%s\t%s\n", w.ID, w.Name, state, reason, p)
		}
		return t.Flush()
	}
	w, err := s.find(pos[0])
	if err != nil {
		return err
	}
	switch command {
	case "enter":
		return s.enter(w)
	case "keep":
		w.Pinned = !flags["--off"]
		return s.save()
	case "done":
		if w.Pinned {
			return errors.New("workspace is pinned; use keep ID --off first")
		}
		_, q, err := s.location(w)
		if err != nil {
			return err
		}
		if q {
			return errors.New("already quarantined")
		}
		w.Done = true
		return s.save()
	case "restore":
		if err := s.restore(w); err != nil {
			return err
		}
		fmt.Fprintln(out, "restored", w.ID)
	case "purge":
		id := w.ID
		if err := s.purge(w); err != nil {
			return err
		}
		fmt.Fprintln(out, "purged", id)
	case "remove":
		if err := s.quarantine(w); err != nil {
			return err
		}
		fmt.Fprintln(out, "quarantined", w.ID)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "desk:", err)
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() > 0 {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
}
