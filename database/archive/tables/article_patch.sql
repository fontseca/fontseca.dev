CREATE TABLE IF NOT EXISTS "archive"."article_patch"
(
  "article_uuid" VARCHAR(36) UNIQUE NOT NULL REFERENCES "archive"."article" ("uuid") ON DELETE CASCADE,
  "title"        VARCHAR(256) CHECK ( "title" <> '' ),
  "topic"        VARCHAR(32) REFERENCES "archive"."topic" ("id") CHECK ( "topic" <> '' ),
  "slug"         VARCHAR(512) CHECK ("slug" <> ''),
  "read_time"    SMALLINT DEFAULT 0 CHECK ( "read_time" >= 0 ),
  "content"      VARCHAR(3145728) CHECK ( "content" <> '' )
);
