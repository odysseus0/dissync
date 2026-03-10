package sync

import (
	"context"
	"fmt"

	"github.com/tengjizhang/dissync/internal/db"
	"github.com/tengjizhang/dissync/internal/discord"
)

func forwardSync(ctx context.Context, client *discord.Client, store *db.DB, channelID string, state *ChannelState) error {
	afterID := state.LatestMessageID
	totalNew := 0

	for {
		msgs, err := client.GetMessagesAfter(ctx, channelID, afterID, 100)
		if err != nil {
			return fmt.Errorf("forward sync: %w", err)
		}
		if len(msgs) == 0 {
			// Empty channel on first sync.
			if state.LatestMessageID == "" && state.BackfillBeforeID == "" {
				state.HistoryComplete = true
				if err := store.SetSyncState(scopeComplete(channelID), "true"); err != nil {
					return err
				}
			}
			break
		}

		ids := make([]string, len(msgs))
		for i, m := range msgs {
			ids[i] = m.ID
		}

		newLatest := discord.MaxID(ids)
		cursors := map[string]string{
			scopeLatest(channelID): newLatest,
		}

		// Seed backfill cursor on first forward page.
		if state.BackfillBeforeID == "" {
			minID := discord.MinID(ids)
			state.BackfillBeforeID = minID
			cursors[scopeBackfill(channelID)] = minID
		}

		if err := store.SavePageWithCursors(msgs, cursors); err != nil {
			return fmt.Errorf("save forward page: %w", err)
		}

		state.LatestMessageID = newLatest
		totalNew += len(msgs)
		afterID = newLatest

		fmt.Printf("  forward: +%d messages (total new: %d)\n", len(msgs), totalNew)

		if len(msgs) < 100 {
			break
		}
	}

	return nil
}
