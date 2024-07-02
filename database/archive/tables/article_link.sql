CREATE TABLE IF NOT EXISTS "archive"."article_link"
(
  "article_uuid"   VARCHAR(36) UNIQUE NOT NULL REFERENCES "archive"."article" ("uuid") ON DELETE CASCADE,
  "shareable_link" VARCHAR(273)       NOT NULL CHECK ( "shareable_link" <> '' ),
  "expires_at"     TIMESTAMP          NOT NULL DEFAULT current_timestamp + INTERVAL '+7 day'
);
