package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the public binary and real startup files, not just generated strings.
func TestShellIntegration(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "desk's \"bin\" $cash `code` \\ [x]")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(binDir, "desk")
	if b, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, b)
	}
	for _, shell := range []string{"bash", "bash-login-profile", "bash-login-bash-login", "bash-login-dot-profile", "zsh", "fish", "oh-my-zsh", "omz-local-custom", "omz-no-source"} {
		t.Run(shell, func(t *testing.T) {
			realShell := shell
			if strings.HasPrefix(shell, "bash") {
				realShell = "bash"
			}
			if strings.HasPrefix(shell, "omz") || shell == "oh-my-zsh" {
				realShell = "zsh"
			}
			shellPath, err := exec.LookPath(realShell)
			if err != nil {
				t.Skipf("%s is not installed", realShell)
			}
			home := isolatedHome(t)
			t.Setenv("SHELL", shellPath)
			t.Setenv("PATH", "/usr/bin:/bin")
			t.Setenv("TERM", "xterm")
			t.Setenv("DESK_HOME", filepath.Join(home, "work's $root \\ [x]"))
			t.Setenv("ZDOTDIR", filepath.Join(home, "zsh configs"))
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg configs"))
			expected := filepath.Join(os.Getenv("DESK_HOME"), "workspaces", "quiet-dune")
			t.Setenv("EXPECTED", expected)
			for _, name := range []string{"quiet-dune", "noisy-fire", "noisy-sky"} {
				if err := os.MkdirAll(filepath.Join(filepath.Dir(expected), name), 0700); err != nil {
					t.Fatal(err)
				}
			}
			initArgs := []string{"init"} // Test $SHELL detection through the public command.
			omzCheck := ""
			if realShell == "bash" {
				mustWrite(t, filepath.Join(home, ".bashrc"), "if shopt -q login_shell; then return; fi\neval \"$(desk init bash)\"\n")
				if strings.HasPrefix(shell, "bash-login") {
					profile := ".bash_profile"
					if shell == "bash-login-bash-login" {
						profile = ".bash_login"
					} else if shell == "bash-login-dot-profile" {
						profile = ".profile"
					}
					mustWrite(t, filepath.Join(home, profile), "# keep my login settings\neval \"$(desk init bash)\"\n")
				}
			} else if realShell == "fish" {
				mustWrite(t, filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "fish", "config.fish"), "desk init fish | source\n")
			} else {
				mustWrite(t, filepath.Join(os.Getenv("ZDOTDIR"), ".zshrc"), "eval \"$(desk init zsh)\"\n")
			}
			if shell == "oh-my-zsh" || strings.HasPrefix(shell, "omz-") {
				initArgs = []string{"init", "oh-my-zsh"}
				t.Setenv("ZSH", filepath.Join(home, "oh my zsh"))
				custom := filepath.Join(home, "my custom's plugins")
				t.Setenv("ZSH_CUSTOM", custom)
				startup := filepath.Join(os.Getenv("ZDOTDIR"), ".zshrc")
				original := "plugins=(git desk docker)\n"
				if shell == "omz-local-custom" {
					t.Setenv("ZSH_CUSTOM", "") // A shell variable is not visible to the installer.
					original = "ZSH_CUSTOM=" + shellQuote(custom, "zsh") + "\nplugins=(git docker)\n"
					omzCheck = "[[ $ZSH_CUSTOM == " + shellQuote(custom, "zsh") + " ]] || exit 33\n"
				}
				if shell != "omz-no-source" {
					// A minimal OMZ loader exercises the real plugin loading contract.
					mustWrite(t, filepath.Join(os.Getenv("ZSH"), "oh-my-zsh.sh"), "for plugin in $plugins; do\n  source \"$ZSH_CUSTOM/plugins/$plugin/$plugin.plugin.zsh\" || return\ndone\n")
					for _, name := range []string{"git", "docker"} {
						mustWrite(t, filepath.Join(custom, "plugins", name, name+".plugin.zsh"), "typeset -g "+name+"_loaded=yes\n")
					}
					original += "source \"$ZSH/oh-my-zsh.sh\"\n"
					omzCheck += "[[ $git_loaded == yes && $docker_loaded == yes ]] || exit 31\n"
				}
				mustWrite(t, startup, original)
				mustWrite(t, filepath.Join(custom, "plugins", "desk", "desk.plugin.zsh"), "# preserve plugin customization\neval \"$(desk init zsh)\"\n")
				if shell == "omz-local-custom" {
					omzCheck += "[[ ${(j: :)plugins} == 'git docker' ]] || exit 32\n"
				} else {
					omzCheck += "[[ ${(j: :)plugins} == 'git desk docker' ]] || exit 32\n"
				}
			}
			for range 2 {
				if b, err := exec.Command(binary, initArgs...).CombinedOutput(); err != nil || !strings.Contains(string(b), "Open a new terminal") {
					t.Fatalf("public init: %v\n%s", err, b)
				}
			}
			script := omzCheck + `desk switch quiet || exit 21
test "$PWD" = "$EXPECTED" || exit 22
desk switch || exit 23
if desk switch missing; then exit 24; fi
if desk switch noisy; then exit 25; fi
test "$PWD" = "$EXPECTED" || exit 26
desk info --path
desk info quiet --path
desk info --path quiet
builtin cd "$HOME"
if desk switch; then exit 27; fi
desk new >/dev/null || exit 28
test "$PWD" != "$HOME" || exit 29
desk version >/dev/null || exit 30
mkdir -p src/deep && builtin cd src/deep || exit 34
desk rename shell-prototype || exit 35
test "$PWD" = "${EXPECTED%/*}/shell-prototype/src/deep" || exit 36
desk finish || exit 37
desk rename noisy-fire other-task || exit 38
test "$PWD" = "${EXPECTED%/*}/shell-prototype/src/deep" || exit 39
printf 'y\n' > "$HOME/answer"
desk rm < "$HOME/answer" || exit 40
test "$PWD" = "$HOME" || exit 41
desk cleanup --yes || exit 42
printf 'integration-ok\n'
`
			if realShell == "fish" {
				script = `desk switch quiet; or exit 21
test "$PWD" = "$EXPECTED"; or exit 22
desk switch; or exit 23
if desk switch missing; exit 24; end
if desk switch noisy; exit 25; end
test "$PWD" = "$EXPECTED"; or exit 26
desk info --path
desk info quiet --path
desk info --path quiet
builtin cd "$HOME"
if desk switch; exit 27; end
desk new >/dev/null; or exit 28
test "$PWD" != "$HOME"; or exit 29
desk version >/dev/null; or exit 30
mkdir -p src/deep; and builtin cd src/deep; or exit 34
desk rename shell-prototype; or exit 35
test "$PWD" = (path dirname "$EXPECTED")/shell-prototype/src/deep; or exit 36
desk finish; or exit 37
desk rename noisy-fire other-task; or exit 38
test "$PWD" = (path dirname "$EXPECTED")/shell-prototype/src/deep; or exit 39
printf 'y\n' > "$HOME/answer"
desk rm < "$HOME/answer"; or exit 40
test "$PWD" = "$HOME"; or exit 41
desk cleanup --yes; or exit 42
printf 'integration-ok\n'
`
			}
			args := []string{"-ic", script}
			if realShell == "bash" {
				args = []string{"--noprofile", "-ic", script}
				if strings.HasPrefix(shell, "bash-login") {
					args = []string{"--login", "-ic", script}
				}
			}
			cmd := exec.Command(shellPath, args...)
			cmd.Dir = home
			var out, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("shell: %v\nstdout: %s\nstderr: %s", err, &out, &stderr)
			}
			if want := strings.Repeat(expected+"\n", 3) + "integration-ok\n"; out.String() != want {
				t.Fatalf("stdout %q; want %q; stderr %s", out.String(), want, &stderr)
			}
			for _, want := range []string{"no workspace", "matches noisy-fire, noisy-sky", "not inside a workspace"} {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("missing failure %q in %s", want, &stderr)
				}
			}
		})
	}
}
