//go:build linux || darwin

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const ampTestThread = "T-00000000-0000-0000-0000-000000000001"
const ampTestSecond = "T-00000000-0000-0000-0000-000000000002"

type ampFixture struct {
	dir, repo, trace, export, list, ready string
}

func ampTestWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

func ampTestSetup(t *testing.T) ampFixture {
	t.Helper()
	root := t.TempDir()
	f := ampFixture{
		dir: filepath.Join(root, "workspaces"), repo: filepath.Join(root, "repository"),
		trace: filepath.Join(root, "trace"), export: filepath.Join(root, "export.json"),
		list: filepath.Join(root, "list.json"), ready: filepath.Join(root, "ready"),
	}
	if err := os.Mkdir(f.repo, 0700); err != nil {
		t.Fatal(err)
	}
	gitRun(t, f.repo, "init", "-q")
	ampTestWrite(t, filepath.Join(f.repo, "tracked.txt"), "original\n")
	gitRun(t, f.repo, "add", ".")
	gitRun(t, f.repo, "commit", "-qm", "fixture")
	ampTestWrite(t, f.list, fmt.Sprintf(`[{"id":%q,"title":"First thread","updated":"2026-11-03"},{"id":%q,"title":"Second thread","updated":"2026-11-02"}]`, ampTestThread, ampTestSecond))
	ampTestWrite(t, f.export, ampTestExport(f.repo))
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(bin, "amp")
	ampTestWrite(t, fake, `#!/bin/sh
printf '%s\n' "$*" >> "$AMP_TEST_TRACE"
case "$1 $2" in
  "threads list")
    [ "$AMP_TEST_FAIL" != list ] || exit 21
    cat "$AMP_TEST_LIST"
    ;;
  "threads export")
    [ "$AMP_TEST_FAIL" != export ] || exit 22
    cat "$AMP_TEST_EXPORT"
    ;;
  "sync "* )
    printf '%s\n' "$PWD" > "$AMP_TEST_READY"
    echo 'amp process stdout'
    echo 'amp process stderr' >&2
    name=${PWD##*/}
    marker="$(dirname "$(dirname "$PWD")")/state/$name.follow"
    case "$2" in
      --apply) [ ! -e "$marker" ] || exit 41 ;;
      *) [ -s "$marker" ] || exit 42 ;;
    esac
    [ "$AMP_TEST_FAIL" != sync ] || exit 23
    if [ "$AMP_TEST_WAIT" = 1 ]; then exec sleep 30; fi
    echo synced > synced.txt
    ;;
  *) echo 'unexpected command (would start an agent)' >&2; exit 99 ;;
esac
`)
	if err := os.Chmod(fake, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("AMP_TEST_TRACE", f.trace)
	t.Setenv("AMP_TEST_EXPORT", f.export)
	t.Setenv("AMP_TEST_LIST", f.list)
	t.Setenv("AMP_TEST_READY", f.ready)
	t.Setenv("AMP_TEST_FAIL", "")
	t.Setenv("AMP_TEST_WAIT", "")
	return f
}

func ampTestExport(source string) string {
	return fmt.Sprintf(`{"messages":[{"content":"PRIVATE MESSAGE MUST NOT APPEAR"}],"env":{"initial":{"workingDirectory":"file:///workspace/repo/src","trees":[{"uri":"file:///workspace/repo","repository":{"type":"git","url":%q}}]}}}`, source)
}

func ampTestInvoke(f ampFixture, input string, args ...string) (string, string, error) {
	var out, stderr bytes.Buffer
	err := ampWorkspace(f.dir, args, strings.NewReader(input), &out, &stderr)
	return out.String(), stderr.String(), err
}

func ampTestTrace(t *testing.T, f ampFixture) string {
	t.Helper()
	b, err := os.ReadFile(f.trace)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(b)
}

func TestAmpCopyDirectThreadFreshWorkspace(t *testing.T) {
	f := ampTestSetup(t)
	var previous string
	for range 2 {
		out, stderr, err := ampTestInvoke(f, "", "https://ampcode.com/threads/"+ampTestThread+"?view=full", "--copy")
		if err != nil {
			t.Fatalf("copy: %v\n%s", err, stderr)
		}
		path := strings.TrimSuffix(out, "\n")
		if out != path+"\n" || filepath.Dir(path) != f.dir || path == previous {
			t.Fatalf("not a fresh, stdout-only path: %q (previous %q)", out, previous)
		}
		previous = path
		if origin(path) != f.repo || git(path, "rev-parse", "HEAD") != git(f.repo, "rev-parse", "HEAD") {
			t.Fatal("not a clone of the thread repository")
		}
		if _, err := os.Stat(filepath.Join(path, "synced.txt")); err != nil {
			t.Fatal("sync did not run in the clone", err)
		}
		if !strings.Contains(stderr, "amp process stdout") || !strings.Contains(stderr, "amp process stderr") || strings.Contains(stderr, "PRIVATE MESSAGE") {
			t.Fatalf("incorrect process output routing: %s", stderr)
		}
	}
	want := "threads export " + ampTestThread + "\nsync --apply " + ampTestThread + " --checkout\n"
	if trace := ampTestTrace(t, f); trace != want+want {
		t.Fatalf("unexpected Amp commands: %q", trace)
	}
}

func TestAmpBrowseAndChooseMode(t *testing.T) {
	for _, mode := range []string{"1", "2"} {
		t.Run(mode, func(t *testing.T) {
			f := ampTestSetup(t)
			out, stderr, err := ampTestInvoke(f, "2\n"+mode+"\n")
			if err != nil {
				t.Fatalf("browse: %v\n%s", err, stderr)
			}
			if !strings.Contains(stderr, "First thread") || !strings.Contains(stderr, "Second thread") || !strings.Contains(stderr, "Copy current changes") {
				t.Fatalf("missing menus: %s", stderr)
			}
			sync := "sync --apply " + ampTestSecond + " --checkout\n"
			if mode == "2" {
				sync = "sync " + ampTestSecond + " --checkout\n"
				if !strings.Contains(stderr, "another terminal use desk switch") || !strings.Contains(stderr, "Ctrl-C") {
					t.Fatalf("missing foreground follow instructions: %s", stderr)
				}
			}
			if trace := ampTestTrace(t, f); trace != "threads list --json --limit 20\nthreads export "+ampTestSecond+"\n"+sync {
				t.Fatalf("unexpected Amp commands: %q", trace)
			}
			marker := filepath.Join(filepath.Dir(f.dir), "state", filepath.Base(strings.TrimSpace(out))+".follow")
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("follow marker was not removed: %v", err)
			}
		})
	}
}

func TestAmpRepositoryResolution(t *testing.T) {
	for _, scenario := range []string{"matching-tree", "single-repository", "missing", "ambiguous", "override"} {
		t.Run(scenario, func(t *testing.T) {
			f := ampTestSetup(t)
			var payload any
			if err := json.Unmarshal([]byte(ampTestExport(f.repo)), &payload); err != nil {
				t.Fatal(err)
			}
			initial := payload.(map[string]any)["env"].(map[string]any)["initial"].(map[string]any)
			tree := initial["trees"].([]any)[0]
			switch scenario {
			case "matching-tree", "ambiguous":
				initial["trees"] = []any{map[string]any{"uri": "file:///other", "repository": map[string]any{"type": "git", "url": "/nonexistent"}}, tree}
				if scenario == "ambiguous" {
					initial["workingDirectory"] = "file:///unknown"
				}
			case "single-repository":
				initial["workingDirectory"] = "file:///unknown"
			case "missing", "override":
				initial["trees"] = []any{}
			}
			b, _ := json.Marshal(payload)
			ampTestWrite(t, f.export, string(b))
			args := []string{ampTestThread, "--copy"}
			if scenario == "override" {
				args = append(args, "--repository="+f.repo)
			}
			out, stderr, err := ampTestInvoke(f, "", args...)
			if scenario == "missing" || scenario == "ambiguous" {
				if err == nil || !strings.Contains(err.Error(), "--repository") || out != "" {
					t.Fatalf("metadata failure: out=%q err=%v", out, err)
				}
			} else if err != nil || origin(strings.TrimSpace(out)) != f.repo {
				t.Fatalf("resolution: %q %v\n%s", out, err, stderr)
			}
			if scenario == "override" && strings.Contains(ampTestTrace(t, f), "export") {
				t.Fatal("explicit repository should not require readable thread metadata")
			}
		})
	}
}

func TestAmpFailuresKeepStdoutEmpty(t *testing.T) {
	for _, scenario := range []string{"list", "export", "sync", "follow-sync", "invalid-json", "empty-list", "clone", "cancel", "bad-selection", "marker"} {
		t.Run(scenario, func(t *testing.T) {
			f := ampTestSetup(t)
			args := []string{ampTestThread, "--copy"}
			input := ""
			switch scenario {
			case "list", "export", "sync":
				t.Setenv("AMP_TEST_FAIL", scenario)
				if scenario == "list" {
					args = []string{"--copy"}
				}
			case "follow-sync":
				t.Setenv("AMP_TEST_FAIL", "sync")
				args = []string{ampTestThread, "--follow"}
			case "invalid-json":
				ampTestWrite(t, f.export, `{"messages":["PRIVATE MESSAGE MUST NOT APPEAR"], broken`)
			case "empty-list":
				ampTestWrite(t, f.list, `[]`)
				args = []string{"--copy"}
			case "clone":
				args = append(args, "--repository", "/nonexistent/desk-test-repository")
			case "cancel", "bad-selection":
				args = []string{"--copy"}
				if scenario == "bad-selection" {
					input = "21\n"
				}
			case "marker":
				ampTestWrite(t, filepath.Join(filepath.Dir(f.dir), "state"), "not a directory")
				args = []string{ampTestThread, "--follow"}
			}
			out, stderr, err := ampTestInvoke(f, input, args...)
			if err == nil || out != "" || strings.Contains(stderr+err.Error(), "PRIVATE MESSAGE") {
				t.Fatalf("failure leaked output or succeeded: out=%q stderr=%q err=%v", out, stderr, err)
			}
			if scenario == "sync" || scenario == "follow-sync" || scenario == "marker" {
				all, _ := names(f.dir)
				if len(all) != 1 || !strings.Contains(stderr+err.Error(), filepath.Join(f.dir, all[0])) {
					t.Fatalf("failed workspace not retained/reported: %v %s %v", all, stderr, err)
				}
				if scenario == "follow-sync" {
					if _, err := os.Stat(filepath.Join(filepath.Dir(f.dir), "state", all[0]+".follow")); !os.IsNotExist(err) {
						t.Fatalf("failed follow left marker: %v", err)
					}
				}
			}
		})
	}
}

func TestAmpInvalidArguments(t *testing.T) {
	f := ampTestSetup(t)
	for _, args := range [][]string{
		{ampTestThread, "--copy", "--follow"}, {ampTestThread, ampTestSecond},
		{"--unknown"}, {"--repository"}, {"--repository="},
		{"--repository", "--copy"}, {"not-a-thread"},
		{"https://example.com/threads/" + ampTestThread},
		{ampTestThread, "--repository", f.repo, "--repository", f.repo},
	} {
		out, _, err := ampTestInvoke(f, "", args...)
		if err == nil || out != "" {
			t.Fatalf("accepted %v: %q %v", args, out, err)
		}
	}
	if trace := ampTestTrace(t, f); trace != "" {
		t.Fatalf("invalid args invoked amp: %s", trace)
	}
}

func TestAmpMissingCLI(t *testing.T) {
	f := ampTestSetup(t)
	t.Setenv("PATH", t.TempDir())
	out, _, err := ampTestInvoke(f, "", ampTestThread, "--copy")
	if err == nil || !strings.Contains(err.Error(), "installed amp CLI") || out != "" {
		t.Fatalf("missing CLI: out=%q err=%v", out, err)
	}
	if all, _ := names(f.dir); len(all) != 0 {
		t.Fatalf("missing CLI created a workspace: %v", all)
	}
}

// Run a real desk call in a subprocess so sending Ctrl-C cannot kill go test.
func TestAmpFollowInterruptHelper(t *testing.T) {
	if os.Getenv("AMP_TEST_HELPER") != "1" {
		return
	}
	err := ampWorkspace(os.Getenv("AMP_TEST_DIR"), []string{ampTestThread, "--follow", "--repository", os.Getenv("AMP_TEST_REPO")}, strings.NewReader(""), os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestAmpFollowInterruptReturnsPathAndRemovesLease(t *testing.T) {
	for _, target := range []string{"process", "terminal-group"} {
		t.Run(target, func(t *testing.T) { checkAmpInterrupt(t, target) })
	}
}

func checkAmpInterrupt(t *testing.T, target string) {
	t.Helper()
	f := ampTestSetup(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestAmpFollowInterruptHelper$")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), "AMP_TEST_HELPER=1", "AMP_TEST_DIR="+f.dir, "AMP_TEST_REPO="+f.repo, "AMP_TEST_WAIT=1")
	outFile := filepath.Join(t.TempDir(), "stdout")
	out, err := os.Create(outFile)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = out, &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	deadline := time.Now().Add(5 * time.Second)
	var path string
	for time.Now().Before(deadline) {
		b, _ := os.ReadFile(f.ready)
		path = strings.TrimSpace(string(b))
		if path != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if path == "" {
		t.Fatal("fake amp did not start")
	}
	marker := filepath.Join(filepath.Dir(f.dir), "state", filepath.Base(path)+".follow")
	b, err := os.ReadFile(marker)
	if err != nil || string(b) != strconv.Itoa(cmd.Process.Pid) {
		t.Fatalf("active lease PID: %q %v", b, err)
	}
	if b, err := os.ReadFile(outFile); err != nil || len(b) != 0 {
		t.Fatalf("follow printed path before stopping: %q %v", b, err)
	}
	if target == "terminal-group" {
		err = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	} else {
		err = cmd.Process.Signal(os.Interrupt)
	}
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		b, readErr := os.ReadFile(outFile)
		if err != nil || readErr != nil || string(b) != path+"\n" {
			t.Fatalf("interrupted follow: %v out=%q (read: %v) stderr=%s", err, b, readErr, stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Ctrl-C did not stop foreground sync")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("follow lease remained after interrupt: %v", err)
	}
}
