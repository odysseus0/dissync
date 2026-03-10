package db

const schemaSQL = `
CREATE TABLE IF NOT EXISTS guilds (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    icon       TEXT,
    fetched_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS channels (
    id         TEXT PRIMARY KEY,
    guild_id   TEXT NOT NULL,
    name       TEXT NOT NULL,
    type       INTEGER NOT NULL,
    parent_id  TEXT,
    position   INTEGER,
    fetched_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id            TEXT PRIMARY KEY,
    channel_id    TEXT NOT NULL,
    guild_id      TEXT,
    author_id     TEXT NOT NULL,
    author_name   TEXT NOT NULL,
    content       TEXT NOT NULL,
    type          INTEGER NOT NULL DEFAULT 0,
    referenced_id TEXT,
    edited_at     TEXT,
    created_at    INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_channel_created
    ON messages(channel_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_channel_id
    ON messages(channel_id, id ASC);

CREATE TABLE IF NOT EXISTS sync_state (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts
    USING fts5(
        author_name,
        content,
        content='messages',
        content_rowid='rowid',
        tokenize='porter unicode61'
    );

CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(rowid, author_name, content)
    VALUES (new.rowid, new.author_name, new.content);
END;

CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, author_name, content)
    VALUES ('delete', old.rowid, old.author_name, old.content);
END;

CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
    INSERT INTO messages_fts(messages_fts, rowid, author_name, content)
    VALUES ('delete', old.rowid, old.author_name, old.content);
    INSERT INTO messages_fts(rowid, author_name, content)
    VALUES (new.rowid, new.author_name, new.content);
END;
`
