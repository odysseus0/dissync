package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var guildsCmd = &cobra.Command{
	Use:   "guilds",
	Short: "List accessible Discord servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		store, err := openDB()
		if err != nil {
			return err
		}
		defer store.Close()

		guilds, err := client.GetGuilds(cmd.Context())
		if err != nil {
			return err
		}

		for _, g := range guilds {
			if err := store.UpsertGuild(g); err != nil {
				return err
			}
			fmt.Printf("%s | %s\n", g.ID, g.Name)
		}

		fmt.Printf("\n%d guilds total\n", len(guilds))
		return nil
	},
}
