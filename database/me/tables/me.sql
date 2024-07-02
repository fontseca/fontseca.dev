CREATE TABLE IF NOT EXISTS "me"."me"
(
  "username"      VARCHAR(64) PRIMARY KEY      DEFAULT 'fontseca.dev' CHECK ("username" = 'fontseca.dev'),
  "first_name"    VARCHAR(6)          NOT NULL DEFAULT 'Jeremy' CHECK ( "last_name" <> '' ),
  "last_name"     VARCHAR(7)          NOT NULL DEFAULT 'Fonseca' CHECK ( "last_name" <> '' ),
  "summary"       VARCHAR(1024)       NOT NULL CHECK ( "summary" <> '' ),
  "job_title"     VARCHAR(64)         NOT NULL CHECK ( "job_title" <> '' ),
  "email"         VARCHAR(254) UNIQUE NOT NULL CHECK ( "email" <> '' ),
  "photo_url"     VARCHAR(2048)       NOT NULL DEFAULT 'about:blank' CHECK ( "photo_url" <> '' ),
  "resume_url"    VARCHAR(2048)       NOT NULL DEFAULT 'about:blank' CHECK ( "resume_url" <> '' ),
  "coding_since"  SMALLINT            NOT NULL DEFAULT 2017 CHECK ( "coding_since" = 2017 ),
  "company"       VARCHAR(64) CHECK ( "company" <> '' ),
  "location"      VARCHAR(64) CHECK ("location" <> ''),
  "hireable"      BOOLEAN             NOT NULL DEFAULT FALSE,
  "github_url"    VARCHAR(2048)       NOT NULL DEFAULT 'https://github.com/fontseca' CHECK ( "github_url" <> '' ),
  "linkedin_url"  VARCHAR(2048)       NOT NULL DEFAULT 'about:blank' CHECK ( "linkedin_url" <> '' ),
  "youtube_url"   VARCHAR(2048)       NOT NULL DEFAULT 'about:blank' CHECK ("youtube_url" <> ''),
  "twitter_url"   VARCHAR(2048)       NOT NULL DEFAULT 'about:blank' CHECK ("twitter_url" <> ''),
  "instagram_url" VARCHAR(2048)       NOT NULL DEFAULT 'about:blank' CHECK ("instagram_url" <> ''),
  "created_at"    TIMESTAMP           NOT NULL DEFAULT current_timestamp,
  "updated_at"    TIMESTAMP           NOT NULL DEFAULT current_timestamp
);
