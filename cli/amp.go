//go:build linux || darwin

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var ampThreadID = regexp.MustCompile(`^T-[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// These are the public `amp threads list --json` and `threads export` shapes.
// Decode metadata only: messages are deliberately neither retained nor printed.
type ampThread struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Updated string `json:"updated"`
}

type ampTree struct {
	URI        string `json:"uri"`
	Repository struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"repository"`
}

type ampExport struct {
	Env struct {
		Initial struct {
			WorkingDirectory string    `json:"workingDirectory"`
			Trees            []ampTree `json:"trees"`
		} `json:"initial"`
	} `json:"env"`
}

func ampNormalizeThread(input string) (string, error) {
	if ampThreadID.MatchString(input) {
		return input, nil
	}
	u, err := url.Parse(input)
	if err == nil && u.Scheme == "https" && u.Host == "ampcode.com" && u.User == nil {
		path := strings.TrimSuffix(u.Path, "/")
		if id, ok := strings.CutPrefix(path, "/threads/"); ok && ampThreadID.MatchString(id) {
			return id, nil
		}
	}
	return "", errors.New("THREAD must be an Amp thread ID (T-…) or https://ampcode.com/threads/T-… URL")
}

func ampReadJSON(stderr io.Writer, result any, args ...string) error {
	cmd := exec.Command("amp", args...)
	cmd.Stderr = stderr
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("amp %s failed: %w", strings.Join(args, " "), err)
	}
	if err := json.Unmarshal(b, result); err != nil {
		// Do not include the payload: exports contain private messages.
		return fmt.Errorf("amp %s returned invalid JSON: %w", strings.Join(args, " "), err)
	}
	return nil
}

func ampChoose(in *bufio.Reader, stderr io.Writer, count int) (int, error) {
	fmt.Fprintf(stderr, "Choose 1-%d (blank to cancel): ", count)
	line, err := in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, errors.New("cancelled")
	}
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > count {
		return 0, errors.New("invalid selection")
	}
	return n - 1, nil
}

func ampRepository(thread string, stderr io.Writer) (string, error) {
	var exported ampExport
	if err := ampReadJSON(stderr, &exported, "threads", "export", thread); err != nil {
		return "", err
	}
	env := exported.Env.Initial
	// Match the working directory to its tree, including nested directories.
	// With multiple trees, never silently clone an unrelated repository.
	best, bestLen := "", -1
	unique := make(map[string]string)
	for _, tree := range env.Trees {
		if tree.Repository.Type != "git" || tree.Repository.URL == "" {
			continue
		}
		unique[sourceKey(tree.Repository.URL)] = tree.Repository.URL
		root := strings.TrimRight(tree.URI, "/")
		if root != "" && (env.WorkingDirectory == root || strings.HasPrefix(env.WorkingDirectory, root+"/")) && len(root) > bestLen {
			best, bestLen = tree.Repository.URL, len(root)
		}
	}
	if best != "" {
		return best, nil
	}
	if len(unique) == 1 {
		for _, source := range unique {
			return source, nil
		}
	}
	return "", errors.New("cannot determine the thread repository; give --repository REPOSITORY")
}

// ampWorkspace clones a fresh workspace and only invokes Amp's sync command;
// it never starts or messages an agent. stdout is reserved for the final path.
func ampWorkspace(dir string, args []string, in io.Reader, out, stderr io.Writer) error {
	thread, source, mode := "", "", ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--copy" || arg == "--follow":
			if mode != "" {
				return errors.New("choose only one of --copy or --follow")
			}
			mode = strings.TrimPrefix(arg, "--")
		case arg == "--repository":
			if source != "" || i+1 == len(args) || strings.HasPrefix(args[i+1], "--") {
				return errors.New("--repository requires one REPOSITORY")
			}
			i++
			source = args[i]
			if source == "" {
				return errors.New("--repository requires one REPOSITORY")
			}
		case strings.HasPrefix(arg, "--repository="):
			if source != "" || arg == "--repository=" {
				return errors.New("--repository requires one REPOSITORY")
			}
			source = strings.TrimPrefix(arg, "--repository=")
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown desk amp option %q", arg)
		default:
			if thread != "" {
				return errors.New("desk amp accepts only one THREAD")
			}
			var err error
			thread, err = ampNormalizeThread(arg)
			if err != nil {
				return err
			}
		}
	}
	if _, err := exec.LookPath("amp"); err != nil {
		return errors.New("desk amp requires the installed amp CLI (https://ampcode.com)")
	}
	reader := bufio.NewReader(in)
	if thread == "" {
		var threads []ampThread
		if err := ampReadJSON(stderr, &threads, "threads", "list", "--json", "--limit", "20"); err != nil {
			return err
		}
		if len(threads) == 0 {
			return errors.New("no recent Amp threads")
		}
		fmt.Fprintln(stderr, "Recent Amp threads:")
		for i, t := range threads {
			fmt.Fprintf(stderr, "%d) %s  %s  %s\n", i+1, strings.Join(strings.Fields(t.Title), " "), t.Updated, t.ID)
		}
		i, err := ampChoose(reader, stderr, len(threads))
		if err != nil {
			return err
		}
		thread, err = ampNormalizeThread(threads[i].ID)
		if err != nil {
			return err
		}
	}
	if mode == "" {
		fmt.Fprintln(stderr, "1) Copy current changes once\n2) Follow live changes in the foreground")
		i, err := ampChoose(reader, stderr, 2)
		if err != nil {
			return err
		}
		mode = []string{"copy", "follow"}[i]
	}
	if source == "" {
		var err error
		source, err = ampRepository(thread, stderr)
		if err != nil {
			return err
		}
	}
	path, err := create(dir, source, stderr)
	if err != nil {
		return err
	}
	syncArgs := []string{"sync", "--apply", thread, "--checkout"}
	if mode == "follow" {
		syncArgs = []string{"sync", thread, "--checkout"}
		marker := filepath.Join(filepath.Dir(dir), "state", filepath.Base(path)+".follow")
		if err := os.MkdirAll(filepath.Dir(marker), 0700); err != nil {
			return fmt.Errorf("workspace kept at %s; cannot protect live follow: %w", path, err)
		}
		if err := os.WriteFile(marker, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			return fmt.Errorf("workspace kept at %s; cannot protect live follow: %w", path, err)
		}
		defer os.Remove(marker)
		fmt.Fprintf(stderr, "desk: following in %s; in another terminal use desk switch %s to test. Ctrl-C stops following and restores the original checkout.\n", path, filepath.Base(path))
	}
	cmd := exec.Command("amp", syncArgs...)
	cmd.Dir, cmd.Env = path, gitEnv()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = in, stderr, stderr
	if mode == "follow" {
		err = ampFollow(cmd)
	} else {
		err = cmd.Run()
	}
	if err != nil {
		fmt.Fprintf(stderr, "desk: sync failed; workspace kept at %s\n", path)
		return fmt.Errorf("amp sync failed: %w", err)
	}
	_, err = fmt.Fprintln(out, path)
	return err
}

func ampFollow(cmd *exec.Cmd) error {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	// Only desk receives terminal signals; forward each once. Otherwise Amp
	// receives both the terminal's Ctrl-C and our forwarded copy, which can
	// interrupt restoration of its original checkout.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	for {
		select {
		case sig := <-signals:
			_ = cmd.Process.Signal(sig)
		case err := <-done:
			if err != nil {
				var exit *exec.ExitError
				if errors.As(err, &exit) {
					status, ok := exit.Sys().(syscall.WaitStatus)
					if (ok && status.Signaled() && status.Signal() == syscall.SIGINT) || exit.ExitCode() == 130 {
						return nil
					}
				}
			}
			return err
		}
	}
}
