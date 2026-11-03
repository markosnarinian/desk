package main

import (
	"os"
	"path/filepath"
	"strconv"
)

// /proc gives us working directories and open files without a watcher/daemon.
func busyPaths() ([]string, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		base := filepath.Join("/proc", e.Name())
		if cwd, err := os.Readlink(filepath.Join(base, "cwd")); err == nil {
			paths = append(paths, cwd)
		}
		fds, _ := os.ReadDir(filepath.Join(base, "fd"))
		for _, fd := range fds {
			if p, err := os.Readlink(filepath.Join(base, "fd", fd.Name())); err == nil && filepath.IsAbs(p) {
				paths = append(paths, p)
			}
		}
	}
	return paths, nil
}
