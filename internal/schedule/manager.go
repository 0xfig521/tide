package schedule

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Manager manages a daemon process lifecycle.
type Manager struct {
	dataDir string
}

// New creates a Schedule manager.
func New(dataDir string) *Manager {
	return &Manager{dataDir: dataDir}
}

func (m *Manager) pidPath() string {
	return filepath.Join(m.dataDir, "daemon.pid")
}

func (m *Manager) logPath() string {
	logsDir := filepath.Join(m.dataDir, "logs")
	os.MkdirAll(logsDir, 0755)
	return filepath.Join(logsDir, "daemon.log")
}

// Start launches tide fetch --daemon as a detached background process.
func (m *Manager) Start(dbPath, interval, concurrency string) error {
	if m.isRunning() {
		return errors.New("daemon is already running")
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find tide executable: %w", err)
	}

	logPath := m.logPath()
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	args := []string{"fetch", "--daemon"}
	if dbPath != "" {
		args = append(args, "--db", dbPath)
	}
	if interval != "" {
		args = append(args, "--interval", interval)
	}
	if concurrency != "" {
		args = append(args, "--concurrency", concurrency)
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Detach: create new session so the process survives parent exit
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start daemon: %w", err)
	}

	// Write PID file
	if err := os.WriteFile(m.pidPath(), []byte(strconv.Itoa(cmd.Process.Pid)+"\n"), 0644); err != nil {
		// Try to kill the started process on failure
		cmd.Process.Kill()
		logFile.Close()
		return fmt.Errorf("write pid file: %w", err)
	}

	// Close the log file in the parent; child inherited the fd
	logFile.Close()

	return nil
}

// Stop terminates a running daemon gracefully.
func (m *Manager) Stop() error {
	pid, err := m.readPID()
	if err != nil {
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(m.pidPath())
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Process may already be dead
		if errors.Is(err, os.ErrProcessDone) {
			os.Remove(m.pidPath())
			return nil
		}
		return fmt.Errorf("signal daemon: %w", err)
	}

	// Wait for graceful shutdown with timeout
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		proc.Kill()
		<-done
	}

	os.Remove(m.pidPath())
	return nil
}

// Status returns the daemon's running state and PID.
func (m *Manager) Status() (running bool, pid int, uptime string) {
	pid, err := m.readPID()
	if err != nil {
		return false, 0, ""
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(m.pidPath())
		return false, 0, ""
	}

	// Signal 0 checks existence without delivering a signal
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		os.Remove(m.pidPath())
		return false, 0, ""
	}

	// Try to get process start time for uptime
	uptime = m.processUptime(pid)
	return true, pid, uptime
}

// Logs returns the last n lines of the daemon log.
func (m *Manager) Logs(n int) ([]string, error) {
	f, err := os.Open(m.logPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase buffer for long lines
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if n > 0 && len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, scanner.Err()
}

// LogPath returns the path to the daemon log file.
func (m *Manager) LogPath() string {
	return m.logPath()
}

// ClearLogs truncates the daemon log file.
func (m *Manager) ClearLogs() error {
	return os.Truncate(m.logPath(), 0)
}

func (m *Manager) readPID() (int, error) {
	data, err := os.ReadFile(m.pidPath())
	if err != nil {
		if os.IsNotExist(err) {
			return 0, errors.New("daemon is not running (no pid file)")
		}
		return 0, fmt.Errorf("read pid file: %w", err)
	}
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid pid file: %w", err)
	}
	return pid, nil
}

func (m *Manager) isRunning() bool {
	running, _, _ := m.Status()
	return running
}

func (m *Manager) processUptime(pid int) string {
	// Read process start time from /proc on Linux, or use ps on macOS
	// Use ps command which works on both platforms
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=").Output()
	if err != nil {
		return ""
	}
	startStr := strings.TrimSpace(string(out))
	if startStr == "" {
		return ""
	}
	// Parse in local timezone since ps outputs local time
	t, err := time.ParseInLocation("Mon Jan 2 15:04:05 2006", startStr, time.Local)
	if err != nil {
		// Try with double-space padding for single-digit days
		t, err = time.ParseInLocation("Mon Jan  2 15:04:05 2006", startStr, time.Local)
		if err != nil {
			return ""
		}
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours < 24 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}
