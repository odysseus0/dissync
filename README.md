# dissync

Incrementally sync Discord channels to a local SQLite database using a user token. Built for maintainers who want to catch up on Discord activity offline via SQL queries.

## Install

### Nix flake

```nix
# flake.nix input
dissync.url = "github:odysseus0/dissync";

# home-manager package
inputs.dissync.packages.${pkgs.system}.default
```

### Go

```bash
go install github.com/tengjizhang/dissync@latest
```

## Usage

```bash
export DISCORD_TOKEN="your-user-token"

# List servers
dissync guilds

# List channels in a server
dissync channels --guild <guild-id>

# Sync specific channels
dissync sync --channel <id> --channel <id>

# Sync all channels in a server
dissync sync --guild <guild-id>

# Show sync stats
dissync status
```

## Querying

The database is at `~/.dissync/dissync.db`. Query it directly:

```sql
-- Recent messages in a channel
SELECT author_name, content, datetime(created_at/1000, 'unixepoch')
FROM messages WHERE channel_id = '...'
ORDER BY created_at DESC LIMIT 20;

-- Full-text search
SELECT m.author_name, m.content, datetime(m.created_at/1000, 'unixepoch')
FROM messages_fts
JOIN messages m ON m.rowid = messages_fts.rowid
WHERE messages_fts MATCH 'search term'
LIMIT 10;

-- Messages from the last 24 hours
SELECT author_name, content, datetime(created_at/1000, 'unixepoch')
FROM messages
WHERE created_at > (strftime('%s', 'now', '-1 day') * 1000)
ORDER BY created_at;
```

## How it works

- Incremental sync with per-page checkpoints — safe to interrupt and resume
- Forward sync catches new messages; backfill walks history backward
- Cursors stored in `sync_state` table — re-running only fetches what's new
- FTS5 index with porter stemming for full-text search
- Uses Discord user token (not bot token) — no server admin access needed

## Getting a user token

1. Open Discord in a browser
2. Open DevTools → Console
3. Run:
```js
(()=>{let m=[];webpackChunkdiscord_app.push([[''],{},e=>{for(let c in e.c)m.push(e.c[c])}]);return m.find(m=>m?.exports?.default?.getToken)?.exports?.default?.getToken()})()
```

**Note:** Using user tokens for automation is against Discord's ToS. Use at your own discretion.
