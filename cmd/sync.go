package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tengjizhang/dissync/internal/sync"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync Discord channels to local SQLite",
	RunE: func(cmd *cobra.Command, args []string) error {
		channelIDs, _ := cmd.Flags().GetStringSlice("channel")
		guildID, _ := cmd.Flags().GetString("guild")

		if len(channelIDs) == 0 && guildID == "" {
			return fmt.Errorf("specify --channel or --guild")
		}

		client, err := newClient()
		if err != nil {
			return err
		}
		store, err := openDB()
		if err != nil {
			return err
		}
		defer store.Close()

		engine := &sync.Engine{Client: client, Store: store}

		if guildID != "" {
			return engine.SyncGuild(cmd.Context(), guildID)
		}
		return engine.SyncChannels(cmd.Context(), channelIDs)
	},
}

func init() {
	syncCmd.Flags().StringSlice("channel", nil, "Channel ID(s) to sync (repeatable)")
	syncCmd.Flags().String("guild", "", "Guild ID to sync all channels")
}
