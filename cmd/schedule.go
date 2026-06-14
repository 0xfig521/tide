package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
	"github.com/0xfig-labs/tide/internal/schedule"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := schedule.New(scheduleDataDir())

		dbArg := ""
		if dataDir != "" {
			dbArg = dataDir + "/tide.db"
		} else {
			dbArg = dbPath
		}
		if schedInterval == "" {
			schedInterval = "30m"
		}
		if schedConcurrency == "" {
			schedConcurrency = "5"
		}

		if err := mgr.Start(dbArg, schedInterval, schedConcurrency); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		output.PrintSuccess(map[string]any{"status": "started", "log_path": mgr.LogPath()}, nil)
		return nil
	},
}

var scheduleStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running fetch daemon",
	Long:  "Gracefully stop the background daemon process.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := schedule.New(scheduleDataDir())

		if err := mgr.Stop(); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		output.PrintSuccess(map[string]any{"status": "stopped"}, nil)
		return nil
	},
}

var scheduleStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	Long:  "Check whether the background daemon is running and show its uptime.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := schedule.New(scheduleDataDir())

		running, pid, uptime := mgr.Status()
		if !running {
			output.PrintSuccess(map[string]any{"running": false}, nil)
			return nil
		}

		output.PrintSuccess(map[string]any{"running": true, "pid": pid, "uptime": uptime}, nil)
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := schedule.New(scheduleDataDir())

		if clearLogs, _ := cmd.Flags().GetBool("clear"); clearLogs {
			if err := mgr.ClearLogs(); err != nil {
				return output.PrintError(output.CodeInternalError, err.Error())
			}
			output.PrintSuccess(map[string]any{"status": "cleared"}, nil)
			return nil
		}

		lines, err := mgr.Logs(schedLogLines)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		if len(lines) == 0 {
			output.PrintSuccess(map[string]any{"lines": []string{}}, nil)
			return nil
		}

		output.PrintSuccess(map[string]any{"lines": lines}, nil)
		return nil
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
