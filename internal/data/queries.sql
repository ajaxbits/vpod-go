-- name: UpsertFeed :exec
INSERT OR REPLACE INTO Feeds (
    id,
    created_at,
    description,
    title,
    updated_at,
    link,
    xml
) VALUES (
    ?,
    ?,
    ?,
    ?,
    CURRENT_TIMESTAMP,
    ?,
    ?
);

-- name: UpsertEpisode :exec
INSERT OR REPLACE INTO Episodes (
    id,
    audio_url,
    audio_length_bytes,
    description,
    duration,
    feed_id,
    released_at,
    thumbnail,
    title,
    video_url
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetEpisodesForFeed :many
SELECT id,
  audio_url,
  audio_length_bytes,
  description,
  duration,
  feed_id,
  released_at,
  thumbnail,
  title,
  video_url
FROM Episodes
WHERE feed_id = ?;

-- name: GetFeedXML :one
SELECT xml FROM Feeds WHERE id = ?;

-- name: GetOlderEpisodesForFeed :many
SELECT *
FROM Episodes as e
WHERE e.feed_id = ?1
AND released_at < (
    SELECT released_at
    FROM Episodes AS e
    WHERE e.id = ?2
      AND e.feed_id = ?1
)
ORDER BY released_at DESC;

-- name: GetAllFeedIds :many
SELECT id
FROM Feeds;

-- name: GetAllFeeds :many
SELECT *
FROM Feeds
LIMIT ?;
