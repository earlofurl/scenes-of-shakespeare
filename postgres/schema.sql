CREATE TABLE scenes
(
    workid      text NOT NULL,
    act         integer NOT NULL,
    scene       integer NOT NULL,
    description text    NOT NULL,
    body        text    NOT NULL
);

CREATE TABLE works
(
    workid text NOT NULL,
    title  text NOT NULL
);

ALTER TABLE ONLY works
    ADD CONSTRAINT works_pkey PRIMARY KEY (workid);

ALTER TABLE ONLY scenes
    ADD CONSTRAINT scenes_pkey PRIMARY KEY (workid, act, scene);

-- Add a generated column that contains the search document
ALTER TABLE ONLY scenes
    ADD COLUMN fts_doc_en tsvector GENERATED ALWAYS AS (
        to_tsvector('english', body || ' ' || description)
        )
	stored;

-- Set tsvector weights
-- UPDATE scenes
-- SET fts_doc_en=
--             setweight(to_tsvector(description), 'A') ||
--             setweight(to_tsvector(body), 'B');

-- Create a GIN index to make searches faster
CREATE INDEX scenes_fts_doc_en_idx ON scenes USING GIN (fts_doc_en);