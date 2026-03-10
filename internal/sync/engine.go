package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/tengjizhang/dissync/internal/db"
	"github.com/tengjizhang/dissync/internal/discord"
)

type Engine struct {
	Client *discord.Client
	Store  *db.DB
}

func (e *Engine) SyncChannel(ctx context.Context, channelID string) error {
	state, err := LoadState(e.Store, channelID)
	if err != nil {
		return err
	}

	fmt.Printf("syncing channel %s (latest: %s, backfill: %s, complete: %v)\n",
		channelID, state.LatestMessageID, state.BackfillBeforeID, state.HistoryComplete)

	if err := forwardSync(ctx, e.Client, e.Store, channelID, &state); err != nil {
		return err
	}

	if err := backfillSync(ctx, e.Client, e.Store, channelID, &state); err != nil {
		return err
	}

	return nil
}

func (e *Engine) SyncChannels(ctx context.Context, channelIDs []string) error {
	for _, id := range channelIDs {
		if err := e.SyncChannel(ctx, id); err != nil {
			if errors.Is(err, discord.ErrForbidden) {
				fmt.Printf("skipping channel %s: no access\n", id)
				continue
			}
			if errors.Is(err, discord.ErrNotFound) {
				fmt.Printf("skipping channel %s: not found\n", id)
				continue
			}
			return err
		}
	}
	return nil
}

func (e *Engine) SyncGuild(ctx context.Context, guildID string) error {
	fmt.Printf("fetching channels for guild %s...\n", guildID)

	channels, err := e.Client.GetGuildChannels(ctx, guildID)
	if err != nil {
		return err
	}

	// Upsert all channels and collect message-bearing ones.
	var textChannelIDs []string
	var parentChannels []discord.Channel
	for _, ch := range channels {
		if err := e.Store.UpsertChannel(ch); err != nil {
			return err
		}
		if ch.HasMessages() {
			textChannelIDs = append(textChannelIDs, ch.ID)
		}
		if ch.Type == discord.ChannelTypeGuildText || ch.Type == discord.ChannelTypeGuildAnnouncement {
			parentChannels = append(parentChannels, ch)
		}
	}

	fmt.Printf("found %d text channels\n", len(textChannelIDs))

	// Discover threads.
	for _, parent := range parentChannels {
		for _, archived := range []bool{false, true} {
			threads, err := e.Client.SearchThreads(ctx, parent.ID, archived)
			if err != nil {
				fmt.Printf("warning: thread search failed for %s: %v\n", parent.ID, err)
				continue
			}
			for _, t := range threads {
				t.GuildID = guildID
				if err := e.Store.UpsertChannel(t); err != nil {
					return err
				}
				textChannelIDs = append(textChannelIDs, t.ID)
			}
		}
	}

	fmt.Printf("total channels+threads to sync: %d\n", len(textChannelIDs))

	return e.SyncChannels(ctx, textChannelIDs)
}
