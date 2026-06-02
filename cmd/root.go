package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	dbPath  string
	dataDir string

	dbConn *db.DB

	// set by ldflags at build time
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "tide",
	Short: "A high-concurrency RSS reader for the terminal",
	Long: `tide is a fast, concurrent RSS reader CLI built in Go.
It uses SQLite for storage and supports categories, search, and more.

All commands output JSON to stdout by default. Progress, logs, and diagnostics
go to stderr.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	defaultDBPath := db.DefaultDBPath()
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", defaultDBPath, "Database path")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "Data directory (overrides --db)")

	cobra.OnInitialize(initDB)
}

func initDB() {
	if dataDir != "" {
		dbPath = dataDir + "/tide.db"
	}
	var err error
	dbConn, err = db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
}

func feedRepo() *repo.FeedRepo           { return repo.NewFeedRepo(dbConn) }
func categoryRepo() *repo.CategoryRepo   { return repo.NewCategoryRepo(dbConn) }
func entryRepo() *repo.EntryRepo         { return repo.NewEntryRepo(dbConn) }
func stateRepo() *repo.EntryStateRepo    { return repo.NewEntryStateRepo(dbConn) }
func ruleRepo() *repo.RuleRepo           { return repo.NewRuleRepo(dbConn) }
func failureRepo() *repo.FeedFailureRepo { return repo.NewFeedFailureRepo(dbConn) }

// parseIDArg parses a string ID argument; returns an error via RunE on failure.
func parseIDArg(arg string) (int64, error) {
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, output.PrintError(output.CodeInvalidArgs, fmt.Sprintf("invalid ID: %s", arg))
	}
	return id, nil
}
