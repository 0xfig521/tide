package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/schedule"
)

var (
	schedInterval    string
	schedConcurrency string
	schedLogLines    int
	schedLogFollow   bool
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage background daemon for automatic feed fetching",
	Long: `Manage a background daemon that periodically fetches RSS feeds.

Commands:
  start   - Start the daemon in the background
  stop    - Stop the running daemon
  status  - Check if the daemon is running
  logs    - View recent daemon logs`,
}

var scheduleStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background fetch daemon",
	Long: `Start tide fetch --daemon as a detached background process.
The daemon will fetch RSS feeds at the specified interval.

Examples:
  tide schedule start
  tide schedule start --interval 1h --concurrency 10`,
	Run: func(cmd *cobra.Command, args []string) {
		mgr := schedule.New(scheduleDataDir())

		dbArg := ""
		if dataDir != "" {
			dbArg = dataDir + "/tide.db"
		} else {
			dbArg = dbPath
		}
		// Use relative interval syntax that fetch --daemon understands
		if schedInterval == "" {
			schedInterval = "30m"
		}
		if schedConcurrency == "" {
			schedConcurrency = "5"
		}

		if err := mgr.Start(dbArg, schedInterval, schedConcurrency); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), output.ErrorMsg(err.Error()))
			return
		}

		fmt.Println(output.Success(fmt.Sprintf("Daemon started. Logs: %s", mgr.LogPath())))
		fmt.Println(output.Warn("Use 'tide schedule status' to check, 'tide schedule stop' to terminate."))
	},
}

var scheduleStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running fetch daemon",
	Long:  "Gracefully stop the background daemon process.",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := schedule.New(scheduleDataDir())

		if err := mgr.Stop(); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), output.ErrorMsg(err.Error()))
			return
		}

		fmt.Println(output.Success("Daemon stopped."))
	},
}

var scheduleStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	Long:  "Check whether the background daemon is running and show its uptime.",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := schedule.New(scheduleDataDir())

		running, pid, uptime := mgr.Status()
		if !running {
			fmt.Println(output.Warn("Daemon is not running."))
			return
		}

		status := fmt.Sprintf("Daemon is running (PID: %d)", pid)
		if uptime != "" {
			status += fmt.Sprintf(", uptime: %s", uptime)
		}
		fmt.Println(output.Success(status))
	},
}

var scheduleLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View daemon logs",
	Long: `View recent log output from the daemon process.

Examples:
  tide schedule logs          # Show all logs
  tide schedule logs -n 20    # Show last 20 lines
  tide schedule logs --clear  # Clear log file`,
	Run: func(cmd *cobra.Command, args []string) {
		mgr := schedule.New(scheduleDataDir())

		if clearLogs, _ := cmd.Flags().GetBool("clear"); clearLogs {
			if err := mgr.ClearLogs(); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), output.ErrorMsg(fmt.Sprintf("clear logs: %v", err)))
				return
			}
			fmt.Println(output.Success("Logs cleared."))
			return
		}

		lines, err := mgr.Logs(schedLogLines)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), output.ErrorMsg(fmt.Sprintf("read logs: %v", err)))
			return
		}

		if len(lines) == 0 {
			fmt.Println(output.Warn("No log entries yet."))
			return
		}

		for _, line := range lines {
			fmt.Println(line)
		}
	},
}

func init() {
	scheduleStartCmd.Flags().StringVar(&schedInterval, "interval", "", "Fetch interval (e.g. 30m, 1h)")
	scheduleStartCmd.Flags().StringVar(&schedConcurrency, "concurrency", "", "Number of concurrent workers")

	scheduleLogsCmd.Flags().IntVarP(&schedLogLines, "lines", "n", 0, "Show last N lines (0 = all)")
	scheduleLogsCmd.Flags().Bool("clear", false, "Clear log file")

	scheduleCmd.AddCommand(scheduleStartCmd)
	scheduleCmd.AddCommand(scheduleStopCmd)
	scheduleCmd.AddCommand(scheduleStatusCmd)
	scheduleCmd.AddCommand(scheduleLogsCmd)
	rootCmd.AddCommand(scheduleCmd)
}

// scheduleDataDir returns the directory for schedule-related data (pid file, logs).
func scheduleDataDir() string {
	if dataDir != "" {
		return dataDir
	}
	return filepath.Dir(dbPath)
}
