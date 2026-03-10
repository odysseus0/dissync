package sync

import (
	"fmt"

	"github.com/tengjizhang/dissync/internal/db"
)

type ChannelState struct {
	LatestMessageID  string
	BackfillBeforeID string
	HistoryComplete  bool
}

func scopeLatest(channelID string) string {
	return fmt.Sprintf("channel:%s:latest_message_id", channelID)
}

func scopeBackfill(channelID string) string {
	return fmt.Sprintf("channel:%s:backfill_before_id", channelID)
}

func scopeComplete(channelID string) string {
	return fmt.Sprintf("channel:%s:history_complete", channelID)
}

func LoadState(store *db.DB, channelID string) (ChannelState, error) {
	latest, err := store.GetSyncState(scopeLatest(channelID))
	if err != nil {
		return ChannelState{}, err
	}
	backfill, err := store.GetSyncState(scopeBackfill(channelID))
	if err != nil {
		return ChannelState{}, err
	}
	complete, err := store.GetSyncState(scopeComplete(channelID))
	if err != nil {
		return ChannelState{}, err
	}
	return ChannelState{
		LatestMessageID:  latest,
		BackfillBeforeID: backfill,
		HistoryComplete:  complete == "true",
	}, nil
}
