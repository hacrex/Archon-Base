-- Archon Base core resources, migration 001.
-- PostgreSQL 15+; application-generated UUIDs are supported, while pgcrypto
-- supplies a safe default for callers that do not generate IDs themselves.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE projects (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     uuid NOT NULL,
    name                text NOT NULL,
    display_name        text NOT NULL,
    description         text NOT NULL DEFAULT '',
    environment         text NOT NULL DEFAULT 'dev',
    region              text NOT NULL DEFAULT 'local',
    spec                jsonb NOT NULL DEFAULT '{}'::jsonb,
    labels              jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations         jsonb NOT NULL DEFAULT '{}'::jsonb,
    phase               text NOT NULL DEFAULT 'Pending',
    message             text NOT NULL DEFAULT '',
    generation          bigint NOT NULL DEFAULT 1,
    observed_generation bigint NOT NULL DEFAULT 0,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,
    CONSTRAINT projects_name_format CHECK (name ~ '^[a-z0-9]([-a-z0-9]*[a-z0-9])?$'),
    CONSTRAINT projects_environment_check CHECK (environment IN ('dev', 'staging', 'prod')),
    CONSTRAINT projects_phase_check CHECK (phase IN ('Pending', 'Ready', 'Degraded', 'Failed', 'Deleting')),
    CONSTRAINT projects_generation_check CHECK (generation > 0),
    CONSTRAINT projects_observed_generation_check CHECK (observed_generation >= 0),
    CONSTRAINT projects_spec_object_check CHECK (jsonb_typeof(spec) = 'object'),
    CONSTRAINT projects_labels_object_check CHECK (jsonb_typeof(labels) = 'object'),
    CONSTRAINT projects_annotations_object_check CHECK (jsonb_typeof(annotations) = 'object')
);

CREATE UNIQUE INDEX projects_active_name_idx
    ON projects (organization_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX projects_org_idx ON projects (organization_id);

CREATE TABLE database_instances (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          uuid NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    name                text NOT NULL,
    engine              text NOT NULL,
    version             text NOT NULL DEFAULT '',
    plan                text NOT NULL,
    replicas            integer NOT NULL DEFAULT 1,
    shards              integer NOT NULL DEFAULT 1,
    storage             jsonb NOT NULL DEFAULT '{}'::jsonb,
    backup              jsonb NOT NULL DEFAULT '{}'::jsonb,
    network             jsonb NOT NULL DEFAULT '{"exposure":"private"}'::jsonb,
    spec                jsonb NOT NULL DEFAULT '{}'::jsonb,
    labels              jsonb NOT NULL DEFAULT '{}'::jsonb,
    annotations         jsonb NOT NULL DEFAULT '{}'::jsonb,
    phase               text NOT NULL DEFAULT 'Pending',
    message             text NOT NULL DEFAULT '',
    generation          bigint NOT NULL DEFAULT 1,
    observed_generation bigint NOT NULL DEFAULT 0,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,
    CONSTRAINT database_instances_name_format CHECK (name ~ '^[a-z0-9]([-a-z0-9]*[a-z0-9])?$'),
    CONSTRAINT database_instances_engine_check CHECK (engine IN ('qdrant', 'chroma', 'weaviate', 'milvus', 'pgvector', 'scylladb')),
    CONSTRAINT database_instances_replicas_check CHECK (replicas > 0),
    CONSTRAINT database_instances_shards_check CHECK (shards > 0),
    CONSTRAINT database_instances_phase_check CHECK (phase IN ('Pending', 'Ready', 'Degraded', 'Failed', 'Deleting')),
    CONSTRAINT database_instances_storage_object_check CHECK (jsonb_typeof(storage) = 'object'),
    CONSTRAINT database_instances_backup_object_check CHECK (jsonb_typeof(backup) = 'object'),
    CONSTRAINT database_instances_network_object_check CHECK (jsonb_typeof(network) = 'object'),
    CONSTRAINT database_instances_spec_object_check CHECK (jsonb_typeof(spec) = 'object'),
    CONSTRAINT database_instances_labels_object_check CHECK (jsonb_typeof(labels) = 'object'),
    CONSTRAINT database_instances_annotations_object_check CHECK (jsonb_typeof(annotations) = 'object')
);

CREATE UNIQUE INDEX database_instances_active_name_idx
    ON database_instances (project_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX database_instances_project_idx ON database_instances (project_id);
CREATE INDEX database_instances_phase_idx ON database_instances (phase);

-- The outbox is written in the same transaction as a resource mutation.
-- A publisher can safely claim rows with FOR UPDATE SKIP LOCKED and publish
-- them to NATS JetStream, then set published_at.
CREATE TABLE outbox_events (
    id              bigserial PRIMARY KEY,
    aggregate_type  text NOT NULL,
    aggregate_id    uuid NOT NULL,
    event_type      text NOT NULL,
    payload         jsonb NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    published_at    timestamptz,
    attempts        integer NOT NULL DEFAULT 0,
    last_error      text,
    CONSTRAINT outbox_payload_object_check CHECK (jsonb_typeof(payload) = 'object'),
    CONSTRAINT outbox_attempts_check CHECK (attempts >= 0)
);

CREATE INDEX outbox_unpublished_idx
    ON outbox_events (id)
    WHERE published_at IS NULL;

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER projects_set_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER database_instances_set_updated_at
    BEFORE UPDATE ON database_instances
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
