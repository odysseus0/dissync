package sync

import (
	"context"
	"fmt"

	"github.com/tengjizhang/dissync/internal/db"
	"github.com/tengjizhang/dissync/internal/discord"
)

func backfillSync(ctx context.Context, client *discord.Client, store *db.DB, channelID string, state *ChannelState) error {
	if state.HistoryComplete {
		return nil
	}
	if state.BackfillBeforeID == "" {
		// No forward sync has produced a cursor yet.
		return nil
	}

	beforeID := state.BackfillBeforeID
	totalBackfill := 0

	for {
		msgs, err := client.GetMessagesBefore(ctx, channelID, beforeID, 100)
		if err != nil {
			return fmt.Errorf("backfill sync: %w", err)
		}

		if len(msgs) == 0 {
			state.HistoryComplete = true
			if err := store.SetSyncState(scopeComplete(channelID), "true"); err != nil {
				return err
			}
			break
		}

		ids := make([]string, len(msgs))
		for i, m := range msgs {
			ids[i] = m.ID
		}

		newBefore := discord.MinID(ids)
		cursors := map[string]string{
			scopeBackfill(channelID): newBefore,
		}

		if err := store.SavePageWithCursors(msgs, cursors); err != nil {
			return fmt.Errorf("save backfill page: %w", err)
		}

		state.BackfillBeforeID = newBefore
		totalBackfill += len(msgs)
		beforeID = newBefore

		fmt.Printf("  backfill: +%d messages (total backfill: %d)\n", len(msgs), totalBackfill)

		if len(msgs) < 100 {
			state.HistoryComplete = true
			if err := store.SetSyncState(scopeComplete(channelID), "true"); err != nil {
				return err
			}
			break
		}
	}

	return nil
}
