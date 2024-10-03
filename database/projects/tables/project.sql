CREATE TABLE IF NOT EXISTS "projects"."project"
(
  "uuid"             VARCHAR(36) PRIMARY KEY       DEFAULT "extensions"."uuid_generate_v4"(),
  "name"             VARCHAR(64) UNIQUE   NOT NULL CHECK ( "name" <> '' ),
  "slug"             VARCHAR(2024) UNIQUE NOT NULL CHECK ("slug" <> ''),
  "homepage"         VARCHAR(2048)        NOT NULL DEFAULT 'about:blank' CHECK ("homepage" <> ''),
  "language"         VARCHAR(64) CHECK ("language" <> ''),
  "summary"          VARCHAR(325)         NOT NULL DEFAULT 'no summary' CHECK ("summary" <> ''),
  "read_time"        SMALLINT             NOT NULL DEFAULT 0 CHECK ("read_time" >= 0),
  "content"          VARCHAR(2145728)     NOT NULL DEFAULT 'no content' CHECK ("content" <> ''),
  "first_image_url"  VARCHAR(2048)        NOT NULL DEFAULT 'about:blank' CHECK ("first_image_url" <> ''),
  "second_image_url" VARCHAR(2048)        NOT NULL DEFAULT 'about:blank' CHECK ("second_image_url" <> ''),
  "github_url"       VARCHAR(2048)        NOT NULL DEFAULT 'about:blank' CHECK ("github_url" <> ''),
  "collection_url"   VARCHAR(2048)        NOT NULL DEFAULT 'about:blank' CHECK ("collection_url" <> ''),
  "playground_url"   VARCHAR(2048)        NOT NULL DEFAULT 'about:blank' CHECK ("playground_url" <> ''),
  "playable"         BOOLEAN              NOT NULL DEFAULT FALSE,
  "archived"         BOOLEAN              NOT NULL DEFAULT FALSE,
  "finished"         BOOLEAN              NOT NULL DEFAULT FALSE,
  "created_at"       TIMESTAMP            NOT NULL DEFAULT current_timestamp,
  "updated_at"       TIMESTAMP            NOT NULL DEFAULT current_timestamp
);