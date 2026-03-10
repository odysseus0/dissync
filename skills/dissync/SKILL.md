---
name: dissync
description: Discord channel sync and offline search via local SQLite. Use when user mentions Discord, Discord messages, catching up on Discord, searching Discord history, syncing Discord channels, or querying Discord data. Supports incremental sync, full-text search, and direct SQL queries against ~/.dissync/dissync.db.
---

# dissync

Incrementally sync Discord channels to a local SQLite database for offline querying.

## Prerequisites

- `dissync` binary on PATH
- `DISCORD_TOKEN` env var set (user token, not bot token)

## Essential Commands

```bash
# List servers you're in
dissync guilds

# List channels in a server
dissync channels --guild <guild-id>

# Sync specific channels (incremental — only fetches new messages)
dissync sync --channel <id> --channel <id>

# Sync all channels in a server
dissync sync --guild <guild-id>

# Show sync statistics per channel
dissync status
```

## Querying the Database

The database is at `~/.dissync/dissync.db`. Use `sqlite3` directly for maximum flexibility.

```bash
# Recent messages in a channel
sqlite3 ~/.dissync/dissync.db "
  SELECT author_name, content, datetime(created_at/1000, 'unixepoch')
  FROM messages WHERE channel_id = '<id>'
  ORDER BY created_at DESC LIMIT 20;
"

# Full-text search across all synced channels
sqlite3 ~/.dissync/dissync.db "
  SELECT m.author_name, substr(m.content, 1, 200), datetime(m.created_at/1000, 'unixepoch')
  FROM messages_fts
  JOIN messages m ON m.rowid = messages_fts.rowid
  WHERE messages_fts MATCH '<search term>'
  ORDER BY m.created_at DESC LIMIT 20;
"

# Messages from the last 24 hours
sqlite3 ~/.dissync/dissync.db "
  SELECT author_name, content, datetime(created_at/1000, 'unixepoch')
  FROM messages
  WHERE created_at > (strftime('%s', 'now', '-1 day') * 1000)
  ORDER BY created_at;
"

# Message count per channel
sqlite3 ~/.dissync/dissync.db "
  SELECT c.name, COUNT(*) as msgs
  FROM messages m JOIN channels c ON c.id = m.channel_id
  GROUP BY c.name ORDER BY msgs DESC;
"

# Messages by a specific author
sqlite3 ~/.dissync/dissync.db "
  SELECT content, datetime(created_at/1000, 'unixepoch')
  FROM messages WHERE author_name = '<username>'
  ORDER BY created_at DESC LIMIT 20;
"
```

## Workflow: Catching Up on Discord

1. First, sync the channels you care about:
   ```bash
   dissync sync --channel <id1> --channel <id2>
   ```
2. Then query for recent activity:
   ```bash
   sqlite3 ~/.dissync/dissync.db "
     SELECT c.name, m.author_name, substr(m.content, 1, 150), datetime(m.created_at/1000, 'unixepoch')
     FROM messages m JOIN channels c ON c.id = m.channel_id
     WHERE m.created_at > (strftime('%s', 'now', '-1 day') * 1000)
     ORDER BY m.created_at DESC;
   "
   ```
3. Search for specific topics:
   ```bash
   sqlite3 ~/.dissync/dissync.db "
     SELECT m.author_name, substr(m.content, 1, 200), datetime(m.created_at/1000, 'unixepoch')
     FROM messages_fts JOIN messages m ON m.rowid = messages_fts.rowid
     WHERE messages_fts MATCH '<topic>'
     ORDER BY m.created_at DESC LIMIT 20;
   "
   ```

## Schema Reference

| Table | Purpose |
|-------|---------|
| `guilds` | Server metadata (id, name) |
| `channels` | Channel metadata (id, guild_id, name, type) |
| `messages` | All synced messages (id, channel_id, author_id, author_name, content, created_at) |
| `messages_fts` | FTS5 virtual table for full-text search (porter stemming) |
| `sync_state` | Incremental sync cursors per channel |

## Notes

- Sync is incremental and interruption-safe — re-running only fetches new messages
- FTS uses porter stemming: searching "running" matches "run", "runs", etc.
- `created_at` is Unix milliseconds (divide by 1000 for `datetime()`)
- The `--channel` flag is repeatable for syncing multiple channels at once
- Voice channels return minimal text data (join/leave events)
