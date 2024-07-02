CREATE TABLE IF NOT EXISTS "projects"."project_tag"
(
  "project_uuid"        VARCHAR(36) NOT NULL REFERENCES "projects"."project" ("uuid") ON DELETE CASCADE,
  "technology_tag_uuid" VARCHAR(36) NOT NULL REFERENCES "projects"."tag" ("uuid") ON DELETE CASCADE
);
