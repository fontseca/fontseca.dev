CREATE TABLE IF NOT EXISTS "archive"."topic"
(
  "id"         VARCHAR(32) UNIQUE PRIMARY KEY CHECK ( "id" <> '' ),
  "name"       VARCHAR(32) UNIQUE NOT NULL CHECK ( "name" <> '' ),
  "created_at" TIMESTAMP          NOT NULL DEFAULT current_timestamp,
  "updated_at" TIMESTAMP          NOT NULL DEFAULT current_timestamp
);
