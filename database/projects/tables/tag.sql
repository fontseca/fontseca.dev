CREATE TABLE IF NOT EXISTS "projects"."tag"
(
  "uuid"       VARCHAR(36)        NOT NULL PRIMARY KEY DEFAULT "extensions"."uuid_generate_v4"(),
  "name"       VARCHAR(64) UNIQUE NOT NULL CHECK ("name" <> ''),
  "created_at" TIMESTAMP          NOT NULL             DEFAULT current_timestamp,
  "updated_at" TIMESTAMP          NOT NULL             DEFAULT current_timestamp
);
