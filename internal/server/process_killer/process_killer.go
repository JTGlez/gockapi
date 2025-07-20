package process_killer

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessKiller defines the interface for killing a process on a given port.
type ProcessKiller interface {
	KillProcessOnPort(port int) error
}

// LinuxProcessKiller implements ProcessKiller for Linux systems.
type LinuxProcessKiller struct{}

func NewLinuxProcessKiller() *LinuxProcessKiller {
	return &LinuxProcessKiller{}
}

// KillProcessOnPort finds and kills the process listening on the given port if it is a gockapi process.
func (pk *LinuxProcessKiller) KillProcessOnPort(port int) error {
	// Use lsof to find the PID listening on the port
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-t")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not find process on port %d: %v", port, out.String())
	}

	pidStr := strings.TrimSpace(out.String())
	if pidStr == "" {
		return fmt.Errorf("no process found listening on port %d", port)
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID found: %s", pidStr)
	}

	// Check the process command line to ensure it's gockapi
	cmdline, err := exec.Command("cat", fmt.Sprintf("/proc/%d/cmdline", pid)).Output()
	if err != nil {
		return fmt.Errorf("could not read cmdline for pid %d: %v", pid, err)
	}
	if !bytes.Contains(cmdline, []byte("gockapi")) {
		return fmt.Errorf("process on port %d is not a gockapi instance", port)
	}

	// Kill the process
	killCmd := exec.Command("kill", strconv.Itoa(pid))
	if err := killCmd.Run(); err != nil {
		return fmt.Errorf("failed to kill process %d: %v", pid, err)
	}

	return nil
}
