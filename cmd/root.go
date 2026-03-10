package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tengjizhang/dissync/internal/db"
	"github.com/tengjizhang/dissync/internal/discord"
)

var (
	tokenFlag string
	dbFlag    string
)

var rootCmd = &cobra.Command{
	Use:   "dissync",
	Short: "Incrementally sync Discord channels to local SQLite",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&tokenFlag, "token", "", "Discord user token (or set DISCORD_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&dbFlag, "db", "", "SQLite database path (default: ~/.dissync/dissync.db)")

	rootCmd.AddCommand(guildsCmd)
	rootCmd.AddCommand(channelsCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statusCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func resolveToken() (string, error) {
	if tokenFlag != "" {
		return tokenFlag, nil
	}
	if t := os.Getenv("DISCORD_TOKEN"); t != "" {
		return t, nil
	}
	return "", fmt.Errorf("no token: set DISCORD_TOKEN or use --token")
}

func resolveDBPath() string {
	if dbFlag != "" {
		return dbFlag
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dissync", "dissync.db")
}

func newClient() (*discord.Client, error) {
	token, err := resolveToken()
	if err != nil {
		return nil, err
	}
	return discord.NewClient(token), nil
}

func openDB() (*db.DB, error) {
	return db.Open(resolveDBPath())
}
