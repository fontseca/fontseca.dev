CREATE TABLE IF NOT EXISTS "archive"."article_tag"
(
  "article_uuid" VARCHAR(36) NOT NULL REFERENCES "archive"."article" ("uuid") ON DELETE CASCADE,
  "tag_id"       VARCHAR(32) NOT NULL REFERENCES "archive"."tag" ("id") ON DELETE CASCADE
);
