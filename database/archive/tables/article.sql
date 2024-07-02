CREATE TABLE IF NOT EXISTS "archive"."article"
(
  "uuid"         VARCHAR(36) PRIMARY KEY      DEFAULT "extensions"."uuid_generate_v4"(),
  "topic"        VARCHAR(32) REFERENCES "archive"."topic" ("id"),
  "author"       VARCHAR(64)         NOT NULL REFERENCES "me"."me" ("username"),
  "title"        VARCHAR(256) UNIQUE NOT NULL CHECK ("title" <> ''),
  "slug"         VARCHAR(512) UNIQUE NOT NULL CHECK ("slug" <> ''),
  "read_time"    SMALLINT            NOT NULL DEFAULT 0 CHECK ("read_time" >= 0),
  "views"        INTEGER             NOT NULL DEFAULT 0 CHECK ("views" >= 0),
  "content"      VARCHAR(3145728)    NOT NULL DEFAULT 'no content' CHECK ("content" <> ''),
  "draft"        BOOLEAN             NOT NULL DEFAULT TRUE,
  "pinned"       BOOLEAN             NOT NULL DEFAULT FALSE,
  "hidden"       BOOLEAN             NOT NULL DEFAULT FALSE,
  "drafted_at"   TIMESTAMP           NOT NULL DEFAULT current_timestamp,
  "published_at" TIMESTAMP                    DEFAULT NULL,
  "updated_at"   TIMESTAMP           NOT NULL DEFAULT current_timestamp,
  "modified_at"  TIMESTAMP                    DEFAULT NULL
);
