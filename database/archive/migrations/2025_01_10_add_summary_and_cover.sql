BEGIN;

ALTER TABLE "archive"."article"
    ADD COLUMN "summary"       VARCHAR(512)  NOT NULL DEFAULT 'no summary' CHECK ("summary" <> ''),
    ADD COLUMN "cover_url"     VARCHAR(2048) NOT NULL DEFAULT 'about:blank' CHECK ("cover_url" <> ''),
    ADD COLUMN "cover_caption" VARCHAR(256) CHECK ("cover_caption" <> '');

COMMIT;
