package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, key := range []string{"ZSH", "ZSH_CUSTOM", "ZDOTDIR", "XDG_CONFIG_HOME", "SHELL"} {
		t.Setenv(key, "")
	}
	return home
}

func mustWrite(t *testing.T, path, data string) {
	t.Helper()
	if err := writeConfig(path, data); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestInitInstallsPreservingAndIdempotent(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			home := isolatedHome(t)
			t.Setenv("SHELL", "/usr/local/bin/"+shell)
			startup := filepath.Join(home, ".bashrc")
			if shell == "zsh" {
				t.Setenv("ZDOTDIR", filepath.Join(home, "zsh configs"))
				startup = filepath.Join(os.Getenv("ZDOTDIR"), ".zshrc")
			} else if shell == "fish" {
				t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg configs"))
				startup = filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "fish", "config.fish")
			}
			original := "# preserve my settings\nset -g example yes\n# no final newline"
			mustWrite(t, startup, original)
			if err := os.Chmod(startup, 0600); err != nil {
				t.Fatal(err)
			}
			message, _, err := invoke(t, "", "init")
			if err != nil || !strings.Contains(message, "Open a new terminal") {
				t.Fatalf("init: %s %v", message, err)
			}
			first := mustRead(t, startup)
			executable, _ := os.Executable()
			if !strings.HasPrefix(first, original+"\n") || !strings.Contains(first, shellQuote(executable, shell)+" init "+shell+" --print") {
				t.Fatalf("startup content: %s", first)
			}
			if _, _, err := invoke(t, "", "init", shell); err != nil {
				t.Fatal(err)
			}
			if got := mustRead(t, startup); got != first {
				t.Fatalf("init not idempotent:\n%s\nvs\n%s", first, got)
			}
			stat, _ := os.Stat(startup)
			if stat.Mode().Perm() != 0600 {
				t.Fatal("existing file permissions changed")
			}
		})
	}
}

func TestInitPrintDoesNotWrite(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "oh-my-zsh"} {
		t.Run(shell, func(t *testing.T) {
			home := isolatedHome(t)
			for _, args := range [][]string{{"init", shell, "--print"}, {"init", "--print", shell}} {
				script, _, err := invoke(t, "", args...)
				if err != nil || !strings.Contains(script, "desk info --path") || !strings.Contains(script, "PATH") {
					t.Fatalf("print: %s %v", script, err)
				}
			}
			entries, _ := os.ReadDir(home)
			if len(entries) != 0 {
				t.Fatalf("--print wrote files: %v", entries)
			}
		})
	}
}

func TestInitOhMyZsh(t *testing.T) {
	for _, mode := range []string{"explicit", "ZSH", "source", "custom", "already-enabled", "no-source"} {
		t.Run(mode, func(t *testing.T) {
			home := isolatedHome(t)
			startup := filepath.Join(home, ".zshrc")
			custom := filepath.Join(home, ".oh-my-zsh", "custom")
			shell := "oh-my-zsh"
			original := "# my settings\nplugins=(git docker)\nsource $ZSH/oh-my-zsh.sh\n# keep me\n"
			switch mode {
			case "ZSH":
				shell = "zsh"
				t.Setenv("ZSH", filepath.Join(home, "my omz"))
				custom = filepath.Join(os.Getenv("ZSH"), "custom")
			case "source":
				shell = "zsh"
			case "custom":
				custom = filepath.Join(home, "custom's plugins")
				t.Setenv("ZSH_CUSTOM", custom)
			case "already-enabled":
				original = strings.ReplaceAll(original, "git docker", "git desk docker")
			case "no-source":
				original = "plugins=(git docker)\n"
			}
			mustWrite(t, startup, original)
			if _, _, err := invoke(t, "", "init", shell); err != nil {
				t.Fatal(err)
			}
			first := mustRead(t, startup)
			plugin := filepath.Join(custom, "plugins", "desk", "desk.plugin.zsh")
			if !strings.Contains(mustRead(t, plugin), " init zsh --print") {
				t.Fatal("plugin does not load integration")
			}
			clean, err := withoutManagedBlock(first)
			if err != nil || clean != original {
				t.Fatalf("original settings changed: %q %v", clean, err)
			}
			if strings.Contains(original, "oh-my-zsh.sh") && strings.Index(first, blockStart) < strings.Index(first, "source $ZSH") {
				t.Fatal("managed block installed before OMZ loader")
			}
			if mode == "no-source" && !strings.Contains(first, "source "+shellQuote(plugin, "zsh")) {
				t.Fatal("missing direct plugin loading fallback")
			}
			if _, _, err := invoke(t, "", "init", shell); err != nil {
				t.Fatal(err)
			}
			if mustRead(t, startup) != first {
				t.Fatal("OMZ init not idempotent")
			}
		})
	}
}

func TestInitFailuresLeaveStartupAlone(t *testing.T) {
	home := isolatedHome(t)
	startup := filepath.Join(home, ".zshrc")
	original := "# existing\n" + blockStart + "\n# incomplete block\n"
	mustWrite(t, startup, original)
	for _, args := range [][]string{{"init", "zsh"}, {"init", "powershell"}, {"init"}, {"init", "--bad"}, {"init", "bash", "zsh"}, {"init", "--print", "--print"}} {
		if _, _, err := invoke(t, "", args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
		if mustRead(t, startup) != original {
			t.Fatalf("%v changed startup on error", args)
		}
	}
	// An unwritable target is reported, not silently treated as an empty file.
	if err := os.Mkdir(filepath.Join(home, ".bashrc"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := invoke(t, "", "init", "bash"); err == nil {
		t.Fatal("init succeeded with a directory for .bashrc")
	}
}

func TestBashLoginPrecedenceAndMigration(t *testing.T) {
	for _, selected := range []string{".bash_profile", ".bash_login", ".profile", "none"} {
		t.Run(selected, func(t *testing.T) {
			home := isolatedHome(t)
			files := []string{".bash_profile", ".bash_login", ".profile"}
			create := false
			for _, name := range files {
				create = create || name == selected
				if create {
					mustWrite(t, filepath.Join(home, name), "# preserve "+name+"\neval \"$(desk init bash)\"\n")
				}
			}
			if selected == "none" {
				selected = ".bash_profile"
			}
			if _, _, err := invoke(t, "", "init", "bash"); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(home, selected)
			first := mustRead(t, path)
			if !strings.Contains(first, "init bash --print") || strings.Contains(first, `$(desk init bash)`) || !strings.Contains(first, `${BASH_VERSION-}`) {
				t.Fatalf("login content: %s", first)
			}
			for _, name := range files {
				if name == selected {
					continue
				}
				content, err := readStartup(filepath.Join(home, name))
				if err != nil || strings.Contains(content, blockStart) {
					t.Fatalf("changed nonselected profile %s: %s %v", name, content, err)
				}
			}
			if _, _, err := invoke(t, "", "init", "bash"); err != nil {
				t.Fatal(err)
			}
			if mustRead(t, path) != first {
				t.Fatal("login initialization is not idempotent")
			}
		})
	}
}

func TestLegacyLoaderMigration(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "oh-my-zsh"} {
		t.Run(shell, func(t *testing.T) {
			home := isolatedHome(t)
			path := filepath.Join(home, ".zshrc")
			legacy := "  eval \"$(desk init zsh)\"\n"
			if shell == "bash" {
				path = filepath.Join(home, ".bashrc")
				legacy = "eval \"$(desk init bash)\"\n"
			} else if shell == "fish" {
				path = filepath.Join(home, ".config", "fish", "config.fish")
				legacy = "desk init fish | source\n"
			}
			unrelated := "# keep these settings\n# eval \"$(desk init zsh)\"\n"
			mustWrite(t, path, unrelated+legacy)
			plugin := filepath.Join(home, ".oh-my-zsh", "custom", "plugins", "desk", "desk.plugin.zsh")
			if shell == "oh-my-zsh" {
				mustWrite(t, plugin, unrelated+legacy+"# other plugin settings\n")
			}
			if _, _, err := invoke(t, "", "init", shell); err != nil {
				t.Fatal(err)
			}
			first := mustRead(t, path)
			clean, err := withoutManagedBlock(first)
			if err != nil || clean != unrelated {
				t.Fatalf("migration changed unrelated content: %q %v", clean, err)
			}
			if shell == "oh-my-zsh" {
				clean, err := withoutManagedBlock(mustRead(t, plugin))
				if err != nil || clean != unrelated+"# other plugin settings\n" {
					t.Fatalf("plugin migration: %q %v", clean, err)
				}
			}
			if _, _, err := invoke(t, "", "init", shell); err != nil || mustRead(t, path) != first {
				t.Fatalf("migration not idempotent: %v", err)
			}
		})
	}
	for _, line := range []string{`if true; then eval "$(desk init zsh)"; fi`, `eval "$(desk init zsh)" # custom comment`, `eval "$(/custom/desk init zsh)"`, `# desk init fish | source`, `eval "$(desk init zsh --print)"`} {
		if got := legacyLoader.ReplaceAllString(line+"\n", ""); got != line+"\n" {
			t.Errorf("rewrote unrecognized shell code %q", line)
		}
	}
}
