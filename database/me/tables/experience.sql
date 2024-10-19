CREATE TABLE IF NOT EXISTS "me"."experience"
(
  "uuid"             VARCHAR(36) PRIMARY KEY   DEFAULT "extensions"."uuid_generate_v4"(),
  "date_start"       DATE             NOT NULL CHECK ( "date_start" > '2017/01/01' ),
  "date_end"         DATE                      DEFAULT NULL CHECK ("date_end" > '2017/01/01'),
  "job_title"        VARCHAR(64)      NOT NULL CHECK ("job_title" <> ''),
  "company"          VARCHAR(64)      NOT NULL CHECK ("company" <> ''),
  "company_homepage" VARCHAR(2048)             DEFAULT NULL CHECK ("company_homepage" <> ''),
  "country"          VARCHAR(64) CHECK ("country" <> ''),
  "summary"          VARCHAR(1145728) NOT NULL CHECK ( "summary" <> '' ),
  "active"           BOOLEAN          NOT NULL DEFAULT FALSE,
  "hidden"           BOOLEAN          NOT NULL DEFAULT FALSE,
  "created_at"       TIMESTAMP        NOT NULL DEFAULT current_timestamp,
  "updated_at"       TIMESTAMP        NOT NULL DEFAULT current_timestamp
);
