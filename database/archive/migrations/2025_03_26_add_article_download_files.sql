CREATE TABLE IF NOT EXISTS "archive"."article_file"
(
    "article"    VARCHAR(36) NOT NULL REFERENCES "archive"."article" ("uuid") ON DELETE CASCADE,
    "lang"       VARCHAR(18) NOT NULL CHECK ("lang" <> ''),
    "lang_short" VARCHAR(5)  NOT NULL CHECK ("lang_short" <> ''),
    "file_link"  VARCHAR(2048) DEFAULT NULL CHECK ("file_link" <> '')
);
