package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "List channels in a guild",
	RunE: func(cmd *cobra.Command, args []string) error {
		guildID, _ := cmd.Flags().GetString("guild")
		if guildID == "" {
			return fmt.Errorf("--guild is required")
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

		channels, err := client.GetGuildChannels(cmd.Context(), guildID)
		if err != nil {
			return err
		}

		typeNames := map[int]string{
			0: "text", 2: "voice", 4: "category", 5: "announcement",
			10: "ann-thread", 11: "pub-thread", 12: "priv-thread",
			13: "stage", 15: "forum", 16: "media",
		}

		for _, ch := range channels {
			if err := store.UpsertChannel(ch); err != nil {
				return err
			}
			typeName := typeNames[ch.Type]
			if typeName == "" {
				typeName = fmt.Sprintf("type-%d", ch.Type)
			}
			fmt.Printf("%s | %-12s | %s\n", ch.ID, typeName, ch.Name)
		}

		fmt.Printf("\n%d channels total\n", len(channels))
		return nil
	},
}

func init() {
	channelsCmd.Flags().String("guild", "", "Guild ID")
}
