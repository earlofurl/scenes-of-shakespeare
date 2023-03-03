-- name: Search :many
SELECT workid,
       act,
       scene,
       description,
       body,
       ts_headline(body, websearch_to_tsquery($1))::text AS headline
      FROM scenes
      WHERE fts_doc_en @@ websearch_to_tsquery($1);

-- name: GetScene :one
SELECT description, body
FROM scenes
WHERE workid = $1
  AND act = $2
  AND scene = $3;

-- name: GetWork :one
SELECT title
FROM works
WHERE workid = $1;