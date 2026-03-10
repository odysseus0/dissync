package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openDB()
		if err != nil {
			return err
		}
		defer store.Close()

		stats, err := store.GetStats()
		if err != nil {
			return err
		}

		if len(stats) == 0 {
			fmt.Println("No synced channels yet.")
			return nil
		}

		totalMsgs := 0
		for _, s := range stats {
			complete := " "
			if s.HistoryComplete {
				complete = "+"
			}
			fmt.Printf("[%s] %-30s %6d msgs  %s — %s\n",
				complete,
				s.ChannelName,
				s.MessageCount,
				s.OldestMessage.Format("2006-01-02"),
				s.NewestMessage.Format("2006-01-02"),
			)
			totalMsgs += s.MessageCount
		}
		fmt.Printf("\n%d channels, %d messages total\n", len(stats), totalMsgs)
		return nil
	},
}
