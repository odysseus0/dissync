package db

import (
	"fmt"
	"time"

	"github.com/tengjizhang/dissync/internal/discord"
)

func (d *DB) UpsertGuild(g discord.Guild) error {
	_, err := d.conn.Exec(`
		INSERT INTO guilds (id, name, icon, fetched_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, icon=excluded.icon, fetched_at=excluded.fetched_at`,
		g.ID, g.Name, g.Icon, time.Now().UnixMilli())
	return err
}

func (d *DB) UpsertChannel(ch discord.Channel) error {
	_, err := d.conn.Exec(`
		INSERT INTO channels (id, guild_id, name, type, parent_id, position, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			guild_id=excluded.guild_id, name=excluded.name, type=excluded.type,
			parent_id=excluded.parent_id, position=excluded.position, fetched_at=excluded.fetched_at`,
		ch.ID, ch.GuildID, ch.Name, ch.Type, ch.ParentID, ch.Position, time.Now().UnixMilli())
	return err
}

// SavePageWithCursors atomically writes a page of messages and updates sync cursors.
func (d *DB) SavePageWithCursors(msgs []discord.Message, cursors map[string]string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO messages (id, channel_id, guild_id, author_id, author_name, content, type, referenced_id, edited_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content=excluded.content, edited_at=excluded.edited_at, author_name=excluded.author_name`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range msgs {
		createdAt := discord.TimestampFromID(m.ID).UnixMilli()
		var refID *string
		if m.MessageReference != nil {
			refID = &m.MessageReference.MessageID
		}
		if _, err := stmt.Exec(m.ID, m.ChannelID, m.GuildID, m.Author.ID, m.Author.Username,
			m.Content, m.Type, refID, m.EditedTimestamp, createdAt); err != nil {
			return fmt.Errorf("upsert message %s: %w", m.ID, err)
		}
	}

	cursorStmt, err := tx.Prepare(`
		INSERT INTO sync_state (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`)
	if err != nil {
		return err
	}
	defer cursorStmt.Close()

	for k, v := range cursors {
		if _, err := cursorStmt.Exec(k, v); err != nil {
			return fmt.Errorf("set cursor %s: %w", k, err)
		}
	}

	return tx.Commit()
}

func (d *DB) GetSyncState(key string) (string, error) {
	var value string
	err := d.conn.QueryRow("SELECT value FROM sync_state WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", nil // missing key = empty string, not an error
	}
	return value, nil
}

func (d *DB) SetSyncState(key, value string) error {
	_, err := d.conn.Exec(`
		INSERT INTO sync_state (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

type ChannelStats struct {
	ChannelID       string
	ChannelName     string
	MessageCount    int
	OldestMessage   time.Time
	NewestMessage   time.Time
	HistoryComplete bool
}

func (d *DB) GetStats() ([]ChannelStats, error) {
	rows, err := d.conn.Query(`
		SELECT
			c.id, c.name,
			COUNT(m.id) as msg_count,
			COALESCE(MIN(m.created_at), 0) as oldest,
			COALESCE(MAX(m.created_at), 0) as newest,
			COALESCE(ss.value, '') as history_complete
		FROM channels c
		LEFT JOIN messages m ON m.channel_id = c.id
		LEFT JOIN sync_state ss ON ss.key = 'channel:' || c.id || ':history_complete'
		GROUP BY c.id, c.name
		HAVING msg_count > 0
		ORDER BY msg_count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ChannelStats
	for rows.Next() {
		var s ChannelStats
		var oldest, newest int64
		var complete string
		if err := rows.Scan(&s.ChannelID, &s.ChannelName, &s.MessageCount, &oldest, &newest, &complete); err != nil {
			return nil, err
		}
		if oldest > 0 {
			s.OldestMessage = time.UnixMilli(oldest)
		}
		if newest > 0 {
			s.NewestMessage = time.UnixMilli(newest)
		}
		s.HistoryComplete = complete == "true"
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
