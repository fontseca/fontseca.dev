CREATE TABLE IF NOT EXISTS "me"."experience"
(
  "uuid"       VARCHAR(36) PRIMARY KEY DEFAULT "extensions"."uuid_generate_v4"(),
  "starts"     SMALLINT      NOT NULL CHECK ( "starts" > 2017 ),
  "ends"       SMALLINT                DEFAULT NULL CHECK ("ends" > 2017 OR "ends" IS NULL),
  "job_title"  VARCHAR(64)   NOT NULL CHECK ("job_title" <> ''),
  "company"    VARCHAR(64)   NOT NULL CHECK ("company" <> ''),
  "country"    VARCHAR(64) CHECK ("country" <> ''),
  "summary"    VARCHAR(1145728) NOT NULL CHECK ( "summary" <> '' ),
  "active"     BOOLEAN       NOT NULL  DEFAULT FALSE,
  "hidden"     BOOLEAN       NOT NULL  DEFAULT FALSE,
  "created_at" TIMESTAMP     NOT NULL  DEFAULT current_timestamp,
  "updated_at" TIMESTAMP     NOT NULL  DEFAULT current_timestamp
);
