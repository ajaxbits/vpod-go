-- name: UpsertFeed :exec
INSERT INTO Feeds (
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
)
ON CONFLICT(id) DO UPDATE SET
    description = excluded.description,
    title = excluded.title,
    updated_at = CURRENT_TIMESTAMP,
    link = excluded.link,
    xml = excluded.xml;

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

-- name: DeleteFeed :exec
DELETE FROM Feeds WHERE id = ?;

-- name: DeleteEpisodesForFeed :exec
DELETE FROM Episodes WHERE feed_id = ?;

-- name: GetFeed :one
SELECT id, created_at, description, title, updated_at, link, xml
FROM Feeds
WHERE id = ?;

-- name: GetAllFeeds :many
WITH Params AS (
    SELECT
      CAST(sqlc.arg(sort_order) AS TEXT) AS sort_order,
      CAST(sqlc.arg(limit) AS INTEGER) AS lim,
      CAST(sqlc.arg(page) AS INTEGER) AS pg
),
FeedData AS (
    SELECT *
    FROM Feeds
    CROSS JOIN Params
    ORDER BY
        CASE WHEN Params.sort_order = 'title_asc' THEN title END ASC,
        CASE WHEN Params.sort_order = 'title_desc' THEN title END DESC,
        CASE WHEN Params.sort_order = 'oldest' THEN created_at END ASC,
        CASE WHEN Params.sort_order = 'newest' THEN created_at END DESC,
        CASE WHEN Params.sort_order NOT IN ('title_asc', 'title_desc', 'oldest', 'newest') THEN created_at END DESC
    LIMIT (SELECT lim FROM Params)
    OFFSET ((SELECT pg FROM Params) - 1) * (SELECT lim FROM Params)
),
TotalCount AS (
    SELECT COUNT(*) AS total_rows
    FROM Feeds
)
SELECT fd.id, fd.created_at, fd.description, fd.title, fd.updated_at, fd.link, fd.xml,
       (SELECT total_rows > ((SELECT pg FROM Params) * (SELECT lim FROM Params)) FROM TotalCount) AS has_more
FROM FeedData fd;
