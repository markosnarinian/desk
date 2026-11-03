package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func busyPaths() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/sbin/lsof", "-n", "-P", "-F", "n")
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lsof failed: %w", err)
	}
	var paths []string
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "n/") {
			paths = append(paths, strings.TrimPrefix(line, "n"))
		}
	}
	return paths, nil
}
