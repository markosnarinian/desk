package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func optionalFlag(args []string, flag string) (string, bool, error) {
	name, found := "", false
	for _, arg := range args {
		if arg == flag && !found {
			found = true
		} else if !strings.HasPrefix(arg, "-") && name == "" {
			name = arg
		} else {
			return "", false, fmt.Errorf("unexpected argument %q; see desk help", arg)
		}
	}
	return name, found, nil
}

func shellQuote(s, shell string) string {
	if shell == "fish" {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), "'", `\'`) + "'"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func integration(shell, executable string) (string, error) {
	if shell == "oh-my-zsh" {
		shell = "zsh"
	}
	b, err := shells.ReadFile("shell/desk." + shell)
	if err != nil {
		return "", fmt.Errorf("unsupported shell %q; use zsh, bash, fish, or oh-my-zsh", shell)
	}
	bin := shellQuote(filepath.Dir(executable), shell)
	// The executable need not have been on PATH when init was invoked.
	prefix := "case :$PATH: in *:" + bin + ":*) ;; *) export PATH=" + bin + ":\"$PATH\" ;; esac\n"
	if shell == "fish" {
		prefix = "if not contains -- " + bin + " $PATH\n    set -gx PATH " + bin + " $PATH\nend\n"
	}
	return prefix + string(b), nil
}

const blockStart = "# >>> desk shell integration >>>"
const blockEnd = "# <<< desk shell integration <<<"

// Replace only our own block, leaving all other bytes intact.
func withoutManagedBlock(s string) (string, error) {
	start := strings.Index(s, blockStart)
	if start < 0 {
		if strings.Contains(s, blockEnd) {
			return "", errors.New("incomplete desk managed block; repair startup config first")
		}
		return s, nil
	}
	end := strings.Index(s[start:], blockEnd)
	if end < 0 {
		return "", errors.New("incomplete desk managed block; repair startup config first")
	}
	end += start + len(blockEnd)
	if end < len(s) && s[end] == '\n' {
		end++
	}
	remaining := s[:start] + s[end:]
	if strings.Contains(remaining, blockStart) || strings.Contains(remaining, blockEnd) {
		return "", errors.New("multiple desk managed blocks; repair startup config first")
	}
	return remaining, nil
}

// Match an active source command, not a comment mentioning Oh My Zsh.
var omzSource = regexp.MustCompile(`(?m)^[\t ]*(?:source|\.)[\t ]+[^\n;]*oh-my-zsh\.sh[^\n]*`)

// Only remove the standalone loaders previously documented by desk. Do not
// rewrite arbitrary shell code, conditional commands, or commented examples.
var legacyLoader = regexp.MustCompile(`(?m)^[\t ]*(?:eval "\$\((?:command[\t ]+)?desk init (?:bash|zsh|oh-my-zsh)\)"|(?:command[\t ]+)?desk init fish[\t ]*\|[\t ]*source)[\t ]*;?[\t ]*(?:\n|$)`)

func appendLoadingBlock(existing, load string) (string, error) {
	existing, err := withoutManagedBlock(existing)
	if err != nil {
		return "", err
	}
	existing = legacyLoader.ReplaceAllString(existing, "")
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	return existing + blockStart + "\n" + load + blockEnd + "\n", nil
}

func bashLoginFile(home string) (string, error) {
	for _, name := range []string{".bash_profile", ".bash_login", ".profile"} {
		path := filepath.Join(home, name)
		file, err := os.Open(path)
		if err == nil {
			file.Close()
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrPermission) {
			return "", err
		}
	}
	return filepath.Join(home, ".bash_profile"), nil
}

func readStartup(path string) (string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return string(b), err
}

func writeConfig(path, data string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0644)
}

func initShell(args []string, out io.Writer) error {
	shell, printOnly, err := optionalFlag(args, "--print")
	if err != nil {
		return err
	}
	if shell == "" {
		shell = filepath.Base(os.Getenv("SHELL"))
	}
	if shell != "bash" && shell != "zsh" && shell != "fish" && shell != "oh-my-zsh" {
		return fmt.Errorf("unsupported shell %q; use zsh, bash, fish, or oh-my-zsh", shell)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	script, err := integration(shell, executable)
	if err != nil {
		return err
	}
	if printOnly {
		_, err = io.WriteString(out, script)
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	startup := filepath.Join(home, ".bashrc")
	if shell == "zsh" || shell == "oh-my-zsh" {
		zdotdir := os.Getenv("ZDOTDIR")
		if zdotdir == "" {
			zdotdir = home
		}
		startup = filepath.Join(zdotdir, ".zshrc")
	} else if shell == "fish" {
		config := os.Getenv("XDG_CONFIG_HOME")
		if config == "" {
			config = filepath.Join(home, ".config")
		}
		startup = filepath.Join(config, "fish", "config.fish")
	}
	existing, err := readStartup(startup)
	if err != nil {
		return err
	}
	existing, err = withoutManagedBlock(existing)
	if err != nil {
		return err
	}
	existing = legacyLoader.ReplaceAllString(existing, "")
	isOMZ := shell == "oh-my-zsh" || (shell == "zsh" && (os.Getenv("ZSH") != "" || omzSource.MatchString(existing)))
	load := "eval \"$(" + shellQuote(executable, shell) + " init " + shell + " --print)\"\n"
	if shell == "fish" {
		load = shellQuote(executable, shell) + " init fish --print | source\n"
	}
	if isOMZ {
		custom := os.Getenv("ZSH_CUSTOM")
		if custom == "" {
			zsh := os.Getenv("ZSH")
			if zsh == "" {
				zsh = filepath.Join(home, ".oh-my-zsh")
			}
			custom = filepath.Join(zsh, "custom")
		}
		plugin := filepath.Join(custom, "plugins", "desk", "desk.plugin.zsh")
		pluginLoader := "eval \"$(" + shellQuote(executable, "zsh") + " init zsh --print)\"\n"
		// Load after the framework and theme, without changing the user's plugin
		// list or shell-local ZSH_CUSTOM (which a child process cannot inspect).
		load = "source " + shellQuote(plugin, "zsh") + "\n"
		pluginContent, err := readStartup(plugin)
		if err != nil {
			return err
		}
		pluginContent, err = appendLoadingBlock(pluginContent, pluginLoader)
		if err != nil {
			return err
		}
		if err := writeConfig(plugin, pluginContent); err != nil {
			return err
		}
	}
	existing, err = appendLoadingBlock(existing, load)
	if err != nil {
		return err
	}
	if err := writeConfig(startup, existing); err != nil {
		return err
	}
	if shell == "bash" {
		login, err := bashLoginFile(home)
		if err != nil {
			return err
		}
		content, err := readStartup(login)
		if err != nil {
			return err
		}
		// Install directly in the actual login file. Do not source all of .bashrc:
		// users may intentionally keep its unrelated settings non-login-only.
		// .profile may also be read by POSIX shells: keep the guard portable.
		content, err = appendLoadingBlock(content, "if [ -n \"${BASH_VERSION-}\" ] && [ \"${-#*i}\" != \"$-\" ] && ! declare -F _desk_prompt >/dev/null; then\n  "+load+"fi\n")
		if err != nil {
			return err
		}
		if err := writeConfig(login, content); err != nil {
			return err
		}
	}
	// Also harmless if an old eval/pipe-to-source loader invokes installing init
	// once during migration: status text must not be executed as shell commands.
	fmt.Fprintf(out, "# desk: installed %s integration in %s. Open a new terminal to use desk switch.\n", shell, startup)
	return nil
}
