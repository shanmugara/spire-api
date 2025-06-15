package spire_grpc

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	pgrepPath = "/usr/bin/pgrep"
)

func findPIDsByName(name string) ([]int, error) {
	cmd := exec.Command(pgrepPath, name)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run pgrep: %v", err)
	}

	var pids []int
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}
