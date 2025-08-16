-- posts table: typed columns + full JSONB
CREATE TABLE IF NOT EXISTS posts (
  user_id      INT    NOT NULL,
  id           INT    NOT NULL,
  title        TEXT   NOT NULL,
  body         TEXT   NOT NULL,
  ingested_at  TIMESTAMPTZ NOT NULL,
  source       TEXT   NOT NULL,
  doc          JSONB  NOT NULL,
  PRIMARY KEY (user_id, id)
);

-- hot filter
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);

-- optional JSONB GIN for exploratory queries
CREATE INDEX IF NOT EXISTS idx_posts_doc_gin ON posts USING GIN (doc);
