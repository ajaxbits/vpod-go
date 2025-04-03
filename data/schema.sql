PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;

CREATE TABLE IF NOT EXISTS Feeds (
    id BLOB PRIMARY KEY NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    title TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    link TEXT NOT NULL,
    xml TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS Episodes (
    id BLOB PRIMARY KEY NOT NULL UNIQUE,
    audio_url TEXT NOT NULL,
    audio_length_bytes INTEGER NOT NULL,
    description TEXT,
    duration INTEGER,
    feed_id TEXT NOT NULL,
    released_at TIMESTAMP,
    thumbnail TEXT,
    title TEXT NOT NULL,
    video_url TEXT,
    FOREIGN KEY (feed_id) REFERENCES Feeds(id)
);
